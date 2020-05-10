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
	"encoding/json"
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
	address             string                     // the publisher node address
	autosaveFolder      string                     // folder to save nodes after update
	discoverCountdown   int                        // countdown each heartbeat
	discoveryInterval   int                        // discovery polling interval
	discoveryHandler    func(publisher *Publisher) // function that performs discovery
	id                  string                     // publisher ID
	Logger              *log.Logger                //
	messenger           messenger.IMessenger       // Message bus messenger to use
	onNodeConfigHandler NodeConfigHandler          // handle before applying configuration
	onNodeInputHandler  NodeInputHandler           // handle to update device/service input

	pollHandler    func(publisher *Publisher)            // function that performs value polling
	pollCountdown  int                                   // countdown each heartbeat
	pollInterval   int                                   // value polling interval in seconds
	zonePublishers map[string]*iotc.NodeDiscoveryMessage // publishers on the network
	signPrivateKey *ecdsa.PrivateKey                     // key for singing published messages
	Zone           string                                // The zone this publisher lives in

	// background publications require a mutex to prevent concurrent access
	exitChannel chan bool
	updateMutex *sync.Mutex // mutex for async updating and publishing

	Nodes          *nodes.NodeList                   // nodes published by this publisher
	isRunning      bool                              // publisher was started and is running
	Inputs         *nodes.InputList                  // inputs published by this publisher
	Outputs        *nodes.OutputList                 // outputs published by this publisher
	outputForecast map[string]iotc.OutputHistoryList // output forecast by address
	OutputValues   *nodes.OutputValueList            // output values and history published by this publisher
}

// Address  returns the publisher's node address
func (publisher *Publisher) Address() string {
	return publisher.address
}

// ID returns the publisher ID
func (publisher *Publisher) ID() string {
	return publisher.id
}

// GetConfigValue convenience function to get a configuration value
// This retuns the 'default' value if no value is set
func GetConfigValue(configMap map[string]iotc.ConfigAttr, attrName string) string {
	config, configExists := configMap[attrName]
	if !configExists {
		return ""
	}
	if config.Value == "" {
		return config.Default
	}
	return config.Value
}

// GetNodeByID returns a node from this publisher or nil if the id isn't found in this publisher
// This is a convenience function as publishers tend to do this quite often
func (publisher *Publisher) GetNodeByID(id string) *iotc.NodeDiscoveryMessage {
	node := publisher.Nodes.GetNodeByID(publisher.Zone, publisher.id, id)
	return node
}

// PublisherNode return this publisher's node
func (publisher *Publisher) PublisherNode() *iotc.NodeDiscoveryMessage {
	node := publisher.Nodes.GetNodeByID(publisher.Zone, publisher.id, iotc.PublisherNodeID)
	return node
}

// SetDiscoveryInterval is a convenience function for periodic update of discovered
// nodes, inputs and outputs. Intended for publishers that need to poll for discovery.
//
// interval in seconds to perform another discovery. Default is DefaultDiscoveryInterval
// handler is the callback with the publisher for publishing discovery
func (publisher *Publisher) SetDiscoveryInterval(interval int, handler func(publisher *Publisher)) {

	publisher.Logger.Infof("discovery interval = %d seconds", interval)
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
			publisher.Logger.Errorf("MyZoneService.SetLogging: Unable to open logfile: %s", err)
		} else {
			publisher.Logger.Warnf("MyZoneService.SetLogging: Send logging output to %s", filename)
			logOut = logFileHandle
		}
	}

	publisher.Logger = &logrus.Logger{
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
	publisher.Logger.SetReportCaller(false) // publisher logging includes caller and file:line#
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
		err = persist.LoadNodes(folder, publisher.id, &nodeList)
		if err == nil {
			publisher.Nodes.UpdateNodes(nodeList)
		}
	}
	return err
}

// SetPollingInterval is a convenience function for periodic update of output values
// interval in seconds to perform another poll. Default is DefaultPollInterval
// intended for publishers that need to poll for values
func (publisher *Publisher) SetPollingInterval(interval int, handler func(publisher *Publisher)) {
	publisher.Logger.Infof("polling interval = %d seconds", interval)
	if interval > 0 {
		publisher.pollInterval = interval
	}
	publisher.pollHandler = handler
}

// SetPollInterval determines the interval between polling
func (publisher *Publisher) SetPollInterval(seconds int,
	handler func(publisher *Publisher)) {
	publisher.pollInterval = seconds
	publisher.pollHandler = handler
}

// SetNodeConfigHandler set the handler for updating node configuration
func (publisher *Publisher) SetNodeConfigHandler(
	handler func(node *iotc.NodeDiscoveryMessage, config iotc.NodeAttrMap) iotc.NodeAttrMap) {
	publisher.onNodeConfigHandler = handler
}

// SetNodeInputHandler set the handler for updating node inputs
func (publisher *Publisher) SetNodeInputHandler(handler func(input *iotc.InputDiscoveryMessage, message *iotc.SetInputMessage)) {
	publisher.onNodeInputHandler = handler
}

