// Package publisher ...
// - Publishes updates to node, inputs and outputs when they are (re)discovered
// - configuration of nodes
// - control of inputs
// - update of security keys and identity signature
// Thread-safe. All public functions can be invoked from multiple goroutines
package publisher

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/hspaay/iotc.golang/iotc"
	"github.com/hspaay/iotc.golang/messenger"
	"github.com/hspaay/iotc.golang/nodes"
	"github.com/hspaay/iotc.golang/persist"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

// reserved keywords
const (
	// DefaultDiscoveryInterval in which node discovery information is republished
	DefaultDiscoveryInterval = 24 * 3600
	// DefaultPollInterval in which the output values are queried for polling based sources
	DefaultPollInterval = 24 * 3600
)

// NodeConfigHandler callback when command to update node config is received
type NodeConfigHandler func(node *iotc.NodeDiscoveryMessage, config iotc.NodeAttrMap) iotc.NodeAttrMap

// NodeInputHandler callback when command to update node input is received
type NodeInputHandler func(input *iotc.InputDiscoveryMessage, message *iotc.SetInputMessage)

// Publisher carries the operating state of 'this' publisher
type Publisher struct {
	Nodes           *nodes.NodeList           // discovered nodes published by this publisher
	Inputs          *nodes.InputList          // discovered inputs published by this publisher
	Outputs         *nodes.OutputList         // discovered outputs published by this publisher
	OutputForecasts *nodes.OutputForecastList // output forecasts values published by this publisher
	OutputValues    *nodes.OutputValueList    // output values published by this publisher

	// address             string                     // the publisher node address
	autosaveFolder      string                         // folder to save nodes after update
	discoverCountdown   int                            // countdown each heartbeat
	discoveryInterval   int                            // discovery polling interval
	discoveryHandler    func(publisher *Publisher)     // function that performs discovery
	identity            *iotc.PublisherIdentityMessage // identity for signing messages
	isRunning           bool                           // publisher was started and is running
	logger              *log.Logger                    // logger for all publisher's logging
	messenger           messenger.IMessenger           // Message bus messenger to use
	signer              messenger.IMessageSigner       // Signing and verification of messages
	onNodeConfigHandler NodeConfigHandler              // handle before applying configuration
	onNodeInputHandler  NodeInputHandler               // handle to update device/service input
	pollHandler         func(publisher *Publisher)     // function that performs value polling
	pollCountdown       int                            // countdown each heartbeat
	pollInterval        int                            // value polling interval in seconds
	publisherID         string                         // publisher ID
	privateKeySigning   *ecdsa.PrivateKey              // key for singing published messages

	domain           string               // The domain this publisher lives in
	domainPublishers *nodes.PublisherList // publishers on the network by discovery address

	// background publications require a mutex to prevent concurrent access
	exitChannel chan bool
	updateMutex *sync.Mutex // mutex for async updating and publishing
}

// Address  returns the publisher's identity address
func (publisher *Publisher) Address() string {
	identityAddr := nodes.MakePublisherIdentityAddress(publisher.domain, publisher.publisherID)
	return identityAddr
}

// PublisherID returns the publisher's ID
func (publisher *Publisher) PublisherID() string {
	return publisher.publisherID
}

// Domain returns the publication domain
func (publisher *Publisher) Domain() string {
	return publisher.domain
}

// Identity return this publisher's identity
func (publisher *Publisher) Identity() *iotc.PublisherIdentityMessage {
	return publisher.identity
}

// Logger returns the publication logger
func (publisher *Publisher) Logger() *log.Logger {
	return publisher.logger
}

// SetDiscoveryInterval is a convenience function for periodic update of discovered
// nodes, inputs and outputs. Intended for publishers that need to poll for discovery.
//
// interval in seconds to perform another discovery. Default is DefaultDiscoveryInterval
// handler is the callback with the publisher for publishing discovery
func (publisher *Publisher) SetDiscoveryInterval(interval int, handler func(publisher *Publisher)) {

	publisher.logger.Infof("Publisher.SetDiscoveryInterval: discovery interval = %d seconds", interval)
	if interval > 0 {
		publisher.discoveryInterval = interval
	}
	publisher.discoveryHandler = handler
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
			publisher.logger.Errorf("Publisher.SetLogging: Unable to open logfile: %s", err)
		} else {
			publisher.logger.Warnf("Publisher.SetLogging: Send logging output to %s", filename)
			logOut = logFileHandle
		}
	}

	publisher.logger = &logrus.Logger{
		Out:   logOut,
		Level: loggingLevel,
		// Formatter: &prefixed.TextFormatter{
		Formatter: &log.TextFormatter{
			// LogFormat: "",
			// DisableColors:   true,
			// DisableLevelTruncation: true,
			TimestampFormat: "2006-01-02 15:04:05.000",
			FullTimestamp:   true,
			// ForceFormatting: true,
		},
	}
	publisher.logger.SetReportCaller(false) // publisher logging includes caller and file:line#
}

