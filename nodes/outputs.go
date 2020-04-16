// Package nodes with handling of node outputs objects
package nodes

import (
	"fmt"
	"sync"

	"github.com/hspaay/iotconnect.golang/standard"
)

// OutputList with output management
type OutputList struct {
	outputMap      map[string]*standard.InOutput
	updateMutex    *sync.Mutex                   // mutex for async updating of outputs
	updatedOutputs map[string]*standard.InOutput // address of outputs that have been rediscovered/updated since last publication
}

// GetOutput returns the output of one of this publisher's nodes
// This method is concurrent safe
// Returns nil if address has no known output
func (outputs *OutputList) GetOutput(
	node *standard.Node, outputType string, instance string) *standard.InOutput {
	// segments := strings.Split(address, "/")
	// segments[3] = standard.CommandOutputDiscovery
	// outputAddr := strings.Join(segments, "/")
	outputAddr := fmt.Sprintf("%s/%s/%s/%s/%s/%s", node.Zone, node.PublisherID, node.ID,
		standard.CommandOutputDiscovery, outputType, instance)

	outputs.updateMutex.Lock()
	var output = outputs.outputMap[outputAddr]
	outputs.updateMutex.Unlock()
	return output
}

// GetNodeOutputs returns a list of all outputs for the given node
// This method is concurrent safe
func (outputs *OutputList) GetNodeOutputs(node *standard.Node) []*standard.InOutput {
	nodeOutputs := []*standard.InOutput{}
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
func (outputs *OutputList) GetOutputByAddress(outputAddr string) *standard.InOutput {
	outputs.updateMutex.Lock()
	var output = outputs.outputMap[outputAddr]
	outputs.updateMutex.Unlock()
	return output
}

// GetUpdatedOutputs returns the list of discovered outputs that have been updated
// clear the update on return
func (outputs *OutputList) GetUpdatedOutputs(clearUpdates bool) []*standard.InOutput {
	var updateList []*standard.InOutput = make([]*standard.InOutput, 0)

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
func (outputs *OutputList) UpdateOutput(output *standard.InOutput) {
	outputs.updateMutex.Lock()
	outputs.outputMap[output.Address] = output
	if outputs.updatedOutputs == nil {
		outputs.updatedOutputs = make(map[string]*standard.InOutput)
	}
	outputs.updatedOutputs[output.Address] = output
	outputs.updateMutex.Unlock()
}

// NewOutput creates a new output for the given node
// If an output of the same type and instance exists, it will be replaced
// node is the node that contains the output
// outputType is one of the predefined output types. See constants in the standard
// instance is the output instance in case of multiple instances of the same type. Use
func (outputs *OutputList) NewOutput(node *standard.Node, outputType string, instance string) *standard.InOutput {
	output := standard.NewOutput(node, outputType, instance)
	outputs.UpdateOutput(output)
	return output
}

// NewOutputList creates a new instance for output management
func NewOutputList() *OutputList {
	outputs := OutputList{
		outputMap:   make(map[string]*standard.InOutput),
		updateMutex: &sync.Mutex{},
	}
	return &outputs
}
