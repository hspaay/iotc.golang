// Package publisher ...
// - Publishes updates to node, inputs and outputs when they are (re)discovered
// - configuration of nodes
// - control of inputs
// - update of security keys and identity signature
// Thread-safe. All public functions can be invoked from multiple goroutines
package publisher

import (
	"crypto/ecdsa"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/iotdomain/iotdomain-go/inputs"
	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/iotdomain/iotdomain-go/nodes"
	"github.com/iotdomain/iotdomain-go/outputs"
	"github.com/iotdomain/iotdomain-go/persist"
	"github.com/iotdomain/iotdomain-go/publishers"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

// reserved keywords
const (
	// DefaultPollInterval in which the registered nodes, inputs and outputs are queried for
	// polling based sources
	DefaultPollInterval = 600
)

// Publisher carries the operating state of 'this' publisher
type Publisher struct {
	domainInputs       *inputs.DomainInputs        // discovered inputs from the domain
	domainNodes        *nodes.DomainNodes          // discovered nodes from the domain
	domainOutputs      *outputs.DomainOutputs      // discovered outputs from the domain
	domainOutputValues *outputs.DomainOutputValues // output values from the domain

	receiveSetInput          *inputs.ReceiveSetInput           // listener for set input for registered inputs
	receiveNodeConfigure     *nodes.ReceiveNodeConfigure       // listener for node configure for registered nodes
	registeredForecastValues *outputs.RegisteredForecastValues // output forecasts values published by this publisher
	registeredInputs         *inputs.RegisteredInputs          // registered/published inputs from this publisher
	registeredNodes          *nodes.RegisteredNodes            // registered/published nodes from this publisher
	receiveNodeSetAlias      *nodes.ReceiveNodeAlias           // listener for set node alias
	registeredOutputs        *outputs.RegisteredOutputs        // registered/published outputs from this publisher
	registeredOutputValues   *outputs.RegisteredOutputValues   // registered/published output values from this publisher

	configFolder     string                    // folder to save publisher, node, inputs and outputs configuration
	domainPublishers *publishers.PublisherList // publishers on the network by discovery address

	fullIdentity        *types.PublisherFullIdentity // this publishers identity
	identityPrivateKey  *ecdsa.PrivateKey            // key for signing and encryption
	isRunning           bool                         // publisher was started and is running
	messenger           messaging.IMessenger         // Message bus messenger to use
	messageSigner       *messaging.MessageSigner     // publishing signed messages
	onNodeConfigHandler nodes.NodeConfigureHandler   // handle before applying configuration
	onNodeInputHandler  inputs.SetInputHandler       // handle to update device/service input
	pollHandler         func(publisher *Publisher)   // function that performs value polling
	pollCountdown       int                          // countdown each heartbeat
	pollInterval        int                          // value polling interval in seconds

	// background publications require a mutex to prevent concurrent access
	exitChannel  chan bool
	signMessages bool        // signing on or off
	updateMutex  *sync.Mutex // mutex for async updating and publishing
}

// Address returns the publisher's identity address
func (publisher *Publisher) Address() string {
	// identityAddr := nodes.MakePublisherIdentityAddress(publisher.Domain(), publisher.PublisherID())
	// return identityAddr
	return publisher.fullIdentity.Address
}

// PublisherID returns the publisher's ID
func (publisher *Publisher) PublisherID() string {
	return publisher.fullIdentity.PublisherID
}

// Domain returns the publication domain
func (publisher *Publisher) Domain() string {
	return publisher.fullIdentity.Domain
}

// Identity return a copy of this publisher's full identity
func (publisher *Publisher) FullIdentity() types.PublisherFullIdentity {
	return *publisher.fullIdentity
}

// SetLogging sets the logging level and output file for this publisher
// Intended for setting logging from configuration
// levelName is the requested logging level: error, warning, info, debug
// filename is the output log file full name including path
func (publisher *Publisher) SetLogging(levelName string, filename string) {
	loggingLevel := log.DebugLevel

	if levelName != "" {
		switch strings.ToLower(levelName) {
		case "error":
			loggingLevel = log.ErrorLevel
		case "warn":
		case "warning":
			loggingLevel = log.WarnLevel
		case "info":
			loggingLevel = log.InfoLevel
		case "debug":
			loggingLevel = log.DebugLevel
		}
	}
	logOut := os.Stderr
	if filename != "" {
		logFileHandle, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			logrus.Errorf("Publisher.SetLogging: Unable to open logfile: %s", err)
		} else {
			logrus.Warnf("Publisher.SetLogging: Send logging output to %s", filename)
			logOut = logFileHandle
		}
	}

	logrus.SetFormatter(
		&log.TextFormatter{
			// LogFormat: "",
			// DisableColors:   true,
			// DisableLevelTruncation: true,
			// PadLevelText:    true,
			TimestampFormat: "2006-01-02 15:04:05.000",
			FullTimestamp:   true,
			// ForceFormatting: true,
		})
	logrus.SetOutput(logOut)
	logrus.SetLevel(loggingLevel)

	logrus.SetReportCaller(false) // publisher logging includes caller and file:line#
}

