// Package publisher with handling and publishing of discovered nodes, inputs and outputs
// (not to discovery of other nodes on the bus)
package publisher

import "github.com/hspaay/iotconnect.golang/standard"

// DiscoverNode is invoked when a node is (re)discovered by this publisher
// The given node replaces the existing node if one exists
func (publisher *PublisherState) DiscoverNode(node *standard.Node) *standard.Node {
	publisher.Logger.Info("Discovered node: ", node.Address)

	publisher.updateMutex.Lock()
	publisher.nodes[node.Address] = node
	if publisher.updatedNodes == nil {
		publisher.updatedNodes = make(map[string]*standard.Node)
	}
	publisher.updatedNodes[node.Address] = node

	if publisher.synchroneous {
		publisher.publishDiscovery()
	}
	publisher.updateMutex.Unlock()
	return node
}

// DiscoverInput is invoked when a publisher (re)discovered the input of one of its nodes
// The given input replaces the existing input if one exists
// If a node alias is set then the input and outputs are published under the alias instead of the node id
// Returns the actual input instance to use
func (publisher *PublisherState) DiscoverInput(input *standard.InOutput) *standard.InOutput {
	publisher.Logger.Info("Discovered input: ", input.Address)

	publisher.updateMutex.Lock()
	publisher.inputs[input.Address] = input
	if publisher.updatedInOutputs == nil {
		publisher.updatedInOutputs = make(map[string]*standard.InOutput)
	}
	publisher.updatedInOutputs[input.Address] = input

	if publisher.synchroneous {
		publisher.publishDiscovery()
	}
	publisher.updateMutex.Unlock()
	return input
}

// DiscoverOutput is invoked when a publishers (re)discovered an output of one of its nodes
// The given output replaces the existing output if one exists
// Returns the actual input instance to use
func (publisher *PublisherState) DiscoverOutput(output *standard.InOutput) *standard.InOutput {
	publisher.Logger.Info("Discovered output: ", output.Address)

	publisher.updateMutex.Lock()

	publisher.outputs[output.Address] = output
	if publisher.updatedInOutputs == nil {
		publisher.updatedInOutputs = make(map[string]*standard.InOutput)
	}
	publisher.updatedInOutputs[output.Address] = output

	if publisher.synchroneous {
		publisher.publishDiscovery()
	}
	publisher.updateMutex.Unlock()
	return output
}

// SetDiscoveryInterval is a convenience function for periodic update of discovered
// nodes, inputs and outputs. Intended for publishers that need to poll for discovery.
//
// interval in seconds to perform another discovery. Default is DefaultDiscoveryInterval
// handler is the callback with the publisher for publishing discovery
func (publisher *PublisherState) SetDiscoveryInterval(interval int, handler func(publisher *PublisherState)) {

	publisher.Logger.Infof("discovery interval = %d seconds", interval)
	publisher.discoveryInterval = interval
	publisher.discoveryHandler = handler
}

// SetPollingInterval is a convenience function for periodic update of output values
// interval in seconds to perform another poll. Default is DefaultPollInterval
// intended for publishers that need to poll for values
func (publisher *PublisherState) SetPollingInterval(interval int, handler func(publisher *PublisherState)) {
	publisher.Logger.Infof("polling interval = %d seconds", interval)
	publisher.pollInterval = interval
	publisher.pollHandler = handler
}

// publishDiscovery publishes updated nodes and in/outputs
func (publisher *PublisherState) publishDiscovery() {
	if publisher.messenger == nil {
		return // can't do anything here, just go home
	}
	// publish updated nodes
	if publisher.updatedNodes != nil {
		for addr, node := range publisher.updatedNodes {
			publisher.Logger.Infof("publish node discovery: %s", addr)
			publisher.publishMessage(addr, true, node)
		}
		publisher.updatedNodes = nil
	}

	// publish updated input or output discovery
	if publisher.updatedInOutputs != nil {
		for addr, inoutput := range publisher.updatedInOutputs {
			aliasAddress := publisher.getOutputAliasAddress(addr)
			publisher.Logger.Infof("publish in/output discovery: %s", aliasAddress)
			publisher.publishMessage(aliasAddress, true, inoutput)
		}
		publisher.updatedInOutputs = nil
	}
}
