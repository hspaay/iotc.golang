// Package nodes with handling of node inputs
package nodes

import (
	"fmt"
	"strings"
	"sync"

	"github.com/iotdomain/iotdomain-go/types"
)

// InputList with input management
type InputList struct {
	inputMap      map[string]*types.InputDiscoveryMessage
	updateMutex   *sync.Mutex                             // mutex for async updating of inputs
	updatedInputs map[string]*types.InputDiscoveryMessage // inputs that have been rediscovered/updated since last publication
}

// GetAllInputs returns the list of inputs
func (inputs *InputList) GetAllInputs() []*types.InputDiscoveryMessage {
	inputs.updateMutex.Lock()
	defer inputs.updateMutex.Unlock()

	var inputList = make([]*types.InputDiscoveryMessage, 0)
	for _, input := range inputs.inputMap {
		inputList = append(inputList, input)
	}
	return inputList
}

// GetInput returns the input of one of this publisher's nodes
// Returns nil if address has no known input
// address with node type and instance. The command will be ignored.
func (inputs *InputList) GetInput(
	nodeAddress string, inputType types.InputType, instance string) *types.InputDiscoveryMessage {
	// segments := strings.Split(address, "/")
	// segments[3] = standard.CommandInputDiscovery
	// inputAddr := strings.Join(segments, "/")
	inputAddr := MakeInputDiscoveryAddress(nodeAddress, inputType, instance)

	inputs.updateMutex.Lock()
	var input = inputs.inputMap[inputAddr]
	inputs.updateMutex.Unlock()
	return input
}

// GetInputByAddress returns an input by its address
// inputAddr must contain the full input address, eg <zone>/<publisher>/<node>/"$input"/<type>/<instance>
// Returns nil if address has no known input
// This method is concurrent safe
func (inputs *InputList) GetInputByAddress(inputAddr string) *types.InputDiscoveryMessage {
	inputs.updateMutex.Lock()
	var input = inputs.inputMap[inputAddr]
	inputs.updateMutex.Unlock()
	return input
}

// GetUpdatedInputs returns the list of discovered inputs that have been updated
// clear the update on return
func (inputs *InputList) GetUpdatedInputs(clearUpdates bool) []*types.InputDiscoveryMessage {
	var updateList []*types.InputDiscoveryMessage = make([]*types.InputDiscoveryMessage, 0)

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
func (inputs *InputList) UpdateInput(input *types.InputDiscoveryMessage) {
	inputs.updateMutex.Lock()
	inputs.inputMap[input.Address] = input
	if inputs.updatedInputs == nil {
		inputs.updatedInputs = make(map[string]*types.InputDiscoveryMessage)
	}
	inputs.updatedInputs[input.Address] = input
	inputs.updateMutex.Unlock()
}

// MakeInputDiscoveryAddress creates the address for the input discovery
func MakeInputDiscoveryAddress(nodeAddress string, inputType types.InputType, instance string) string {
	segments := strings.Split(nodeAddress, "/")
	zone := segments[0]
	publisherID := segments[1]
	nodeID := segments[2]

	address := fmt.Sprintf("%s/%s/%s"+"/%s/%s/"+types.MessageTypeInputDiscovery,
		zone, publisherID, nodeID, inputType, instance)
	return address
}

// MakeInputSetAddress creates the address used to update a node input value
// nodeAddress is an address containing the node.
func MakeInputSetAddress(nodeAddress string, ioType string, instance string) string {
	segments := strings.Split(nodeAddress, "/")
	zone := segments[0]
	publisherID := segments[1]
	nodeID := segments[2]

	address := fmt.Sprintf("%s/%s/%s"+"/%s/%s/"+types.MessageTypeSet,
		zone, publisherID, nodeID, ioType, instance)
	return address
}

// NewInput instance
// To add it to the inputlist use 'UpdateInput'
func NewInput(nodeAddr string, inputType types.InputType, instance string) *types.InputDiscoveryMessage {
	address := MakeInputDiscoveryAddress(nodeAddr, inputType, instance)
	// segments := strings.Split(nodeAddress, "/")
	input := &types.InputDiscoveryMessage{
		Address:   address,
		Instance:  instance,
		InputType: inputType,
		// NodeID:     segments[2],
	}
	return input
}

// NewInputList creates a new instance for input management
func NewInputList() *InputList {
	inputs := InputList{
		inputMap:    make(map[string]*types.InputDiscoveryMessage),
		updateMutex: &sync.Mutex{},
	}
	return &inputs
}
