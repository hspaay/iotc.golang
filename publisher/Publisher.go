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

	inputFromHTTP        *inputs.InputFromHTTP        // trigger inputs with http poll result
	inputFromFiles       *inputs.InputFromFiles       // trigger inputs on file changes
	inputFromOutputs     *inputs.InputFromOutputs     // subscribe input to an output (latest) value
	inputFromSetCommands *inputs.InputFromSetCommands // trigger inputs with set commands for registered inputs

	receiveNodeConfigure     *nodes.ReceiveNodeConfigure       // listener for node configure for registered nodes
	registeredForecastValues *outputs.RegisteredForecastValues // output forecasts values published by this publisher
	registeredInputs         *inputs.RegisteredInputs          // registered/published inputs from this publisher
	registeredNodes          *nodes.RegisteredNodes            // registered/published nodes from this publisher
	receiveNodeSetAlias      *nodes.ReceiveNodeAlias           // listener for set node alias
	registeredOutputs        *outputs.RegisteredOutputs        // registered/published outputs from this publisher
	registeredOutputValues   *outputs.RegisteredOutputValues   // registered/published output values from this publisher

	configFolder     string                    // folder to save publisher, node, inputs and outputs configuration
	domainPublishers *publishers.PublisherList // publishers on the network by discovery address

	fullIdentity        *types.PublisherFullIdentity                         // this publishers identity
	identityPrivateKey  *ecdsa.PrivateKey                                    // key for signing and encryption
	isRunning           bool                                                 // publisher was started and is running
	messenger           messaging.IMessenger                                 // Message bus messenger to use
	messageSigner       *messaging.MessageSigner                             // publishing signed messages
	onNodeConfigHandler nodes.NodeConfigureHandler                           // handle before applying configuration
	onNodeInputHandler  func(address string, message *types.SetInputMessage) // handle to update device/service input
	pollHandler         func(pub *Publisher)                                 // function that performs value polling
	pollCountdown       int                                                  // countdown each heartbeat
	pollInterval        int                                                  // value polling interval in seconds

	// background publications require a mutex to prevent concurrent access
	exitChannel  chan bool
	signMessages bool        // signing on or off
	updateMutex  *sync.Mutex // mutex for async updating and publishing
}

// Address returns the publisher's identity address
func (pub *Publisher) Address() string {
	// identityAddr := nodes.MakePublisherIdentityAddress(pub.Domain(), pub.PublisherID())
	// return identityAddr
	return pub.fullIdentity.Address
}

// PublisherID returns the publisher's ID
func (pub *Publisher) PublisherID() string {
	return pub.fullIdentity.PublisherID
}

// Domain returns the publication domain
func (pub *Publisher) Domain() string {
	return pub.fullIdentity.Domain
}

// FullIdentity return a copy of this publisher's full identity
func (pub *Publisher) FullIdentity() types.PublisherFullIdentity {
	return *pub.fullIdentity
}

