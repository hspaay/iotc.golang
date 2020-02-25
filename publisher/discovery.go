// Package publisher with discovery methods
package publisher

import "iotzone/standard"

// DiscoverNode is invoked when a node is (re)discovered by this publisher
// The given node replaces the existing node if one exists
func (publisher *ThisPublisherState) DiscoverNode(node *standard.Node) {
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
}

// DiscoverInput is invoked when a node input is (re)discovered by this publisher
// The given input replaces the existing input if one exists
// If a node alias is set then the input and outputs are published under the alias instead of the node id
func (publisher *ThisPublisherState) DiscoverInput(input *standard.InOutput) {
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
}

// DiscoverOutput is invoked when a node output is (re)discovered by this publisher
// The given output replaces the existing output if one exists
func (publisher *ThisPublisherState) DiscoverOutput(output *standard.InOutput) {
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
}

// SetDiscoveryInterval is a convenience function for periodic update of discovered
// nodes, inputs and outputs. Intended for publishers that need to poll for discovery.
//
// interval in seconds to perform another discovery. Default is DefaultDiscoveryInterval
// handler is the callback with the publisher for publishing discovery
func (publisher *ThisPublisherState) SetDiscoveryInterval(interval int, handler func(publisher *ThisPublisherState)) {
	publisher.Logger.Infof("discovery interval = %d seconds", interval)
	publisher.discoveryInterval = interval
	publisher.discoveryHandler = handler
}

// SetPollingInterval is a convenience function for periodic update of output values
// interval in seconds to perform another poll. Default is DefaultPollInterval
// intended for publishers that need to poll for values
func (publisher *ThisPublisherState) SetPollingInterval(interval int, handler func(publisher *ThisPublisherState)) {
	publisher.Logger.Infof("polling interval = %d seconds", interval)
	publisher.pollInterval = interval
	publisher.pollHandler = handler
}

// UpdateOutputValue is invoked when an output value is updated
// Ignores the value if such output has not yet been discovered
func (publisher *ThisPublisherState) UpdateOutputValue(outputAddress string, newValue string) {
	var output = publisher.GetOutput(outputAddress)
	if output != nil {
		publisher.updateMutex.Lock()

		standard.UpdateValue(output, newValue)
		if publisher.updatedOutputValues == nil {
			publisher.updatedOutputValues = make(map[string]*standard.InOutput)
		}
		publisher.updatedOutputValues[output.Address] = output

		if publisher.synchroneous {
			publisher.publishDiscovery()
		}
		publisher.updateMutex.Unlock()
	} else {
		publisher.Logger.Warningf("Output to update not found. Address %s", outputAddress)
	}
}

// publishDiscovery publishes pending node and in/output discovery messages
func (publisher *ThisPublisherState) publishDiscovery() {
	if publisher.messenger == nil {
		return // can't do anything here, just go home
	}
	// publish updated nodes
	if publisher.updatedNodes != nil {
		for addr, node := range publisher.updatedNodes {
			publisher.Logger.Infof("publish node discovery: %s", addr)
			publisher.publishMessage(addr, node)
		}
		publisher.updatedNodes = nil
	}

	// publish updated input or output discovery
	if publisher.updatedInOutputs != nil {
		for addr, inoutput := range publisher.updatedInOutputs {
			aliasAddress := publisher.getAliasAddress(addr)
			publisher.Logger.Infof("publish in/output discovery: %s", aliasAddress)
			publisher.publishMessage(aliasAddress, inoutput)
		}
		publisher.updatedInOutputs = nil
	}
}
