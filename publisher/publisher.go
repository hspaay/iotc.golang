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

	"github.com/hspaay/iotconnect.golang/messenger"
	"github.com/hspaay/iotconnect.golang/nodes"
	"github.com/hspaay/iotconnect.golang/standard"
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

// PublisherState carries the operating state of 'this' publisher
type PublisherState struct {
	discoverCountdown   int                                                        // countdown each heartbeat
	discoveryInterval   int                                                        // discovery polling interval
	discoveryHandler    func(publisher *PublisherState)                            // function that performs discovery
	Logger              *log.Logger                                                //
	messenger           messenger.IMessenger                                       // Message bus messenger to use
	onNodeConfigHandler func(node *nodes.Node, config nodes.AttrMap) nodes.AttrMap // handle before applying configuration
	onNodeInputHandler  func(input *nodes.Input, message *standard.SetMessage)     // handle to update device/service input

	pollHandler    func(publisher *PublisherState) // function that performs value polling
	pollCountdown  int                             // countdown each heartbeat
	pollInterval   int                             // value polling interval in seconds
	publisherID    string                          // for easy access to the pub ID
	PublisherNode  *nodes.Node                     // This publisher's node
	zonePublishers map[string]*nodes.Node          // publishers on the network
	signPrivateKey *ecdsa.PrivateKey               // key for singing published messages
	Zone           string                          // The zone this publisher lives in

	// background publications require a mutex to prevent concurrent access
	exitChannel chan bool
	updateMutex *sync.Mutex                     // mutex for async updating and publishing
	configs     map[string]*nodes.ConfigAttrMap // node configuration
	// nodes               map[string]*nodes.Node          // nodes by discovery address
	Nodes          *nodes.NodeList                 // Node management
	isRunning      bool                            // publisher was started and is running
	Inputs         *nodes.InputList                // Node input management
	Outputs        *nodes.OutputList               // Node output management
	outputForecast map[string]standard.HistoryList // output forecast by address
	OutputHistory  *nodes.OutputHistoryList        // output history by address
}

// GetConfigValue convenience function to get a configuration value
// This retuns the 'default' value if no value is set
func GetConfigValue(configMap nodes.ConfigAttrMap, attrName string) string {
	config, configExists := configMap[attrName]
	if !configExists {
		return ""
	}
	if config.Value == "" {
		return config.Default
	}
	return config.Value
}

// GetNode returns a node from this publisher or nil if the id isn't found in this publisher
// This is a convenience function as publishers tend to do this quite often
func (publisher *PublisherState) GetNode(id string) *nodes.Node {
	node := publisher.Nodes.GetNodeByID(publisher.Zone, publisher.publisherID, id)
	return node
}

// SetErrorStatus provides the error reported by an output
func (publisher *PublisherState) SetErrorStatus(node *nodes.Node, errorText string) {
	if node != nil {
		// TODO: track errors in node status
	}
}

// SetDiscoveryInterval is a convenience function for periodic update of discovered
// nodes, inputs and outputs. Intended for publishers that need to poll for discovery.
//
// interval in seconds to perform another discovery. Default is DefaultDiscoveryInterval
// handler is the callback with the publisher for publishing discovery
func (publisher *PublisherState) SetDiscoveryInterval(interval int, handler func(publisher *PublisherState)) {

	publisher.Logger.Infof("discovery interval = %d seconds", interval)
	if interval > 0 {
		publisher.discoveryInterval = interval
	}
	publisher.discoveryHandler = handler
}

// SetPollingInterval is a convenience function for periodic update of output values
// interval in seconds to perform another poll. Default is DefaultPollInterval
// intended for publishers that need to poll for values
func (publisher *PublisherState) SetPollingInterval(interval int, handler func(publisher *PublisherState)) {
	publisher.Logger.Infof("polling interval = %d seconds", interval)
	if interval > 0 {
		publisher.pollInterval = interval
	}
	publisher.pollHandler = handler
}

// SetPollInterval determines the interval between polling
func (publisher *PublisherState) SetPollInterval(seconds int,
	handler func(publisher *PublisherState)) {
	publisher.pollInterval = seconds
	publisher.pollHandler = handler
}

