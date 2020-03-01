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
	"iotconnect/messenger"
	"iotconnect/standard"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// reserved keywords
const (
	// DefaultDiscoveryInterval in which node discovery information is republished
	DefaultDiscoveryInterval = 24 * 3600
	// DefaultPollInterval in which the output values are queried for polling based sources
	DefaultPollInterval = 24 * 3600
)

// ThisPublisherState carries the operating state of 'this' publisher
type ThisPublisherState struct {
	discoverCountdown int                                                                 // countdown each heartbeat
	discoveryInterval int                                                                 // discovery polling interval
	discoveryHandler  func(publisher *ThisPublisherState)                                 // function that performs discovery
	Logger            *log.Logger                                                         //
	messenger         messenger.IMessenger                                                // Message bus messenger to use
	onConfig          func(node *standard.Node, config standard.AttrMap) standard.AttrMap // handle before applying configuration
	onSetMessage      func(input *standard.InOutput, message *standard.SetMessage)        // handle to update device/service input
	pollHandler       func(publisher *ThisPublisherState)                                 // function that performs value polling
	pollCountdown     int                                                                 // countdown each heartbeat
	pollInterval      int                                                                 // value polling interval
	publisherID       string                                                              // for easy access to the pub ID
	publisherNode     *standard.Node                                                      // This publisher's node
	zonePublishers    map[string]*standard.Node                                           // publishers on the network
	signPrivateKey    *ecdsa.PrivateKey                                                   // key for singing published messages
	synchroneous      bool                                                                // publish synchroneous with updates for testing
	zoneID            string                                                              // Easy access to zone ID

	// handle updates in the background or synchroneous. background publications require a mutex to prevent concurrent access
	exitChannel         chan bool
	updateMutex         *sync.Mutex                        // mutex for async updating and publishing
	configs             map[string]*standard.ConfigAttrMap // node configuration
	nodes               map[string]*standard.Node          // nodes by discovery address
	isRunning           bool                               // publisher was started and is running
	inputs              map[string]*standard.InOutput      // inputs by discovery address
	outputs             map[string]*standard.InOutput      // outputs by discovery address
	updatedNodes        map[string]*standard.Node          // nodes that have been rediscovered/updated since last publication
	updatedInOutputs    map[string]*standard.InOutput      // in/output that have been rediscovered/updated since last publication
	updatedOutputValues map[string]*standard.InOutput      // outputs whose values have updated since last publication
}

// GetNode returns a discovered node by the node address
// address of the node, only the zone, publisher and nodeID are used. Any command suffix is ignored
// Returns nil if address has no known node
func (publisher *ThisPublisherState) GetNode(address string) *standard.Node {
	segments := strings.Split(address, "/")
	segments[3] = standard.CommandNodeDiscovery
	nodeAddr := strings.Join(segments[:4], "/")

	publisher.updateMutex.Lock()
	var node = publisher.nodes[nodeAddr]
	publisher.updateMutex.Unlock()
	return node
}

// GetInput returns a discovered input by its discovery address
// Returns nil if address has no known input
// address with node type and instance. The command will be ignored.
func (publisher *ThisPublisherState) GetInput(address string) *standard.InOutput {
	segments := strings.Split(address, "/")
	segments[3] = standard.CommandInputDiscovery
	inputAddr := strings.Join(segments, "/")

	publisher.updateMutex.Lock()
	var input = publisher.inputs[inputAddr]
	publisher.updateMutex.Unlock()
	return input
}

// GetOutput returns a discovered output by its discovery address
// Returns nil if address has no known output
// address with node type and instance. The command will be ignored.
func (publisher *ThisPublisherState) GetOutput(address string) *standard.InOutput {
	segments := strings.Split(address, "/")
	segments[3] = standard.CommandOutputDiscovery
	outputAddr := strings.Join(segments, "/")
	publisher.updateMutex.Lock()
	var output = publisher.outputs[outputAddr]
	publisher.updateMutex.Unlock()
	return output
}

// Start publishing and listen for configuration and input messages
// synchroneous publications for testing
// onConfig handles updates to configuration, nil if no config to process
// onInput handles commands to update inputs, nil if there are no inputs to control
func (publisher *ThisPublisherState) Start(
	synchroneous bool,
	onConfig func(node *standard.Node, config standard.AttrMap) standard.AttrMap,
	onSetMessage func(input *standard.InOutput, message *standard.SetMessage)) {

	publisher.synchroneous = synchroneous
	publisher.onConfig = onConfig
	publisher.onSetMessage = onSetMessage
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
		configAddr := fmt.Sprintf("%s/%s/+/%s", publisher.zoneID, publisher.publisherID, standard.CommandConfigure)
		publisher.messenger.Subscribe(configAddr, publisher.handleNodeConfigCommand)

		inputAddr := fmt.Sprintf("%s/%s/+/%s/+/+", publisher.zoneID, publisher.publisherID, standard.CommandSet)
		publisher.messenger.Subscribe(inputAddr, publisher.handleNodeInput)

		// subscribe to publisher nodes to verify signature for input commands
		pubAddr := fmt.Sprintf("%s/+/%s/%s", publisher.zoneID, standard.PublisherNodeID, standard.CommandNodeDiscovery)
		publisher.messenger.Subscribe(pubAddr, publisher.handlePublisherDiscovery)

		publisher.Logger.Warningf("Publisher %s started", publisher.publisherID)
	}
}