// SetNodeConfigHandler set the handler for updating node configuration
func (publisher *Publisher) SetNodeConfigHandler(
	handler func(address string, config types.NodeAttrMap) types.NodeAttrMap) {

	publisher.receiveNodeConfigure.SetConfigureNodeHandler(handler)
}

// SetNodeInputHandler set the handler for updating node inputs
func (publisher *Publisher) SetNodeInputHandler(
	handler func(address string, message *types.SetInputMessage)) {

	publisher.receiveSetInput.SetNodeInputHandler(handler)
}

// LoadConfig loads previously saved configuration
// If a node file exists in the given folder the nodes will be added/updated. Existing nodes will be replaced.
// If autosave is set then save this publisher's nodes and configs when updated.
// folder with the configuration files. Use "" for default, which is persist.DefaultConfigFolder
//   autosave indicates to save updates to node configuration
// returns error if folder doesn't exist
func (publisher *Publisher) LoadConfig(folder string, autosave bool) error {
	var err error = nil
	if folder == "" {
		folder = persist.DefaultConfigFolder
	}
	if autosave {
		publisher.configFolder = folder
	}
	if folder != "" {
		nodeList := make([]*types.NodeDiscoveryMessage, 0)
		err = persist.LoadNodes(folder, publisher.PublisherID(), &nodeList)
		if err == nil {
			publisher.registeredNodes.UpdateNodes(nodeList)
		}
	}
	return err
}

// SetPollInterval is a convenience function for periodic polling of updates to registered
// nodes, inputs, outputs and output values.
// seconds interval to perform another poll. Default (0) is DefaultPollInterval
// intended for publishers that need to poll for values
func (publisher *Publisher) SetPollInterval(seconds int, handler func(publisher *Publisher)) {
	logrus.Infof("Publisher.SetPoll: interval = %d seconds", seconds)
	if seconds > 0 {
		publisher.pollInterval = seconds
	} else {
		publisher.pollInterval = DefaultPollInterval
	}
	publisher.pollHandler = handler
}

// SetPublisherID changes the publisher's ID. Use before calling Start
// func (publisher *Publisher) SetPublisherID(id string) {
// 	publisher.id = id
// }

// SetSigningOnOff turns signing of publications on or off.
// the default is provided when creating the publisher.
func (publisher *Publisher) SetSigningOnOff(on bool) {
	publisher.signMessages = on
}

// Start publishing and listen for configuration and input messages
// This will create the publisher node and load previously saved nodes
// Start will fail if no messenger has been provided.
// persistNodes will load previously saved nodes at startup and save them on configuration change
func (publisher *Publisher) Start() {

	if publisher.messenger == nil {
		logrus.Errorf("Publisher.Start: Can't start publisher %s/%s without a messenger. See SetMessenger()",
			publisher.Domain(), publisher.PublisherID())
		return
	}
	if !publisher.isRunning {
		logrus.Warningf("Publisher.Start: Starting publisher %s/%s", publisher.Domain(), publisher.PublisherID())
		publisher.updateMutex.Lock()
		publisher.isRunning = true
		publisher.updateMutex.Unlock()

		go publisher.heartbeatLoop()
		// wait for the heartbeat to start
		<-publisher.exitChannel

		// TODO: support LWT
		publisher.messenger.Connect("", "")

		publisher.domainNodes.Start()
		publisher.domainInputs.Start()
		publisher.domainOutputs.Start()
		publisher.receiveSetInput.Start()
		publisher.receiveNodeSetAlias.Start()
		publisher.receiveNodeConfigure.Start()

		// subscribe to publisher nodes to verify signature for input commands
		pubAddr := publishers.MakePublisherIdentityAddress(publisher.Domain(), "+")
		publisher.messenger.Subscribe(pubAddr, publisher.handlePublisherDiscovery)

		// publish discovery of this publisher
		publisher.PublishIdentity(*publisher.messageSigner)

		logrus.Infof("Publisher.Start: Publisher %s started", publisher.PublisherID())
	}
}

// Stop publishing
// Wait until the heartbeat loop has finished processing messages
func (publisher *Publisher) Stop() {
	logrus.Warningf("Publisher.Stop: Stopping publisher %s", publisher.PublisherID())
	publisher.updateMutex.Lock()
	if publisher.isRunning {
		publisher.isRunning = false

		publisher.receiveNodeConfigure.Stop()
		publisher.receiveSetInput.Stop()
		publisher.domainOutputs.Stop()
		publisher.domainInputs.Stop()
		publisher.domainNodes.Stop()

		publisher.updateMutex.Unlock()
		// wait for heartbeat to end
		<-publisher.exitChannel
	} else {
		publisher.updateMutex.Unlock()
	}
	logrus.Info("... bye bye")
}

// WaitForSignal waits until a TERM or INT signal is received
func (publisher *Publisher) WaitForSignal() {

	// catch all signals since not explicitly listing
	exitChannel := make(chan os.Signal, 1)

	//signal.Notify(exitChannel, syscall.SIGTERM|syscall.SIGHUP|syscall.SIGINT)
	signal.Notify(exitChannel, syscall.SIGINT, syscall.SIGTERM)

	sig := <-exitChannel
	log.Warningf("RECEIVED SIGNAL: %s", sig)
	fmt.Println()
	fmt.Println(sig)
}

