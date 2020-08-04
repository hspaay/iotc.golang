// Package publisher ...
// - Publishes updates to node, inputs and outputs when they are (re)discovered
// - configuration of nodes
// - control of inputs
// - update of security keys and identity signature
// Thread-safe. All public functions can be invoked from multiple goroutines
package publisher

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/iotdomain/iotdomain-go/identities"
	"github.com/iotdomain/iotdomain-go/inputs"
	"github.com/iotdomain/iotdomain-go/lib"
	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/iotdomain/iotdomain-go/nodes"
	"github.com/iotdomain/iotdomain-go/outputs"
	"github.com/iotdomain/iotdomain-go/persist"
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
	domainIdentities   *identities.DomainPublisherIdentities // discovered publisher identities
	domainInputs       *inputs.DomainInputs                  // discovered inputs from the domain
	domainNodes        *nodes.DomainNodes                    // discovered nodes from the domain
	domainOutputs      *outputs.DomainOutputs                // discovered outputs from the domain
	domainOutputValues *outputs.DomainOutputValues           // output values from the domain

	inputFromHTTP        *inputs.ReceiveFromHTTP        // trigger inputs with http poll result
	inputFromFiles       *inputs.ReceiveFromFiles       // trigger inputs on file changes
	inputFromOutputs     *inputs.ReceiveFromOutputs     // subscribe input to an output (latest) value
	inputFromSetCommands *inputs.ReceiveFromSetCommands // trigger inputs with set commands for registered inputs

	receiveMyIdentityUpdate *identities.ReceiveRegisteredIdentityUpdate
	receiveDomainIdentities *identities.ReceiveDomainPublisherIdentities // listener for identity updates
	receiveNodeConfigure    *nodes.ReceiveNodeConfigure                  // listener for node configure for registered nodes
	receiveNodeSetAlias     *nodes.ReceiveNodeAlias                      // listener for set node alias

	registeredForecastValues *outputs.RegisteredForecastValues // output forecasts values published by this publisher
	registeredIdentity       *identities.RegisteredIdentity    // registered/published identity of this publisher
	registeredInputs         *inputs.RegisteredInputs          // registered/published inputs from this publisher
	registeredNodes          *nodes.RegisteredNodes            // registered/published nodes from this publisher
	registeredOutputs        *outputs.RegisteredOutputs        // registered/published outputs from this publisher
	registeredOutputValues   *outputs.RegisteredOutputValues   // registered/published output values from this publisher

	configFolder string // folder to save publisher, node, inputs and outputs configuration
	// domainPublishers *publishers.PublisherList // publishers on the network by discovery address

	// fullIdentity        *types.PublisherFullIdentity                         // this publishers identity
	// identityPrivateKey  *ecdsa.PrivateKey                                    // key for signing and encryption
	isRunning           bool                                                 // publisher was started and is running
	messenger           messaging.IMessenger                                 // Message bus messenger to use
	messageSigner       *messaging.MessageSigner                             // publishing signed messages
	onNodeConfigHandler nodes.NodeConfigureHandler                           // handle before applying configuration
	onNodeInputHandler  func(address string, message *types.SetInputMessage) // handle to update device/service input
	pollHandler         func(pub *Publisher)                                 // function that performs value polling
	pollCountdown       int                                                  // countdown each heartbeat
	pollInterval        int                                                  // value polling interval in seconds

	// background publications require a mutex to prevent concurrent access
	heartbeatChannel chan bool
	updateMutex      *sync.Mutex // mutex for async updating and publishing
}

// // FullIdentity return a copy of this publisher's full identity
// func (pub *Publisher) FullIdentity() types.PublisherFullIdentity {
// 	ident, _ := pub.registeredIdentity.GetIdentity()
// 	return *ident
// }

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

// SetLogging sets the logging level and output file for this publisher
// Intended for setting logging from configuration
//  levelName is the requested logging level: error, warning, info, debug
//  filename is the output log file full name including path, use "" for stderr
func (pub *Publisher) SetLogging(levelName string, filename string) error {
	loggingLevel := log.DebugLevel
	var err error

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
			err = lib.MakeErrorf("Publisher.SetLogging: Unable to open logfile: %s", err)
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
	return err
}