// SetLogging sets the logging level and output file for this publisher
// Intended for setting logging from configuration
// levelName is the requested logging level: error, warning, info, debug
// filename is the output log file full name including path
func (pub *Publisher) SetLogging(levelName string, filename string) {
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
func (pub *Publisher) SetNodeConfigHandler(
	handler func(address string, config types.NodeAttrMap) types.NodeAttrMap) {

	pub.receiveNodeConfigure.SetConfigureNodeHandler(handler)
}

// LoadConfig loads previously saved configuration
// If a node file exists in the given folder the nodes will be added/updated. Existing nodes will be replaced.
// If autosave is set then save this publisher's nodes and configs when updated.
// folder with the configuration files. Use "" for default, which is persist.DefaultConfigFolder
//   autosave indicates to save updates to node configuration
// returns error if folder doesn't exist
func (pub *Publisher) LoadConfig(folder string, autosave bool) error {
	var err error = nil
	if folder == "" {
		folder = persist.DefaultConfigFolder
	}
	if autosave {
		pub.configFolder = folder
	}
	if folder != "" {
		nodeList := make([]*types.NodeDiscoveryMessage, 0)
		err = persist.LoadNodes(folder, pub.PublisherID(), &nodeList)
		if err == nil {
			pub.registeredNodes.UpdateNodes(nodeList)
		}
	}
	return err
}

// SetPollInterval is a convenience function for periodic polling of updates to registered
// nodes, inputs, outputs and output values.
// seconds interval to perform another poll. Default (0) is DefaultPollInterval
// intended for publishers that need to poll for values
func (pub *Publisher) SetPollInterval(seconds int, handler func(pub *Publisher)) {
	logrus.Infof("Publisher.SetPoll: interval = %d seconds", seconds)
	if seconds > 0 {
		pub.pollInterval = seconds
	} else {
		pub.pollInterval = DefaultPollInterval
	}
	pub.pollHandler = handler
}

// SetPublisherID changes the publisher's ID. Use before calling Start
// func (pub *Publisher) SetPublisherID(id string) {
// 	pub.id = id
// }

// SetSigningOnOff turns signing of publications on or off.
// the default is provided when creating the pub.
func (pub *Publisher) SetSigningOnOff(on bool) {
	pub.signMessages = on
}

// Start publishing and listen for configuration and input messages
// This will create the publisher node and load previously saved nodes
// Start will fail if no messenger has been provided.
// persistNodes will load previously saved nodes at startup and save them on configuration change
func (pub *Publisher) Start() {

	if pub.messenger == nil {
		logrus.Errorf("Publisher.Start: Can't start publisher %s/%s without a messenger. See SetMessenger()",
			pub.Domain(), pub.PublisherID())
		return
	}
	if !pub.isRunning {
		logrus.Warningf("Publisher.Start: Starting publisher %s/%s", pub.Domain(), pub.PublisherID())
		pub.updateMutex.Lock()
		pub.isRunning = true
		pub.updateMutex.Unlock()

		go pub.heartbeatLoop()
		// wait for the heartbeat to start
		<-pub.exitChannel

		// TODO: support LWT
		pub.messenger.Connect("", "")

		pub.domainNodes.Start()
		pub.domainInputs.Start()
		pub.domainOutputs.Start()
		pub.receiveNodeSetAlias.Start()
		pub.receiveNodeConfigure.Start()

		// subscribe to publisher nodes to verify signature for input commands
		pubAddr := publishers.MakePublisherIdentityAddress(pub.Domain(), "+")
		pub.messenger.Subscribe(pubAddr, pub.handlePublisherDiscovery)

		// publish discovery of this publisher
		pub.PublishIdentity(*pub.messageSigner)

		logrus.Infof("Publisher.Start: Publisher %s started", pub.PublisherID())
	}
}

// Stop publishing
// Wait until the heartbeat loop has finished processing messages
func (pub *Publisher) Stop() {
	logrus.Warningf("Publisher.Stop: Stopping publisher %s", pub.PublisherID())
	pub.updateMutex.Lock()
	if pub.isRunning {
		pub.isRunning = false

		pub.receiveNodeConfigure.Stop()
		pub.domainOutputs.Stop()
		pub.domainInputs.Stop()
		pub.domainNodes.Stop()

		pub.updateMutex.Unlock()
		// wait for heartbeat to end
		<-pub.exitChannel
	} else {
		pub.updateMutex.Unlock()
	}
	logrus.Info("... bye bye")
}

// WaitForSignal waits until a TERM or INT signal is received
func (pub *Publisher) WaitForSignal() {

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
func (pub *Publisher) heartbeatLoop() {
	logrus.Infof("Publisher.heartbeatLoop: starting heartbeat loop")
	pub.exitChannel <- false

	for {
		time.Sleep(time.Second)

		// FIXME: The duration of publishing these updates adds to the heartbeat which delays the heartbeat
		pub.PublishUpdates()

		// poll for discovery and values of registered nodes, inputs and outputs
		if (pub.pollCountdown <= 0) && (pub.pollHandler != nil) {
			pub.pollHandler(pub)
			pub.pollCountdown = pub.pollInterval
		}
		pub.pollCountdown--

		pub.updateMutex.Lock()
		isRunning := pub.isRunning
		pub.updateMutex.Unlock()
		if !isRunning {
			break
		}
	}
	pub.exitChannel <- true
	logrus.Infof("Publisher.heartbeatLoop: Ending loop of publisher %s", pub.PublisherID())
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
// domain and publisherID identity this pub. If the identity file does not match these, it
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
	domainInputs := inputs.NewDomainInputs(messageSigner)
	domainOutputs := outputs.NewDomainOutputs(messageSigner)
	domainNodes := nodes.NewDomainNodes(messageSigner)
	registeredNodes := nodes.NewRegisteredNodes(domain, publisherID)
	registeredInputs := inputs.NewRegisteredInputs(domain, publisherID)
	registeredOutputs := outputs.NewRegisteredOutputs(domain, publisherID)
	registeredOutputValues := outputs.NewRegisteredOutputValues(domain, publisherID)
	registeredForecastValues := outputs.NewRegisteredForecastValues(domain, publisherID)

	receiveNodeConfigure := nodes.NewReceiveNodeConfigure(
		domain, publisherID, nil, messageSigner, registeredNodes, privateKey)
	receiveNodeSetAlias := nodes.NewReceiveNodeAlias(
		domain, publisherID, nil, messageSigner, privateKey)

	var publisher = &Publisher{
		domainNodes:      domainNodes,
		domainInputs:     domainInputs,
		domainOutputs:    domainOutputs,
		domainPublishers: domainPublishers,

		inputFromSetCommands: inputs.NewInputFromSetCommands(
			domain, publisherID, messageSigner, registeredInputs),
		inputFromHTTP:    inputs.NewInputFromHTTP(registeredInputs),
		inputFromFiles:   inputs.NewInputFromFiles(registeredInputs),
		inputFromOutputs: inputs.NewInputFromOutputs(messageSigner, registeredInputs),

		exitChannel:        make(chan bool),
		fullIdentity:       identity,
		identityPrivateKey: privateKey,

		messenger:            messenger,
		messageSigner:        messageSigner,
		pollCountdown:        0,
		pollInterval:         DefaultPollInterval,
		receiveNodeConfigure: receiveNodeConfigure,
		receiveNodeSetAlias:  receiveNodeSetAlias,

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
