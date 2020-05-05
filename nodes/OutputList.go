// Package nodes with handling of node outputs objects
package nodes

import (
	"fmt"
	"sync"

	"github.com/hspaay/iotconnect.golang/messaging"
)

// OutputList with output management
type OutputList struct {
	outputMap      map[string]*Output
	updateMutex    *sync.Mutex        // mutex for async updating of outputs
	updatedOutputs map[string]*Output // address of outputs that have been rediscovered/updated since last publication
}

// GetAllOutputs returns the list of outputs
func (outputs *OutputList) GetAllOutputs() []*Output {
	outputs.updateMutex.Lock()
	defer outputs.updateMutex.Unlock()

	var outputList = make([]*Output, 0)
	for _, output := range outputs.outputMap {
		outputList = append(outputList, output)
	}
	return outputList
}

// GetOutput returns the output of one of this publisher's nodes
// This method is concurrent safe
// Returns nil if address has no known output
func (outputs *OutputList) GetOutput(
	node *Node, outputType string, instance string) *Output {
	// segments := strings.Split(address, "/")
	// segments[3] = standard.CommandOutputDiscovery
	// outputAddr := strings.Join(segments, "/")
	outputAddr := fmt.Sprintf("%s/%s/%s/%s/%s/%s", node.Zone, node.PublisherID, node.ID,
		messaging.MessageTypeOutputDiscovery, outputType, instance)

	outputs.updateMutex.Lock()
	var output = outputs.outputMap[outputAddr]
	outputs.updateMutex.Unlock()
	return output
}

// GetNodeOutputs returns a list of all outputs for the given node
// This method is concurrent safe
func (outputs *OutputList) GetNodeOutputs(node *Node) []*Output {
	nodeOutputs := []*Output{}
	for _, output := range outputs.outputMap {
		if output.NodeID == node.ID {
			nodeOutputs = append(nodeOutputs, output)
		}
	}
	return nodeOutputs
}

// GetOutputByAddress returns an output by its address
// outputAddr must contain the full output address, eg <zone>/<publisher>/<node>/"$output"/<type>/<instance>
// Returns nil if address has no known output
// This method is concurrent safe
func (outputs *OutputList) GetOutputByAddress(outputAddr string) *Output {
	outputs.updateMutex.Lock()
	var output = outputs.outputMap[outputAddr]
	outputs.updateMutex.Unlock()
	return output
}

// GetUpdatedOutputs returns the list of discovered outputs that have been updated
// clear the update on return
func (outputs *OutputList) GetUpdatedOutputs(clearUpdates bool) []*Output {
	var updateList []*Output = make([]*Output, 0)

	outputs.updateMutex.Lock()
	if outputs.updatedOutputs != nil {
		for _, output := range outputs.updatedOutputs {
			updateList = append(updateList, output)
		}
		if clearUpdates {
			outputs.updatedOutputs = nil
		}
	}
	outputs.updateMutex.Unlock()
	return updateList
}

// UpdateOutput replaces the output using the node.Address
// This method is concurrent safe
func (outputs *OutputList) UpdateOutput(output *Output) {
	outputs.updateMutex.Lock()
	outputs.outputMap[output.Address] = output
	if outputs.updatedOutputs == nil {
		outputs.updatedOutputs = make(map[string]*Output)
	}
	outputs.updatedOutputs[output.Address] = output
	outputs.updateMutex.Unlock()
}

// NewOutput creates a new output for the given node and adds it to the output list
// If an output of the same type and instance exists, it will be replaced
// node is the node that contains the output
// outputType is one of the predefined output types. See constants in the standard
// instance is the output instance in case of multiple instances of the same type. Use
func (outputs *OutputList) NewOutput(node *Node, outputType string, instance string) *Output {
	output := NewOutput(node, outputType, instance)
	outputs.UpdateOutput(output)
	return output
}

// NewOutputList creates a new instance for output management
func NewOutputList() *OutputList {
	outputs := OutputList{
		outputMap:   make(map[string]*Output),
		updateMutex: &sync.Mutex{},
	}
	return &outputs
}
