// Package nodes with handling of node inputs
package nodes

import (
	"fmt"
	"sync"

	"github.com/hspaay/iotc.golang/messaging"
)

// InputList with input management
type InputList struct {
	inputMap      map[string]*Input
	updateMutex   *sync.Mutex       // mutex for async updating of inputs
	updatedInputs map[string]*Input // inputs that have been rediscovered/updated since last publication
}

// GetAllInputs returns the list of inputs
func (inputs *InputList) GetAllInputs() []*Input {
	inputs.updateMutex.Lock()
	defer inputs.updateMutex.Unlock()

	var inputList = make([]*Input, 0)
	for _, input := range inputs.inputMap {
		inputList = append(inputList, input)
	}
	return inputList
}

// GetInput returns the input of one of this publisher's nodes
// Returns nil if address has no known input
// address with node type and instance. The command will be ignored.
func (inputs *InputList) GetInput(
	node *Node, outputType string, instance string) *Input {
	// segments := strings.Split(address, "/")
	// segments[3] = standard.CommandInputDiscovery
	// inputAddr := strings.Join(segments, "/")
	inputAddr := fmt.Sprintf("%s/%s/%s/%s/%s/%s", node.Zone, node.PublisherID, node.ID,
		messaging.MessageTypeInputDiscovery, outputType, instance)

	inputs.updateMutex.Lock()
	var input = inputs.inputMap[inputAddr]
	inputs.updateMutex.Unlock()
	return input
}

// GetInputByAddress returns an input by its address
// inputAddr must contain the full input address, eg <zone>/<publisher>/<node>/"$input"/<type>/<instance>
// Returns nil if address has no known input
// This method is concurrent safe
func (inputs *InputList) GetInputByAddress(inputAddr string) *Input {
	inputs.updateMutex.Lock()
	var input = inputs.inputMap[inputAddr]
	inputs.updateMutex.Unlock()
	return input
}

// GetUpdatedInputs returns the list of discovered inputs that have been updated
// clear the update on return
func (inputs *InputList) GetUpdatedInputs(clearUpdates bool) []*Input {
	var updateList []*Input = make([]*Input, 0)

	inputs.updateMutex.Lock()
	if inputs.updatedInputs != nil {
		for _, output := range inputs.updatedInputs {
			updateList = append(updateList, output)
		}
		if clearUpdates {
			inputs.updatedInputs = nil
		}
	}
	inputs.updateMutex.Unlock()
	return updateList
}

// UpdateInput replaces the input using the node.Address
// This method is concurrent safe
func (inputs *InputList) UpdateInput(input *Input) {
	inputs.updateMutex.Lock()
	inputs.inputMap[input.Address] = input
	if inputs.updatedInputs == nil {
		inputs.updatedInputs = make(map[string]*Input)
	}
	inputs.updatedInputs[input.Address] = input
	inputs.updateMutex.Unlock()
}

// NewInputList creates a new instance for input management
func NewInputList() *InputList {
	inputs := InputList{
		inputMap:    make(map[string]*Input),
		updateMutex: &sync.Mutex{},
	}
	return &inputs
}