// SetNodeConfigHandler set the handler for updating node configuration
func (publisher *PublisherState) SetNodeConfigHandler(handler func(node *nodes.Node, config nodes.AttrMap) nodes.AttrMap) {
	publisher.onNodeConfigHandler = handler
}

// SetNodeInputHandler set the handler for updating node inputs
func (publisher *PublisherState) SetNodeInputHandler(handler func(input *nodes.Input, message *standard.SetMessage)) {
	publisher.onNodeInputHandler = handler
}

// Start publishing and listen for configuration and input messages
// synchroneous publications for testing
// onConfig handles updates to configuration, nil if no config to process
// onSetInput handles commands to update inputs, nil if there are no inputs to control
func (publisher *PublisherState) Start() {

	if !publisher.isRunning {
		publisher.Logger.Warningf("Starting publisher %s", publisher.publisherID)
		publisher.updateMutex.Lock()
		publisher.isRunning = true
		publisher.updateMutex.Unlock()
		go publisher.heartbeatLoop()
		// wait for the heartbeat to start
		<-publisher.exitChannel

		// TODO: support LWT
		publisher.messenger.Connect("", "")

		// Subscribe to receive configuration and set messages
		configAddr := fmt.Sprintf("%s/%s/+/%s", publisher.Zone, publisher.publisherID, standard.CommandConfigure)
		publisher.messenger.Subscribe(configAddr, publisher.handleNodeConfigCommand)

		inputAddr := fmt.Sprintf("%s/%s/+/%s/+/+", publisher.Zone, publisher.publisherID, standard.CommandSet)
		publisher.messenger.Subscribe(inputAddr, publisher.handleNodeInput)

		// subscribe to publisher nodes to verify signature for input commands
		pubAddr := fmt.Sprintf("%s/+/%s/%s", publisher.Zone, nodes.PublisherNodeID, standard.CommandNodeDiscovery)
		publisher.messenger.Subscribe(pubAddr, publisher.handlePublisherDiscovery)

		// publish discovery of this publisher
		publisher.PublishUpdates()

		publisher.Logger.Warningf("Publisher %s started", publisher.publisherID)
	}
}