// SetNodeConfigHandler set the handler for updating node configuration
func (publisher *Publisher) SetNodeConfigHandler(
	handler func(node *iotc.NodeDiscoveryMessage, config iotc.NodeAttrMap) iotc.NodeAttrMap) {
	publisher.onNodeConfigHandler = handler
}

// SetNodeInputHandler set the handler for updating node inputs
func (publisher *Publisher) SetNodeInputHandler(
	handler func(input *iotc.InputDiscoveryMessage, message *iotc.SetInputMessage)) {
	publisher.onNodeInputHandler = handler
}

// SetPersistNodes set the folder for loading and saving this publisher's nodes.
// If a node file exists in the given folder the nodes will be added/updated.
// Existing nodes will be replaced.
// If autosave is set then save this publisher's nodes and configs when updated.
//
// - folder of the configuration files.
//     Use "" for default, which is persist.DefaultConfigFolder: <userhome>/.config/iotconnect
// - autosave indicates to save updates to node configuration
// returns error if folder doesn't exist
func (publisher *Publisher) SetPersistNodes(folder string, autosave bool) error {
	var err error = nil
	if folder == "" {
		folder = persist.DefaultConfigFolder
	}
	if autosave {
		publisher.autosaveFolder = folder
	}
	if folder != "" {
		nodeList := make([]*iotc.NodeDiscoveryMessage, 0)
		err = persist.LoadNodes(folder, publisher.publisherID, &nodeList)
		if err == nil {
			publisher.Nodes.UpdateNodes(nodeList)
		}
	}
	return err
}

