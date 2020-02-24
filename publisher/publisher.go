// Package publisher ...
// - Publishes updates to node, inputs and outputs when they are (re)discovered
// - configuration of nodes
// - control of inputs
// - update of security keys and identity signature
// Thread-safe. All public functions can be invoked from multiple goroutines
package publisher

import (
	"fmt"
	"iotzone/messenger"
	"iotzone/nodes"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// reserved keywords
const (
	// LocalZone ID for local-only zones (eg, no sharing outside this zone)
	LocalZoneID = "$local"
	// PublisherNodeID to use when none is provided
	PublisherNodeID = "$publisher"
	// DefaultDiscoveryInterval in which node discovery information is republished
	DefaultDiscoveryInterval = 24 * 3600
	// DefaultPollInterval in which the output values are queried for polling based sources
	DefaultPollInterval = 24 * 3600
	//
	ConfigureCommand = "$configure"
	EventCommand     = "$event"
	HistoryCommand   = "$history"
	LatestCommand    = "$latest"
	SetCommand       = "$set"
	ValueCommand     = "$value"
)

// ThisPublisherState carries the operating state of 'this' publisher
type ThisPublisherState struct {
	discoverCountdown int                                                        // countdown each heartbeat
	discoveryInterval int                                                        // discovery polling interval
	discoveryHandler  func(publisher *ThisPublisherState)                        // function that performs discovery
	Logger            *log.Logger                                                //
	messenger         messenger.IMessenger                                       // Message bus messenger to use
	onConfig          func(node *nodes.Node, config nodes.AttrMap) nodes.AttrMap // handle before applying configuration
	onInput           func(input *nodes.InOutput, message string)                // handle to update device/service input
	pollHandler       func(publisher *ThisPublisherState)                        // function that performs value polling
	pollCountdown     int                                                        // countdown each heartbeat
	pollInterval      int                                                        // value polling interval
	publisherID       string                                                     // for easy access to the pub ID
	publisherNode     *nodes.Node                                                // This publisher's node
	synchroneous      bool                                                       // publish synchroneous with updates for testing
	zoneID            string                                                     // Easy access to zone ID

	// handle updates in the background or synchroneous. background publications require a mutex to prevent concurrent access
	exitChannel         chan bool
	updateMutex         *sync.Mutex                     // mutex for async updating and publishing
	configs             map[string]*nodes.ConfigAttrMap // node configuration
	nodes               map[string]*nodes.Node          // nodes by discovery address
	isRunning           bool                            // publisher was started and is running
	inputs              map[string]*nodes.InOutput      // inputs by discovery address
	outputs             map[string]*nodes.InOutput      // outputs by discovery address
	updatedNodes        map[string]*nodes.Node          // nodes that have been rediscovered/updated since last publication
	updatedInOutputs    map[string]*nodes.InOutput      // in/output that have been rediscovered/updated since last publication
	updatedOutputValues map[string]*nodes.InOutput      // outputs whose values have updated since last publication
}

// GetNode returns a discovered node by the node address
// address of the node, only the zone, publisher and nodeID are used. Any command suffix is ignored
// Returns nil if address has no known node
func (publisher *ThisPublisherState) GetNode(address string) *nodes.Node {
	segments := strings.Split(address, "/")
	segments[3] = "$node"
	nodeAddr := strings.Join(segments[:4], "/")

	publisher.updateMutex.Lock()
	var node = publisher.nodes[nodeAddr]
	publisher.updateMutex.Unlock()
	return node
}

// GetInput returns a discovered input by its discovery address
// Returns nil if address has no known input
func (publisher *ThisPublisherState) GetInput(address string) *nodes.InOutput {
	publisher.updateMutex.Lock()
	var input = publisher.inputs[address]
	publisher.updateMutex.Unlock()
	return input
}

// GetOutput returns a discovered output by its discovery address
// Returns nil if address has no known output
func (publisher *ThisPublisherState) GetOutput(address string) *nodes.InOutput {
	publisher.updateMutex.Lock()
	var output = publisher.outputs[address]
	publisher.updateMutex.Unlock()
	return output
}

// Start publishing and listen for configuration and input messages
// synchroneous publications for testing
// onConfig handles updates to configuration, nil if no config to process
// onInput handles commands to update inputs, nil if there are no inputs to control
func (publisher *ThisPublisherState) Start(
	synchroneous bool,
	onConfig func(node *nodes.Node, config nodes.AttrMap) nodes.AttrMap,
	onInput func(input *nodes.InOutput, message string)) {

	publisher.synchroneous = synchroneous
	publisher.onConfig = onConfig
	publisher.onInput = onInput
	if !publisher.isRunning {
		publisher.Logger.Warningf("Starting publisher %s", publisher.publisherID)
		publisher.updateMutex.Lock()
		publisher.isRunning = true
		publisher.updateMutex.Unlock()
		go publisher.heartbeatLoop()
		// wait for the heartbeat to start
		<-publisher.exitChannel

		// TODO: support LWT
		messenger.NewDummyMessenger().Connect("", "")
		// handle configuration and set messages
		configAddr := fmt.Sprintf("%s/%s/+/%s", publisher.zoneID, publisher.publisherID, ConfigureCommand)
		messenger.NewDummyMessenger().Subscribe(configAddr, publisher.handleNodeConfig)
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
func (publisher *ThisPublisherState) getNode(address string) *nodes.Node {
	segments := strings.Split(address, "/")
	segments[3] = "$node"
	nodeAddr := strings.Join(segments[:4], "/")

	var node = publisher.nodes[nodeAddr]
	return node
}

// getNodeOutputs returns a list of outputs for the given node
// This method is not thread safe and should only be used in a locked section
func (publisher *ThisPublisherState) getNodeOutputs(node *nodes.Node) []*nodes.InOutput {
	outputs := []*nodes.InOutput{}
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
			publisher.publishUpdates()
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

	var pubNode = nodes.NewNode(zoneID, publisherID, PublisherNodeID)

	// IotZone core running state of the publisher
	var publisher = &ThisPublisherState{
		discoveryInterval: DefaultDiscoveryInterval,
		exitChannel:       make(chan bool),
		inputs:            make(map[string]*nodes.InOutput, 0),
		Logger:            log.New(),
		messenger:         messenger,
		nodes:             make(map[string]*nodes.Node),
		outputs:           make(map[string]*nodes.InOutput),
		pollInterval:      DefaultPollInterval,
		publisherID:       publisherID,
		publisherNode:     pubNode,
		updateMutex:       &sync.Mutex{},
		zoneID:            zoneID,
	}
	publisher.Logger.SetReportCaller(true) // publisher logging includes caller and file:line#
	publisher.DiscoverNode(pubNode)
	return publisher
}