// Stop publishing
// Wait until the heartbeat loop has finished processing messages
func (publisher *PublisherState) Stop() {
	publisher.Logger.Warningf("Stopping publisher %s", publisher.publisherID)
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

// VerifyMessageSignature Verify the message is signed by the sender
// The node of the sender must have been received for its public key
func (publisher *PublisherState) VerifyMessageSignature(
	sender string, message json.RawMessage, base64signature string) bool {

	publisher.updateMutex.Lock()
	node := publisher.zonePublishers[sender]
	publisher.updateMutex.Unlock()

	if node == nil {
		publisher.Logger.Warningf("VerifyMessageSignature unknown sender %s", sender)
		return false
	}
	var pubKey *ecdsa.PublicKey = standard.DecodePublicKey(node.Identity.PublicKeySigning)
	valid := standard.VerifyEcdsaSignature(message, base64signature, pubKey)
	return valid
}

// WaitForSignal waits until a TERM or INT signal is received
func (publisher *PublisherState) WaitForSignal() {

	// catch all signals since not explicitly listing
	exitChannel := make(chan os.Signal, 1)

	//signal.Notify(exitChannel, syscall.SIGTERM|syscall.SIGHUP|syscall.SIGINT)
	signal.Notify(exitChannel, syscall.SIGINT, syscall.SIGTERM)

	sig := <-exitChannel
	log.Warningf("RECEIVED SIGNAL: %s", sig)
	fmt.Println()
	fmt.Println(sig)
}

// handle an incoming a set command for one of our nodes. This:
// - check if the signature is valid
// - check if the node is valid
// - pass the input value update to the adapter's onNodeInputHandler callback
func (publisher *PublisherState) handleNodeInput(address string, publication *messenger.Publication) {
	// Check that address is one of our inputs
	segments := strings.Split(address, "/")
	segments[3] = standard.CommandInputDiscovery
	inputAddr := strings.Join(segments, "/")

	input := publisher.Inputs.GetInputByAddress(inputAddr)

	if input == nil || publication.Message == nil {
		publisher.Logger.Infof("handleNodeInput unknown input for address %s or missing message", address)
		return
	}
	// Decode the message into a SetMessage type
	var setMessage standard.SetMessage
	err := json.Unmarshal([]byte(publication.Message), &setMessage)
	if err != nil {
		publisher.Logger.Infof("Unable to unmarshal SetMessage in %s", address)
		return
	}
	// Verify that the message comes from the sender using the sender's public key
	isValid := publisher.VerifyMessageSignature(setMessage.Sender, publication.Message, publication.Signature)
	if !isValid {
		publisher.Logger.Warningf("Incoming message verification failed for sender: %s", setMessage.Sender)
		return
	}
	if publisher.onNodeInputHandler != nil {
		publisher.onNodeInputHandler(input, &setMessage)
	}
}

// Main heartbeat loop to publish, discove and poll value updates
func (publisher *PublisherState) heartbeatLoop() {
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
	publisher.Logger.Warningf("Ending loop of publisher %s", publisher.publisherID)
}

// handlePublisherDiscovery stores discovered (remote) publishers in the zone for their public key
// Used to verify signatures of incoming configuration and input messages
// address contains the publisher's discovery address: zone/publisher/$publisher/$node
// publication contains a message with the publisher node info
func (publisher *PublisherState) handlePublisherDiscovery(address string, publication *messenger.Publication) {
	var pubNode nodes.Node
	err := json.Unmarshal(publication.Message, &pubNode)
	if err != nil {
		publisher.Logger.Warningf("Unable to unmarshal Publisher Node in %s: %s", address, err)
		return
	}
	publisher.updateMutex.Lock()
	publisher.zonePublishers[address] = &pubNode
	publisher.updateMutex.Unlock()
	publisher.Logger.Infof("Discovered publisher %s", address)
}

// NewPublisher creates a publisher instance and node for use in publications
// zoneID for the zone this publisher lives in
// publisherID of this publisher, unique within the zone
// messenger for publishing onto the message bus
// onConfig method handles incoming configuration requests. Default is to update the config directly
// onInput method handles commands to control published inputs
func NewPublisher(
	zoneID string,
	publisherID string,
	messenger messenger.IMessenger,
) *PublisherState {

	var pubNode = nodes.NewNode(zoneID, publisherID, nodes.PublisherNodeID)

	// IotConnect core running state of the publisher
	var publisher = &PublisherState{
		discoveryInterval: DefaultDiscoveryInterval,
		exitChannel:       make(chan bool),
		Inputs:            nodes.NewInputList(),
		// Logger:            log.New(),
		messenger:      messenger,
		Nodes:          nodes.NewNodeList(),
		Outputs:        nodes.NewOutputList(),
		OutputHistory:  nodes.NewOutputHistoryList(),
		outputForecast: make(map[string]standard.HistoryList),
		// outputHistory:     make(map[string]standard.HistoryList),
		pollCountdown:  1, // run discovery before poll
		pollInterval:   DefaultPollInterval,
		publisherID:    publisherID,
		PublisherNode:  pubNode,
		updateMutex:    &sync.Mutex{},
		Zone:           zoneID,
		zonePublishers: make(map[string]*nodes.Node),
	}
	publisher.Logger = &logrus.Logger{
		Out:   os.Stderr,
		Level: logrus.DebugLevel,
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

	// generate private/public key for signing and store the public key in the publisher identity
	// TODO: store keys
	rng := rand.Reader
	curve := elliptic.P256()
	privKey, err := ecdsa.GenerateKey(curve, rng)
	publisher.signPrivateKey = privKey
	if err != nil {
		publisher.Logger.Errorf("Failed to create keys for signing: %s", err)
	}
	privStr, pubStr := standard.EncodeKeys(privKey, &privKey.PublicKey)
	_ = privStr

	timeStampStr := time.Now().Format("2006-01-02T15:04:05.000-0700")
	pubNode.Identity = &nodes.Identity{
		Address:          pubNode.Address,
		PublicKeySigning: pubStr,
		Publisher:        publisherID,
		Timestamp:        timeStampStr,
		Zone:             zoneID,
	}
	publisher.Nodes.UpdateNode(pubNode)
	return publisher
}
