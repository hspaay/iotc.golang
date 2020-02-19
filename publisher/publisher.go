// Package publisher ...
// - Publishes updates to node, inputs and outputs when they are (re)discovered
// - configuration of nodes
// - control of inputs
// - update of security keys and identity signature
package publisher

import (
	"encoding/json"
	"myzone/messenger"
	node "myzone/node"
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
	discoverCountdown int                       // countdown each heartbeat
	discoveryInterval int                       // discovery polling interval
	discoveryHandler  func(publisher *State)    // function that performs discovery
	nodes             map[string]*node.Node     // nodes by discovery address
	inputs            map[string]*node.InOutput // inputs by discovery address
	isRunning         bool                      // publisher was started and is running
	logger            *log.Logger               //
	messenger         messenger.IMessenger      // Message bus messenger to use
	outputs           map[string]*node.InOutput // outputs by discovery address
	pollHandler       func(publisher *State)    // function that performs value polling
	pollCountdown     int                       // countdown each heartbeat
	pollInterval      int                       // value polling interval
	publisherID       string                    // for easy access to the pub ID
	publisherNode     *node.Node                // This publisher's node
	synchroneous      bool                      // publish synchroneous with updates for testing
	updatedNodes      map[string]*node.Node
	updatedInOutputs  map[string]*node.InOutput
	zoneID            string // Easy access to zone ID
}

// DiscoverNode is invoked when a node is (re)discovered by this publisher
// The given node replaces the existing node if one exists
func (publisher *State) DiscoverNode(discoNode *node.Node) {
	log.Info("publisher.DiscoverNode: Node Address=", discoNode.Address)
	publisher.nodes[discoNode.Address] = discoNode
	if publisher.updatedNodes == nil {
		publisher.updatedNodes = make(map[string]*node.Node)
	}
	publisher.updatedNodes[discoNode.Address] = discoNode
	if publisher.synchroneous {
		publisher.publishUpdates()
	}
}

// DiscoverInput is invoked when a node input is (re)discovered by this publisher
// The given input replaces the existing input if one exists
func (publisher *State) DiscoverInput(input *node.InOutput) {
	publisher.inputs[input.Address] = input
	if publisher.updatedInOutputs == nil {
		publisher.updatedInOutputs = make(map[string]*node.InOutput)
	}
	publisher.updatedInOutputs[input.Address] = input
	if publisher.synchroneous {
		publisher.publishUpdates()
	}
}

// DiscoverOutput is invoked when a node output is (re)discovered by this publisher
// The given output replaces the existing output if one exists
func (publisher *State) DiscoverOutput(output *node.InOutput) {
	publisher.outputs[output.Address] = output
	if publisher.updatedInOutputs == nil {
		publisher.updatedInOutputs = make(map[string]*node.InOutput)
	}
	publisher.updatedInOutputs[output.Address] = output
	if publisher.synchroneous {
		publisher.publishUpdates()
	}
}

// GetNode returns a discovered node by its discovery address
// Returns nil if address has no known node
func (publisher *State) GetNode(address string) *node.Node {
	var node = publisher.nodes[address]
	return node
}

// GetInput returns a discovered input by its discovery address
// Returns nil if address has no known input
func (publisher *State) GetInput(address string) *node.InOutput {
	var input = publisher.inputs[address]
	return input
}

// GetOutput returns a discovered output by its discovery address
// Returns nil if address has no known output
func (publisher *State) GetOutput(address string) *node.InOutput {
	var output = publisher.outputs[address]
	return output
}

// SetDiscoveryInterval is a convenience function for periodic update of discovered
// nodes, inputs and outputs. Intended for publishers that need to poll for discovery.
//
// interval in seconds to perform another discovery. Default is DefaultDiscoveryInterval
// handler is the callback with the publisher for publishing discovery
func (publisher *State) SetDiscoveryInterval(interval int, handler func(publisher *State)) {
	publisher.discoveryInterval = interval
	publisher.discoveryHandler = handler
}

// SetPollingInterval is a convenience function for periodic update of output values
// interval in seconds to perform another poll. Default is DefaultPollInterval
// intended for publishers that need to poll for values
func (publisher *State) SetPollingInterval(interval int, handler func(publisher *State)) {
	publisher.pollInterval = interval
	publisher.pollHandler = handler
}

// Start publishing
// synchroneous publications for testing
func (publisher *State) Start(synchroneous bool) {
	publisher.synchroneous = synchroneous
	if !publisher.isRunning {
		publisher.logger.Warningf("Publisher.Start: Started publisher %s", publisher.publisherID)
		publisher.isRunning = true
		go publisher.heartbeatLoop()
	}
}

// Stop publishing
func (publisher *State) Stop() {
	publisher.logger.Warningf("Publisher.Stop: Stopping publisher %s", publisher.publisherID)
	publisher.isRunning = false
}

// UpdateOutputValue is invoked when an output value is updated
// Ignores the value if such output is not known
func (publisher *State) UpdateOutputValue(address string, newValue string) {
	var output = publisher.GetOutput(address)
	if output != nil {
		node.UpdateValue(output, newValue)
	}
}

// Main heartbeat loop to publish, discove and poll value updates
func (publisher *State) heartbeatLoop() {
	for publisher.isRunning {
		time.Sleep(time.Second)

		// publish changes onto the bus (when synchroneous is set there won't be any)
		publisher.publishUpdates()

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

	}
	publisher.logger.Warningf("Publisher.heartbeatLoop: Ended loop of publisher %s", publisher.publisherID)
}

// publish updates onto the message bus
// use goroutines if synchroneous is false
func (publisher *State) publishUpdates() {
	// publish changes to nodes
	if publisher.messenger == nil {
		return // can't do anything here, just go home
	}
	if publisher.updatedNodes != nil {
		for addr, node := range publisher.updatedNodes {
			if publisher.synchroneous {
				publisher.messenger.Publish(addr, node)
			} else {
				go publisher.messenger.Publish(addr, node)
			}
		}
		publisher.updatedNodes = nil
	}

	// publish changes to inputs or outputs
	if publisher.updatedInOutputs != nil {
		for addr, inoutput := range publisher.updatedInOutputs {
			buffer, err := json.MarshalIndent(inoutput, " ", " ")
			if err != nil {
				publisher.logger.Errorf("convention.publishUpdates: Error marshalling in/output '"+inoutput.Address+"' to json:", err)
			} else {
				publisher.logger.Infof("Convention.publishUpdates: node '%s'", inoutput.Address)
				if publisher.synchroneous {
					publisher.messenger.Publish(addr, string(buffer))
				} else {
					go publisher.messenger.Publish(addr, string(buffer))
				}
			}
		}
		publisher.updatedInOutputs = nil
	}

}

// NewPublisher creates a publisher instance and node for use in publications
// zoneID for the zone this publisher lives in
// publisherID of this publisher, unique within the zone
// messenger for publishing onto the message bus
func NewPublisher(zoneID string, publisherID string, messenger messenger.IMessenger) *State {

	var pubNode = node.NewNode(zoneID, publisherID, PublisherNodeID)

	// MyZone core running state of the publisher
	var publisher = &State{
		discoveryInterval: DefaultDiscoveryInterval,
		inputs:            make(map[string]*node.InOutput, 0),
		logger:            log.New(),
		messenger:         messenger,
		nodes:             make(map[string]*node.Node),
		outputs:           make(map[string]*node.InOutput),
		pollInterval:      DefaultPollInterval,
		publisherID:       publisherID,
		publisherNode:     pubNode,
		zoneID:            zoneID,
	}
	publisher.DiscoverNode(pubNode)
	return publisher
}
