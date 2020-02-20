// Package publisher ...
// - Publishes updates to node, inputs and outputs when they are (re)discovered
// - configuration of nodes
// - control of inputs
// - update of security keys and identity signature
package publisher

import (
	"encoding/json"
	"myzone/messenger"
	"myzone/nodes"
	"strings"
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
	discoverCountdown int                        // countdown each heartbeat
	discoveryInterval int                        // discovery polling interval
	discoveryHandler  func(publisher *State)     // function that performs discovery
	nodes             map[string]*nodes.Node     // nodes by discovery address
	inputs            map[string]*nodes.InOutput // inputs by discovery address
	isRunning         bool                       // publisher was started and is running
	logger            *log.Logger                //
	messenger         messenger.IMessenger       // Message bus messenger to use
	outputs           map[string]*nodes.InOutput // outputs by discovery address
	pollHandler       func(publisher *State)     // function that performs value polling
	pollCountdown     int                        // countdown each heartbeat
	pollInterval      int                        // value polling interval
	publisherID       string                     // for easy access to the pub ID
	publisherNode     *nodes.Node                // This publisher's node
	synchroneous      bool                       // publish synchroneous with updates for testing
	updatedNodes      map[string]*nodes.Node
	updatedInOutputs  map[string]*nodes.InOutput
	zoneID            string // Easy access to zone ID
}

// DiscoverNode is invoked when a node is (re)discovered by this publisher
// The given node replaces the existing node if one exists
func (publisher *State) DiscoverNode(node *nodes.Node) {
	log.Info("publisher.DiscoverNode: Node Address=", node.Address)
	publisher.nodes[node.Address] = node
	if publisher.updatedNodes == nil {
		publisher.updatedNodes = make(map[string]*nodes.Node)
	}
	publisher.updatedNodes[node.Address] = node
	if publisher.synchroneous {
		publisher.publishUpdates()
	}
}

// DiscoverInput is invoked when a node input is (re)discovered by this publisher
// The given input replaces the existing input if one exists
// If a node alias is set then the input and outputs are published under the alias instead of the node id
func (publisher *State) DiscoverInput(input *nodes.InOutput) {
	publisher.inputs[input.Address] = input
	if publisher.updatedInOutputs == nil {
		publisher.updatedInOutputs = make(map[string]*nodes.InOutput)
	}
	publisher.updatedInOutputs[input.Address] = input
	if publisher.synchroneous {
		publisher.publishUpdates()
	}
}

// DiscoverOutput is invoked when a node output is (re)discovered by this publisher
// The given output replaces the existing output if one exists
func (publisher *State) DiscoverOutput(output *nodes.InOutput) {
	publisher.outputs[output.Address] = output
	if publisher.updatedInOutputs == nil {
		publisher.updatedInOutputs = make(map[string]*nodes.InOutput)
	}
	publisher.updatedInOutputs[output.Address] = output
	if publisher.synchroneous {
		publisher.publishUpdates()
	}
}

// GetNode returns a discovered node by the node address
// address of the node, only the zone, publisher and nodeID are used. Any command suffix is ignored
// Returns nil if address has no known node
func (publisher *State) GetNode(address string) *nodes.Node {
	segments := strings.Split(address, "/")
	segments[3] = "$node"
	nodeAddr := strings.Join(segments[:4], "/")
	var node = publisher.nodes[nodeAddr]
	return node
}

// GetInput returns a discovered input by its discovery address
// Returns nil if address has no known input
func (publisher *State) GetInput(address string) *nodes.InOutput {
	var input = publisher.inputs[address]
	return input
}

// GetOutput returns a discovered output by its discovery address
// Returns nil if address has no known output
func (publisher *State) GetOutput(address string) *nodes.InOutput {
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

// UpdateNodeConfig updates a node's configuration
// Call this after the configuration has been processed by the publisher and
// only apply the configuration that take effect immediately. If the configuration
// has to be processed by a node then excluded it from the map and wait for the node's
// confirmation.
func (publisher *State) UpdateNodeConfig(node *nodes.Node, param map[string]string) {
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
// Ignores the value if such output is not known
func (publisher *State) UpdateOutputValue(address string, newValue string) {
	var output = publisher.GetOutput(address)
	if output != nil {
		nodes.UpdateValue(output, newValue)
		// publish output value, latest, history, event, ...
	}
}

// Replace the address with the node's alias instead the node ID, if available
// return the address if the node doesn't have an alias
func (publisher *State) getAliasAddress(address string) string {
	node := publisher.GetNode(address)
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
	for publisher.isRunning {
		time.Sleep(time.Second)

		// Dont mess with pending changes during debugging
		if !publisher.synchroneous {
			publisher.publishUpdates()
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
			aliasAddress := publisher.getAliasAddress(addr)
			buffer, err := json.MarshalIndent(inoutput, " ", " ")
			if err != nil {
				publisher.logger.Errorf("convention.publishUpdates: Error marshalling in/output '"+aliasAddress+"' to json:", err)
			} else {
				publisher.logger.Infof("Convention.publishUpdates: in/output '%s'", aliasAddress)
				if publisher.synchroneous {
					publisher.messenger.Publish(aliasAddress, string(buffer))
				} else {
					go publisher.messenger.Publish(aliasAddress, string(buffer))
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

	var pubNode = nodes.NewNode(zoneID, publisherID, PublisherNodeID)

	// MyZone core running state of the publisher
	var publisher = &State{
		discoveryInterval: DefaultDiscoveryInterval,
		inputs:            make(map[string]*nodes.InOutput, 0),
		logger:            log.New(),
		messenger:         messenger,
		nodes:             make(map[string]*nodes.Node),
		outputs:           make(map[string]*nodes.InOutput),
		pollInterval:      DefaultPollInterval,
		publisherID:       publisherID,
		publisherNode:     pubNode,
		zoneID:            zoneID,
	}
	publisher.DiscoverNode(pubNode)
	return publisher
}