// SetPollInterval is a convenience function for periodic update of output values
// seconds interval to perform another poll. Default (0) is DefaultPollInterval
// intended for publishers that need to poll for values
func (publisher *Publisher) SetPollInterval(seconds int, handler func(publisher *Publisher)) {
	publisher.logger.Infof("Publisher.SetPoll: interval = %d seconds", seconds)
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

// SetupPublisherIdentity creates the publisher node with identity and public/private signing keys
// The identity is restored from storage to retain the keys signed by the ZCAS.
//
// * domain of this publisher
// * publisherID of this publisher
func (publisher *Publisher) SetupPublisherIdentity(domain string, publisherID string) {

	// publisher.address = pubNode.Address
	timestampStr := time.Now().Format(iotc.TimeFormat)
	validUntil := time.Now().Add(time.Hour * 24 * 365) // valid for 1 year
	validUntilStr := validUntil.Format(iotc.TimeFormat)

	// generate private/public key for signing and store the public key in the publisher identity
	// TODO: store keys
	rng := rand.Reader
	curve := elliptic.P256()
	privKey, err := ecdsa.GenerateKey(curve, rng)
	publisher.privateKeySigning = privKey
	if err != nil {
		publisher.logger.Errorf("Publisher.NewPublisher: Failed to create keys for signing: %s", err)
	}
	privStr, pubStr := messenger.EncodeKeys(privKey, &privKey.PublicKey)
	_ = privStr

	addr := publisher.Address()
	publisher.signer = NewJWSSigner(publisher.logger, addr, privKey)

	publisher.identity = &iotc.PublisherIdentityMessage{
		Address: addr,
		Identity: iotc.PublisherIdentity{
			Domain:       domain,
			IssuerName:   publisherID, // self issued, will be replaced by ZCAS
			Location:     "local",
			Organization: "", // todo: get from messenger configuration
			// PublicKeyCrypto:  cryptoStr,
			PublicKeySigning: pubStr,
			PublisherID:      publisherID,
			Timestamp:        timestampStr,
			ValidUntil:       validUntilStr,
		},
		Signature: "",
		Signer:    publisherID,
		Timestamp: timestampStr,
	}
}

// Start publishing and listen for configuration and input messages
// This will create the publisher node and load previously saved nodes
// Start will fail if no messenger has been provided.
// persistNodes will load previously saved nodes at startup and save them on configuration change
func (publisher *Publisher) Start() {

	if publisher.messenger == nil {
		publisher.logger.Errorf("Publisher.Start: Can't start publisher %s/%s without a messenger. See SetMessenger()",
			publisher.domain, publisher.publisherID)
		return
	}
	if !publisher.isRunning {
		publisher.logger.Warningf("Publisher.Start: Starting publisher %s/%s", publisher.domain, publisher.publisherID)
		publisher.updateMutex.Lock()
		publisher.isRunning = true
		publisher.updateMutex.Unlock()

		go publisher.heartbeatLoop()
		// wait for the heartbeat to start
		<-publisher.exitChannel

		// TODO: support LWT
		publisher.messenger.Connect("", "")

		// Subscribe to receive configuration and set messages for any of our nodes
		configAddr := nodes.MakeNodeAddress(publisher.domain, publisher.publisherID, "+", iotc.MessageTypeConfigure)
		publisher.messenger.Subscribe(configAddr, publisher.handleNodeConfigCommand)

		inputAddr := nodes.MakeInputSetAddress(configAddr, "+", "+")
		publisher.messenger.Subscribe(inputAddr, publisher.handleNodeInput)

		// subscribe to publisher nodes to verify signature for input commands
		pubAddr := nodes.MakePublisherIdentityAddress(publisher.domain, "+")
		publisher.messenger.Subscribe(pubAddr, publisher.handlePublisherDiscovery)

		// publish discovery of this publisher
		publisher.PublishIdentity()
		publisher.PublishUpdatedDiscoveries()

		publisher.logger.Warningf("Publisher.Start: Publisher %s started", publisher.publisherID)
	}
}

// Stop publishing
// Wait until the heartbeat loop has finished processing messages
func (publisher *Publisher) Stop() {
	publisher.logger.Warningf("Publisher.Stop: Stopping publisher %s", publisher.publisherID)
	publisher.updateMutex.Lock()
	if publisher.isRunning {
		publisher.isRunning = false
		// go messenger.NewDummyMessenger().Disconnect()
		publisher.updateMutex.Unlock()
		// wait for heartbeat to end
		<-publisher.exitChannel
	} else {
		publisher.updateMutex.Unlock()
	}
	publisher.logger.Info("... bye bye")
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
	publisher.logger.Warningf("Publisher.heartbeatLoop: starting heartbeat loop")
	publisher.exitChannel <- false

	for {
		time.Sleep(time.Second)

		// FIXME: the publishUpdates duration adds to the heartbeat. This can also take a
		//  while unless the messenger unloads using channels (which it should)
		//  we want to be sure it has completed when the heartbeat ends
		publisher.PublishUpdatedDiscoveries()
		publisher.PublishUpdatedOutputValues()

		// discover new nodes
		if (publisher.discoverCountdown <= 0) && (publisher.discoveryHandler != nil) {
			go publisher.discoveryHandler(publisher)
			publisher.discoverCountdown = publisher.discoveryInterval
		}
		publisher.discoverCountdown--

		// poll for values
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
	publisher.logger.Warningf("Publisher.heartbeatLoop: Ending loop of publisher %s", publisher.publisherID)
}

// // VerifyMessageSignature Verify a received message is signed by the sender
// // The node of the sender must have been received for its public key
// func (publisher *Publisher) VerifyMessageSignature(
// 	sender string, message string, base64signature string) bool {

// 	if node == nil {
// 		publisher.Logger.Warningf("Publisher.VerifyMessageSignature: unknown sender %s", sender)
// 		return false
// 	}
// 	var pubKey *ecdsa.PublicKey = messenger.DecodePublicKey(publisher.identity.PublicKeySigning)
// 	valid := messenger.VerifyEcdsaSignature(message, base64signature, pubKey)
// 	return valid
// }

// NewPublisher creates a publisher instance and node for use in publications
//
// appID is the application ID used to load the publisher configuration and nodes
//     <appID.yaml> for the publisher configuration -> publisherID
//     <appID-nodes.json> for the nodes
// messenger to use fo publications and for the domain to publish in
// logger is the optional logger to use.
//
// domain the publisher uses to create addresses. If not provided iotc.LocalDomain is used
// publisherID of this publisher, unique within the domain. See also SetPublisherID
// messenger for publishing onto the message bus
func NewPublisher(
	domain string,
	publisherID string,
	messenger messenger.IMessenger,
) *Publisher {

	if domain == "" {
		domain = iotc.LocalDomainID
	}

	// IotConnect core running state of the publisher
	var publisher = &Publisher{
		discoveryInterval: DefaultDiscoveryInterval,
		exitChannel:       make(chan bool),
		publisherID:       publisherID,
		Inputs:            nodes.NewInputList(),
		messenger:         messenger,
		Nodes:             nodes.NewNodeList(),
		Outputs:           nodes.NewOutputList(),
		OutputValues:      nodes.NewOutputValueList(),
		OutputForecasts:   nodes.NewOutputForecastList(),
		pollCountdown:     1, // run discovery before poll
		pollInterval:      DefaultPollInterval,
		// address:           pubNode.Address,
		updateMutex:      &sync.Mutex{},
		domain:           domain,
		domainPublishers: nodes.NewPublisherList(),
	}
	publisher.SetLogging("debug", "")

	// create a default publisher node with identity and signatures
	publisher.SetupPublisherIdentity(domain, publisherID)

	return publisher
}