// Stop publishing
// Wait until the heartbeat loop has finished processing messages
func (publisher *ThisPublisherState) Stop() {
	publisher.Logger.Warningf("Stopping publisher %s", publisher.publisherID)
	publisher.updateMutex.Lock()
	if publisher.isRunning {
		publisher.isRunning = false
		go messenger.NewDummyMessenger().Disconnect()
		publisher.updateMutex.Unlock()
		// wait for heartbeat to end
		<-publisher.exitChannel
	} else {
		publisher.updateMutex.Unlock()
	}
	publisher.Logger.Info("... bye bye")
}

// getNode returns a discovered node by the node address
// address of the node, only the zone, publisher and nodeID are used. Any command suffix is ignored
// This method is not thread safe and should only be used in a locked section
// Returns nil if address has no known node
func (publisher *ThisPublisherState) getNode(address string) *standard.Node {
	segments := strings.Split(address, "/")
	segments[3] = standard.CommandNodeDiscovery
	nodeAddr := strings.Join(segments[:4], "/")

	var node = publisher.nodes[nodeAddr]
	return node
}

// getNodeOutputs returns a list of outputs for the given node
// This method is not thread safe and should only be used in a locked section
func (publisher *ThisPublisherState) getNodeOutputs(node *standard.Node) []*standard.InOutput {
	outputs := []*standard.InOutput{}
	for _, output := range publisher.outputs {
		if output.NodeID == node.ID {
			outputs = append(outputs, output)
		}
	}
	return outputs
}

// Main heartbeat loop to publish, discove and poll value updates
func (publisher *ThisPublisherState) heartbeatLoop() {
	publisher.Logger.Warningf("starting heartbeat loop")
	publisher.exitChannel <- false

	for {
		time.Sleep(time.Second)

		// Dont mess with pending changes during debugging
		if !publisher.synchroneous {
			publisher.updateMutex.Lock()
			// FIXME: the publishUpdates duration adds to the heartbeat. This can also take a
			//  while unless the messenger unloads using channels (which it should)
			//  we want to be sure it has completed when the heartbeat ends
			publisher.publishDiscovery()
			publisher.publishOutputValues()
			publisher.updateMutex.Unlock()
		}

		// discover new nodes
		if (publisher.discoverCountdown <= 0) && (publisher.discoveryHandler != nil) {
			go publisher.discoveryHandler(publisher)
			publisher.discoverCountdown = publisher.discoveryInterval
		}
		publisher.discoverCountdown--

		// poll for values
		if (publisher.pollCountdown <= 0) && (publisher.pollHandler != nil) {
			go publisher.pollHandler(publisher)
			publisher.pollCountdown = publisher.pollInterval
		}
		publisher.discoverCountdown--

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
func (publisher *ThisPublisherState) handlePublisherDiscovery(address string, publication *messenger.Publication) {
	var pubNode standard.Node
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
) *ThisPublisherState {

	var pubNode = standard.NewNode(zoneID, publisherID, standard.PublisherNodeID)

	// IotZone core running state of the publisher
	var publisher = &ThisPublisherState{
		discoveryInterval: DefaultDiscoveryInterval,
		exitChannel:       make(chan bool),
		inputs:            make(map[string]*standard.InOutput, 0),
		Logger:            log.New(),
		messenger:         messenger,
		nodes:             make(map[string]*standard.Node),
		outputs:           make(map[string]*standard.InOutput),
		pollInterval:      DefaultPollInterval,
		publisherID:       publisherID,
		publisherNode:     pubNode,
		updateMutex:       &sync.Mutex{},
		zoneID:            zoneID,
		zonePublishers:    make(map[string]*standard.Node),
	}
	publisher.Logger.SetReportCaller(true) // publisher logging includes caller and file:line#

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
	pubNode.Identity = &standard.Identity{
		Address:          pubNode.Address,
		PublicKeySigning: pubStr,
		Publisher:        publisherID,
		Timestamp:        timeStampStr,
		Zone:             zoneID,
	}
	publisher.DiscoverNode(pubNode)
	return publisher
}