// SetNodeConfigHandler set the handler for updating node configuration
func (pub *Publisher) SetNodeConfigHandler(
	handler func(nodeAddress string, config types.NodeAttrMap) types.NodeAttrMap) {

	pub.receiveNodeConfigure.SetConfigureNodeHandler(handler)
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

// Start publishing and listen for configuration and input messages
// This will create the publisher node and load previously saved nodes
// Start will fail if no messenger has been provided.
// persistNodes will load previously saved nodes at startup and save them on configuration change
func (pub *Publisher) Start() {
	logrus.Warningf("Publisher.Start: Starting publisher %s/%s", pub.Domain(), pub.PublisherID())

	if pub.messenger == nil {
		logrus.Errorf("Publisher.Start: Can't start publisher %s/%s without a messenger. See SetMessenger()",
			pub.Domain(), pub.PublisherID())
		return
	}
	if !pub.isRunning {
		pub.updateMutex.Lock()
		pub.isRunning = true
		pub.updateMutex.Unlock()

		go pub.heartbeatLoop()
		// wait for the heartbeat to start
		<-pub.heartbeatChannel

		// our own identity is first
		myIdent, _ := pub.registeredIdentity.GetIdentity()
		pub.domainIdentities.AddIdentity(&myIdent.PublisherIdentityMessage)

		// TODO: support LWT
		// receive domain entities, eg identities, nodes, inputs and outputs
		pub.receiveDomainIdentities.Start()
		pub.domainNodes.Start()
		pub.domainInputs.Start()
		pub.domainOutputs.Start()
		// receive commands
		pub.receiveNodeSetAlias.Start()
		pub.receiveNodeConfigure.Start()
		pub.receiveMyIdentityUpdate.Start()
		//  listening
		pub.messenger.Connect("", "")
		pub.registeredIdentity.PublishIdentity(pub.messageSigner)
	}
}

// Stop publishing
// Wait until the heartbeat loop has finished processing messages
func (pub *Publisher) Stop() {
	logrus.Warningf("Publisher.Stop: Stopping publisher %s", pub.PublisherID())
	pub.updateMutex.Lock()
	if pub.isRunning {
		pub.isRunning = false

		pub.receiveMyIdentityUpdate.Stop()
		pub.receiveDomainIdentities.Stop()
		pub.receiveNodeConfigure.Stop()
		pub.receiveNodeSetAlias.Stop()
		pub.domainOutputs.Stop()
		pub.domainInputs.Stop()
		pub.domainNodes.Stop()

		pub.updateMutex.Unlock()
		// wait for heartbeat to end
		<-pub.heartbeatChannel
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
	pub.heartbeatChannel <- false

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
	pub.heartbeatChannel <- true
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
	messenger messaging.IMessenger,
) *Publisher {

	logrus.SetLevel(log.DebugLevel)
	if domain == "" {
		domain = types.LocalDomainID
	}
	// ours and domain identities
	registeredIdentity, privateKey := identities.NewRegisteredIdentity(
		identityFolder, domain, publisherID)
	domainIdentities := identities.NewDomainPublisherIdentities()

	// These are the basis for signing and identifying publishers
	messageSigner := messaging.NewMessageSigner(messenger, privateKey, domainIdentities.GetPublisherKey)

	// application services
	domainInputs := inputs.NewDomainInputs(messageSigner)
	domainNodes := nodes.NewDomainNodes(messageSigner)
	domainOutputs := outputs.NewDomainOutputs(messageSigner)
	domainOutputValues := outputs.NewDomainOutputValues(messageSigner)
	registeredInputs := inputs.NewRegisteredInputs(domain, publisherID)
	registeredNodes := nodes.NewRegisteredNodes(domain, publisherID)
	registeredOutputs := outputs.NewRegisteredOutputs(domain, publisherID)
	registeredOutputValues := outputs.NewRegisteredOutputValues(domain, publisherID)
	registeredForecastValues := outputs.NewRegisteredForecastValues(domain, publisherID)

	receiveMyIdentityUpdate := identities.NewReceiveRegisteredIdentityUpdate(
		registeredIdentity, messageSigner)
	receiveDomainIdentities := identities.NewReceivePublisherIdentities(domain,
		domainIdentities, messageSigner)
	receiveNodeConfigure := nodes.NewReceiveNodeConfigure(
		domain, publisherID, nil, messageSigner, registeredNodes, privateKey)
	receiveNodeSetAlias := nodes.NewReceiveNodeAlias(
		domain, publisherID, nil, messageSigner, privateKey)

	var publisher = &Publisher{
		domainIdentities:   domainIdentities,
		domainInputs:       domainInputs,
		domainNodes:        domainNodes,
		domainOutputs:      domainOutputs,
		domainOutputValues: domainOutputValues,

		inputFromSetCommands: inputs.NewReceiveFromSetCommands(
			domain, publisherID, messageSigner, registeredInputs),
		inputFromHTTP:    inputs.NewReceiveFromHTTP(registeredInputs),
		inputFromFiles:   inputs.NewReceiveFromFiles(registeredInputs),
		inputFromOutputs: inputs.NewReceiveFromOutputs(messageSigner, registeredInputs),

		heartbeatChannel: make(chan bool),
		// fullIdentity:       identity,
		// identityPrivateKey: privateKey,

		messenger:               messenger,
		messageSigner:           messageSigner,
		pollCountdown:           0,
		pollInterval:            DefaultPollInterval,
		receiveDomainIdentities: receiveDomainIdentities,
		receiveMyIdentityUpdate: receiveMyIdentityUpdate,
		receiveNodeConfigure:    receiveNodeConfigure,
		receiveNodeSetAlias:     receiveNodeSetAlias,

		registeredForecastValues: registeredForecastValues,
		registeredIdentity:       registeredIdentity,
		registeredInputs:         registeredInputs,
		registeredNodes:          registeredNodes,
		registeredOutputs:        registeredOutputs,
		registeredOutputValues:   registeredOutputValues,

		updateMutex: &sync.Mutex{},
	}
	receiveNodeSetAlias.SetAliasHandler(publisher.HandleAliasCommand)

	return publisher
}