// Main heartbeat loop to publish, discove and poll value updates
func (publisher *Publisher) heartbeatLoop() {
	logrus.Infof("Publisher.heartbeatLoop: starting heartbeat loop")
	publisher.exitChannel <- false

	for {
		time.Sleep(time.Second)

		// FIXME: The duration of publishing these updates adds to the heartbeat which delays the heartbeat
		publisher.PublishUpdates()

		// poll for discovery and values of registered nodes, inputs and outputs
		if (publisher.pollCountdown <= 0) && (publisher.pollHandler != nil) {
			publisher.pollHandler(publisher)
			publisher.pollCountdown = publisher.pollInterval
		}
		publisher.pollCountdown--

		publisher.updateMutex.Lock()
		isRunning := publisher.isRunning
		publisher.updateMutex.Unlock()
		if !isRunning {
			break
		}
	}
	publisher.exitChannel <- true
	logrus.Infof("Publisher.heartbeatLoop: Ending loop of publisher %s", publisher.PublisherID())
}

// NewPublisher creates a publisher instance. This is used for all publications.
//
// The identityFolder contains the publisher identity file <publisherID>-identity.json. "" for default config folder.
// The identity is written here when it is first created or is renewed by the domain security service.
// This file only needs to be accessible during publisher startup.
//
// The cacheFolder contains stored discovered nodes, inputs, outputs, and external publishers. It also
// contains node configuration so deleting these files will remove custom node configuration, for example
// configuration of alias and name.
//
// domain and publisherID identity this publisher. If the identity file does not match these, it
// is discarded and a new identity is created. If the publisher has joined the domain and the DSS has issued
// the identity then changing domain or publisherID invalidates the publisher and it has to rejoin
// the domain. If no domain is provided, the default 'local' is used.
//
// signingMethod indicates if and how publications must be signed. The default is jws. For testing 'none' can be used.
//
// messenger for publishing onto the message bus
func NewPublisher(
	identityFolder string,
	cacheFolder string,
	domain string,
	publisherID string,
	signMessages bool,
	messenger messaging.IMessenger,
) *Publisher {

	logrus.SetLevel(log.DebugLevel)
	if domain == "" {
		domain = types.LocalDomainID
	}

	// These are the basis for signing and identifying publishers
	identity, privateKey := SetupPublisherIdentity(identityFolder, domain, publisherID)
	domainPublishers := publishers.NewPublisherList()
	messageSigner := messaging.NewMessageSigner(signMessages, domainPublishers.GetPublisherKey, messenger, privateKey)
	domainPublishers.UpdatePublisher(&identity.PublisherIdentityMessage)

	// application services
	domainInputs := inputs.NewDomainInputs(domainPublishers.GetPublisherKey, messageSigner)
	domainOutputs := outputs.NewDomainOutputs(domainPublishers.GetPublisherKey, messageSigner)
	domainNodes := nodes.NewDomainNodes(domainPublishers.GetPublisherKey, messageSigner)
	registeredNodes := nodes.NewRegisteredNodes(domain, publisherID)
	registeredInputs := inputs.NewRegisteredInputs(domain, publisherID)
	registeredOutputs := outputs.NewRegisteredOutputs(domain, publisherID)
	registeredOutputValues := outputs.NewRegisteredOutputValues(domain, publisherID)
	registeredForecastValues := outputs.NewRegisteredForecastValues(domain, publisherID)

	// the handlers are provided with SetXxx functions
	receiveNodeConfigure := nodes.NewReceiveNodeConfigure(domain, publisherID, nil, messageSigner,
		registeredNodes, privateKey, domainPublishers.GetPublisherKey)
	receiveSetInput := inputs.NewReceiveSetInput(domain, publisherID, nil, messageSigner,
		registeredInputs, privateKey, domainPublishers.GetPublisherKey)
	receiveNodeSetAlias := nodes.NewReceiveNodeAlias(domain, publisherID, nil, messageSigner,
		privateKey, domainPublishers.GetPublisherKey)

	var publisher = &Publisher{
		domainNodes:      domainNodes,
		domainInputs:     domainInputs,
		domainOutputs:    domainOutputs,
		domainPublishers: domainPublishers,

		exitChannel:        make(chan bool),
		fullIdentity:       identity,
		identityPrivateKey: privateKey,

		messenger:            messenger,
		messageSigner:        messageSigner,
		pollCountdown:        0,
		pollInterval:         DefaultPollInterval,
		receiveNodeConfigure: receiveNodeConfigure,
		receiveNodeSetAlias:  receiveNodeSetAlias,
		receiveSetInput:      receiveSetInput,

		registeredForecastValues: registeredForecastValues,
		registeredNodes:          registeredNodes,
		registeredInputs:         registeredInputs,
		registeredOutputs:        registeredOutputs,
		registeredOutputValues:   registeredOutputValues,

		signMessages: signMessages,
		updateMutex:  &sync.Mutex{},
	}

	return publisher
}
