// Package publisher ...
// - Publishes updates to node, inputs and outputs when they are (re)discovered
// - configuration of nodes
// - control of inputs
// - update of security keys and identity signature
// Thread-safe. The publisher can be invoked from multiple goroutines
package publisher

import (
	"myzone/messenger"
	"myzone/nodes"
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
)

// State carries the operating state of this publisher
type State struct {
	// configHandler     func(map[string]string) map[string]string // handle before applying configuration
	discoverCountdown int                    // countdown each heartbeat
	discoveryInterval int                    // discovery polling interval
	discoveryHandler  func(publisher *State) // function that performs discovery
	Logger            *log.Logger            //
	messenger         messenger.IMessenger   // Message bus messenger to use
	pollHandler       func(publisher *State) // function that performs value polling
	pollCountdown     int                    // countdown each heartbeat
	pollInterval      int                    // value polling interval
	publisherID       string                 // for easy access to the pub ID
	publisherNode     *nodes.Node            // This publisher's node
	synchroneous      bool                   // publish synchroneous with updates for testing
	zoneID            string                 // Easy access to zone ID

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

// DiscoverNode is invoked when a node is (re)discovered by this publisher
// The given node replaces the existing node if one exists
func (publisher *State) DiscoverNode(node *nodes.Node) {
	publisher.Logger.Info("Discovered node: ", node.Address)

	publisher.updateMutex.Lock()
	publisher.nodes[node.Address] = node
	if publisher.updatedNodes == nil {
		publisher.updatedNodes = make(map[string]*nodes.Node)
	}
	publisher.updatedNodes[node.Address] = node

	if publisher.synchroneous {
		publisher.publishUpdates()
	}
	publisher.updateMutex.Unlock()
}

// DiscoverInput is invoked when a node input is (re)discovered by this publisher
// The given input replaces the existing input if one exists
// If a node alias is set then the input and outputs are published under the alias instead of the node id
func (publisher *State) DiscoverInput(input *nodes.InOutput) {
	publisher.Logger.Info("Discovered input: ", input.Address)

	publisher.updateMutex.Lock()
	publisher.inputs[input.Address] = input
	if publisher.updatedInOutputs == nil {
		publisher.updatedInOutputs = make(map[string]*nodes.InOutput)
	}
	publisher.updatedInOutputs[input.Address] = input

	if publisher.synchroneous {
		publisher.publishUpdates()
	}
	publisher.updateMutex.Unlock()
}

// DiscoverOutput is invoked when a node output is (re)discovered by this publisher
// The given output replaces the existing output if one exists
func (publisher *State) DiscoverOutput(output *nodes.InOutput) {
	publisher.Logger.Info("Discovered output: ", output.Address)

	publisher.updateMutex.Lock()
	publisher.outputs[output.Address] = output
	if publisher.updatedInOutputs == nil {
		publisher.updatedInOutputs = make(map[string]*nodes.InOutput)
	}
	publisher.updatedInOutputs[output.Address] = output

	if publisher.synchroneous {
		publisher.publishUpdates()
	}
	publisher.updateMutex.Unlock()
}

// GetNode returns a discovered node by the node address
// address of the node, only the zone, publisher and nodeID are used. Any command suffix is ignored
// Returns nil if address has no known node
func (publisher *State) GetNode(address string) *nodes.Node {
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
func (publisher *State) GetInput(address string) *nodes.InOutput {
	publisher.updateMutex.Lock()
	var input = publisher.inputs[address]
	publisher.updateMutex.Unlock()
	return input
}

// GetOutput returns a discovered output by its discovery address
// Returns nil if address has no known output
func (publisher *State) GetOutput(address string) *nodes.InOutput {
	publisher.updateMutex.Lock()
	var output = publisher.outputs[address]
	publisher.updateMutex.Unlock()
	return output
}

// SetDiscoveryInterval is a convenience function for periodic update of discovered
// nodes, inputs and outputs. Intended for publishers that need to poll for discovery.
//
// interval in seconds to perform another discovery. Default is DefaultDiscoveryInterval
// handler is the callback with the publisher for publishing discovery
func (publisher *State) SetDiscoveryInterval(interval int, handler func(publisher *State)) {
	publisher.Logger.Infof("discovery interval = %d seconds", interval)
	publisher.discoveryInterval = interval
	publisher.discoveryHandler = handler
}

// SetPollingInterval is a convenience function for periodic update of output values
// interval in seconds to perform another poll. Default is DefaultPollInterval
// intended for publishers that need to poll for values
func (publisher *State) SetPollingInterval(interval int, handler func(publisher *State)) {
	publisher.Logger.Infof("polling interval = %d seconds", interval)
	publisher.pollInterval = interval
	publisher.pollHandler = handler
}

// Start publishing
// synchroneous publications for testing
func (publisher *State) Start(synchroneous bool) {
	publisher.synchroneous = synchroneous
	if !publisher.isRunning {
		publisher.Logger.Warningf("Starting publisher %s", publisher.publisherID)
		publisher.updateMutex.Lock()
		publisher.isRunning = true
		publisher.updateMutex.Unlock()
		go publisher.heartbeatLoop()
		// wait for the heartbeat to start
		<-publisher.exitChannel
		publisher.Logger.Warningf("Publisher %s started", publisher.publisherID)
	}
}

// Stop publishing
// Wait until the heartbeat loop has finished processing messages
func (publisher *State) Stop() {
	publisher.Logger.Warningf("Stopping publisher %s", publisher.publisherID)
	publisher.updateMutex.Lock()
	if publisher.isRunning {
		publisher.isRunning = false
		publisher.updateMutex.Unlock()
		// wait for heartbeat to end
		<-publisher.exitChannel
	} else {
		publisher.updateMutex.Unlock()
	}
	publisher.Logger.Info("... bye bye")
}

// UpdateNodeConfig updates a node's configuration
// Call this after the configuration has been processed by the publisher and
// only apply the configuration that take effect immediately. If the configuration
// has to be processed by a node then excluded it from the map and wait for the node's
// confirmation.
// address must start with the node address zone/publisher/node
func (publisher *State) UpdateNodeConfig(address string, param map[string]string) {
	node := publisher.GetNode(address)

	var appliedParams map[string]string = param
	for key, value := range appliedParams {
		config := node.Config[key]
		if config == nil {
			config = &nodes.ConfigAttr{}
			node.Config[key] = config
		}
		config.Value = value
	}
	// re-discover the node for publication
	publisher.DiscoverNode(node)
}

// UpdateOutputValue is invoked when an output value is updated
// Ignores the value if such output has not yet been discovered
func (publisher *State) UpdateOutputValue(address string, newValue string) {
	var output = publisher.GetOutput(address)
	if output != nil {
		nodes.UpdateValue(output, newValue)

		publisher.updateMutex.Lock()
		if publisher.updatedOutputValues == nil {
			publisher.updatedOutputValues = make(map[string]*nodes.InOutput)
		}
		publisher.updatedOutputValues[output.Address] = output

		if publisher.synchroneous {
			publisher.publishUpdates()
		}
		publisher.updateMutex.Unlock()
	}
}

// GetNode returns a discovered node by the node address
// address of the node, only the zone, publisher and nodeID are used. Any command suffix is ignored
// This method is not thread safe and should only be used in a locked section
// Returns nil if address has no known node
func (publisher *State) getNode(address string) *nodes.Node {
	segments := strings.Split(address, "/")
	segments[3] = "$node"
	nodeAddr := strings.Join(segments[:4], "/")

	var node = publisher.nodes[nodeAddr]
	return node
}

// Replace the address with the node's alias instead the node ID, if available
// return the address if the node doesn't have an alias
// This method is not thread safe and should only be used in a locked section
func (publisher *State) getAliasAddress(address string) string {
	node := publisher.getNode(address)
	if node == nil {
		return address
	}
	aliasConfig := node.Config[nodes.AttrNameAlias]
	if (aliasConfig == nil) || (aliasConfig.Value == "") {
		return address
	}
	parts := strings.Split(address, "/")
	parts[2] = aliasConfig.Value
	aliasAddr := strings.Join(parts, "/")
	return aliasAddr
}

// Main heartbeat loop to publish, discove and poll value updates
func (publisher *State) heartbeatLoop() {
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

// Publish discovery and value updates onto the message bus
// The order is nodes first, followed by in/outputs, followed by values.
//   the sequence within nodes, in/outputs, and values does not follow the discovery sequence
//   as the map used to record updates is unordered.
// This method is not thread safe and should only be used in a locked section
func (publisher *State) publishUpdates() {
	// publish changes to nodes
	if publisher.messenger == nil {
		return // can't do anything here, just go home
	}
	// publish updated nodes
	if publisher.updatedNodes != nil {
		for addr, node := range publisher.updatedNodes {
			publisher.Logger.Infof("publish node discovery: %s", addr)
			publisher.messenger.Publish(addr, node)
		}
		publisher.updatedNodes = nil
	}

	// publish updated inputs or outputs
	if publisher.updatedInOutputs != nil {
		for addr, inoutput := range publisher.updatedInOutputs {
			aliasAddress := publisher.getAliasAddress(addr)
			publisher.Logger.Infof("publish in/output discovery: %s", aliasAddress)
			publisher.messenger.Publish(aliasAddress, inoutput)
		}
		publisher.updatedInOutputs = nil
	}
	// publish updated output values using alias address if configured
	if publisher.updatedOutputValues != nil {
		for addr, output := range publisher.updatedOutputValues {
			aliasAddress := publisher.getAliasAddress(addr)
			publisher.Logger.Infof("publish output value: %s", aliasAddress)
			var latestValue = nodes.GetLatest(output)
			publisher.messenger.Publish(aliasAddress, latestValue.Value) // raw
		}
		publisher.updatedOutputValues = nil
	}

}

// NewPublisher creates a publisher instance and node for use in publications
// zoneID for the zone this publisher lives in
// publisherID of this publisher, unique within the zone
// messenger for publishing onto the message bus
func NewPublisher(zoneID string, publisherID string, messenger messenger.IMessenger) *State {

	var pubNode = nodes.NewNode(zoneID, publisherID, PublisherNodeID)

	// MyZone core running state of the publisher
	var publisher = &State{
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
