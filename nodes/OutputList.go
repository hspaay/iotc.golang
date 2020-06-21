// Package nodes with handling of node outputs objects
package nodes

import (
	"fmt"
	"strings"
	"sync"

	"github.com/hspaay/iotc.golang/iotc"
)

// OutputList with output management
type OutputList struct {
	outputMap      map[string]*iotc.OutputDiscoveryMessage // output discovery address - object map
	updateMutex    *sync.Mutex                             // mutex for async updating of outputs
	updatedOutputs map[string]*iotc.OutputDiscoveryMessage // address of outputs that have been rediscovered/updated since last publication
}

// GetAllOutputs returns the list of outputs
func (outputs *OutputList) GetAllOutputs() []*iotc.OutputDiscoveryMessage {
	outputs.updateMutex.Lock()
	defer outputs.updateMutex.Unlock()

	var outputList = make([]*iotc.OutputDiscoveryMessage, 0)
	for _, output := range outputs.outputMap {
		outputList = append(outputList, output)
	}
	return outputList
}

// GetOutput returns the output of one of this publisher's nodes
// This method is concurrent safe
// Returns nil if address has no known output
func (outputs *OutputList) GetOutput(
	nodeAddress string, outputType iotc.OutputType, instance string) *iotc.OutputDiscoveryMessage {

	outputAddr := MakeOutputDiscoveryAddress(nodeAddress, outputType, instance)

	outputs.updateMutex.Lock()
	var output = outputs.outputMap[outputAddr]
	outputs.updateMutex.Unlock()
	return output
}

// GetNodeOutputs returns all outputs for the given node in this list
// This method is concurrent safe
func (outputs *OutputList) GetNodeOutputs(node *iotc.NodeDiscoveryMessage) []*iotc.OutputDiscoveryMessage {
	nodeOutputs := []*iotc.OutputDiscoveryMessage{}
	segments := strings.Split(node.Address, "/")
	prefix := strings.Join(segments[:3], "/")

	for _, output := range outputs.outputMap {
		if strings.HasPrefix(output.Address, prefix) {
			nodeOutputs = append(nodeOutputs, output)
		}
	}
	return nodeOutputs
}

// GetOutputByAddress returns an output by its address
// outputAddr must contain the full output address, eg <zone>/<publisher>/<node>/"$output"/<type>/<instance>
// Returns nil if address has no known output
// This method is concurrent safe
func (outputs *OutputList) GetOutputByAddress(outputAddr string) *iotc.OutputDiscoveryMessage {
	outputs.updateMutex.Lock()
	var output = outputs.outputMap[outputAddr]
	outputs.updateMutex.Unlock()
	return output
}

// GetUpdatedOutputs returns the list of discovered outputs that have been updated
// clear the update on return
func (outputs *OutputList) GetUpdatedOutputs(clearUpdates bool) []*iotc.OutputDiscoveryMessage {
	var updateList []*iotc.OutputDiscoveryMessage = make([]*iotc.OutputDiscoveryMessage, 0)

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

// UpdateOutput replaces the output
// The output will be added to the list of updated outputs
func (outputs *OutputList) UpdateOutput(output *iotc.OutputDiscoveryMessage) {
	outputs.updateMutex.Lock()
	outputs.outputMap[output.Address] = output
	if outputs.updatedOutputs == nil {
		outputs.updatedOutputs = make(map[string]*iotc.OutputDiscoveryMessage)
	}
	outputs.updatedOutputs[output.Address] = output
	outputs.updateMutex.Unlock()
}

// MakeOutputDiscoveryAddress for publishing or subscribing
// nodeAddress is the address containing the node. Any node, input or output address will do
func MakeOutputDiscoveryAddress(nodeAddress string, outputType iotc.OutputType, instance string) string {
	segments := strings.Split(nodeAddress, "/")
	zone := segments[0]
	publisherID := segments[1]
	nodeID := segments[2]

	address := fmt.Sprintf("%s/%s/%s"+"/%s/%s/"+iotc.MessageTypeOutputDiscovery,
		zone, publisherID, nodeID, outputType, instance)
	return address
}

// NewOutput creates a new output for the given node.
// It is not immediately added to allow for further updates of the ouput definition.
// To add it to the list use 'UpdateOutput'
// node is the node that contains the output
// outputType is one of the predefined output types. See constants in the standard
// instance is the output instance in case of multiple instances of the same type. Use
func NewOutput(nodeAddress string, outputType iotc.OutputType, instance string) *iotc.OutputDiscoveryMessage {
	// output := NewOutput(node, outputType, instance)
	address := MakeOutputDiscoveryAddress(nodeAddress, outputType, instance)
	// segments := strings.Split(nodeAddress, "/")

	output := &iotc.OutputDiscoveryMessage{
		Address:    address,
		Instance:   instance,
		OutputType: outputType,
		// NodeID:     segments[2],
		// PublisherID: node.PublisherID,
		// History:  make([]*HistoryValue, 1),
	}
	return output
}

// NewOutputList creates a new instance for output management
func NewOutputList() *OutputList {
	outputs := OutputList{
		outputMap:   make(map[string]*iotc.OutputDiscoveryMessage),
		updateMutex: &sync.Mutex{},
	}
	return &outputs
}