// Start publishing and listen for configuration and input messages
// This will load previously saved nodes
// Start will fail if no messenger has been provided.
// persistNodes will load previously saved nodes at startup and save them on configuration change
func (publisher *Publisher) Start() {

	if publisher.messenger == nil {
		publisher.Logger.Errorf("Can't start publisher %s without a messenger. See SetMessenger()", publisher.id)
		return
	}
	if !publisher.isRunning {
		publisher.Logger.Warningf("Starting publisher %s", publisher.id)
		publisher.updateMutex.Lock()
		publisher.isRunning = true
		publisher.updateMutex.Unlock()

		go publisher.heartbeatLoop()
		// wait for the heartbeat to start
		<-publisher.exitChannel

		// TODO: support LWT
		publisher.messenger.Connect("", "")

		// Subscribe to receive configuration and set messages for any of our nodes
		configAddr := fmt.Sprintf("%s/%s/+/%s", publisher.Zone, publisher.id, iotc.MessageTypeConfigure)
		publisher.messenger.Subscribe(configAddr, publisher.handleNodeConfigCommand)

		inputAddr := fmt.Sprintf("%s/%s/+/%s/+/+", publisher.Zone, publisher.id, iotc.MessageTypeSet)
		publisher.messenger.Subscribe(inputAddr, publisher.handleNodeInput)

		// subscribe to publisher nodes to verify signature for input commands
		pubAddr := fmt.Sprintf("%s/+/%s/%s", publisher.Zone, iotc.PublisherNodeID, iotc.MessageTypeNodeDiscovery)
		publisher.messenger.Subscribe(pubAddr, publisher.handlePublisherDiscovery)

		// publish discovery of this publisher
		publisher.PublishUpdates()

		publisher.Logger.Warningf("Publisher %s started", publisher.id)
	}
}

// Stop publishing
// Wait until the heartbeat loop has finished processing messages
func (publisher *Publisher) Stop() {
	publisher.Logger.Warningf("Stopping publisher %s", publisher.id)
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
	publisher.Logger.Info("... bye bye")
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
	publisher.Logger.Warningf("starting heartbeat loop")
	publisher.exitChannel <- false

	for {
		time.Sleep(time.Second)

		// FIXME: the publishUpdates duration adds to the heartbeat. This can also take a
		//  while unless the messenger unloads using channels (which it should)
		//  we want to be sure it has completed when the heartbeat ends
		publisher.PublishUpdates()
		publisher.publishOutputValues()

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
	publisher.Logger.Warningf("Ending loop of publisher %s", publisher.id)
}

// VerifyMessageSignature Verify a received message is signed by the sender
// The node of the sender must have been received for its public key
func (publisher *Publisher) VerifyMessageSignature(
	sender string, message json.RawMessage, base64signature string) bool {

	publisher.updateMutex.Lock()
	node := publisher.zonePublishers[sender]
	publisher.updateMutex.Unlock()

	if node == nil {
		publisher.Logger.Warningf("VerifyMessageSignature unknown sender %s", sender)
		return false
	}
	var pubKey *ecdsa.PublicKey = messenger.DecodePublicKey(node.Identity.PublicKeySigning)
	valid := messenger.VerifyEcdsaSignature(message, base64signature, pubKey)
	return valid
}

// NewPublisher creates a publisher instance and node for use in publications
//
// appID is the application ID used to load the publisher configuration and nodes
//     <appID.yaml> for the publisher configuration -> publisherID
//     <appID-nodes.json> for the nodes
// messenger to use fo publications and for the zone to publish in
// logger is the optional logger to use.
//
// zone the publisher uses to create addresses
// publisherID of this publisher, unique within the zone
// messenger for publishing onto the message bus
func NewPublisher(
	zone string,
	publisherID string,
	sender messenger.IMessenger,
) *Publisher {
	var pubNode = nodes.NewNode(zone, publisherID, iotc.PublisherNodeID, iotc.NodeTypeAdapter)

	// IotConnect core running state of the publisher
	var publisher = &Publisher{
		discoveryInterval: DefaultDiscoveryInterval,
		exitChannel:       make(chan bool),
		id:                publisherID,
		Inputs:            nodes.NewInputList(),
		messenger:         sender,
		Nodes:             nodes.NewNodeList(),
		Outputs:           nodes.NewOutputList(),
		OutputValues:      nodes.NewOutputValue(),
		outputForecast:    make(map[string]iotc.OutputHistoryList),
		pollCountdown:     1, // run discovery before poll
		pollInterval:      DefaultPollInterval,
		address:           pubNode.Address,
		updateMutex:       &sync.Mutex{},
		Zone:              zone,
		zonePublishers:    make(map[string]*iotc.NodeDiscoveryMessage),
	}
	publisher.SetLogging("debug", "")

	// generate private/public key for signing and store the public key in the publisher identity
	// TODO: store keys
	rng := rand.Reader
	curve := elliptic.P256()
	privKey, err := ecdsa.GenerateKey(curve, rng)
	publisher.signPrivateKey = privKey
	if err != nil {
		publisher.Logger.Errorf("Failed to create keys for signing: %s", err)
	}
	privStr, pubStr := messenger.EncodeKeys(privKey, &privKey.PublicKey)
	_ = privStr

	timeStampStr := time.Now().Format("2006-01-02T15:04:05.000-0700")
	pubNode.Identity = &iotc.PublisherIdentity{
		Address:          pubNode.Address,
		PublicKeySigning: pubStr,
		Publisher:        publisherID,
		Timestamp:        timeStampStr,
		Zone:             zone,
	}

	publisher.Nodes.UpdateNode(pubNode)
	return publisher
}
