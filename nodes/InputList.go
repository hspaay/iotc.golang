// Package nodes with handling of node inputs
package nodes

import (
	"fmt"
	"sync"

	"github.com/hspaay/iotc.golang/iotc"
)

// InputList with input management
type InputList struct {
	inputMap      map[string]*iotc.InputDiscoveryMessage
	updateMutex   *sync.Mutex                            // mutex for async updating of inputs
	updatedInputs map[string]*iotc.InputDiscoveryMessage // inputs that have been rediscovered/updated since last publication
}

// GetAllInputs returns the list of inputs
func (inputs *InputList) GetAllInputs() []*iotc.InputDiscoveryMessage {
	inputs.updateMutex.Lock()
	defer inputs.updateMutex.Unlock()

	var inputList = make([]*iotc.InputDiscoveryMessage, 0)
	for _, input := range inputs.inputMap {
		inputList = append(inputList, input)
	}
	return inputList
}

// GetInput returns the input of one of this publisher's nodes
// Returns nil if address has no known input
// address with node type and instance. The command will be ignored.
func (inputs *InputList) GetInput(
	node *iotc.NodeDiscoveryMessage, outputType string, instance string) *iotc.InputDiscoveryMessage {
	// segments := strings.Split(address, "/")
	// segments[3] = standard.CommandInputDiscovery
	// inputAddr := strings.Join(segments, "/")
	inputAddr := fmt.Sprintf("%s/%s/%s/%s/%s/%s", node.Zone, node.PublisherID, node.ID,
		iotc.MessageTypeInputDiscovery, outputType, instance)

	inputs.updateMutex.Lock()
	var input = inputs.inputMap[inputAddr]
	inputs.updateMutex.Unlock()
	return input
}

// GetInputByAddress returns an input by its address
// inputAddr must contain the full input address, eg <zone>/<publisher>/<node>/"$input"/<type>/<instance>
// Returns nil if address has no known input
// This method is concurrent safe
func (inputs *InputList) GetInputByAddress(inputAddr string) *iotc.InputDiscoveryMessage {
	inputs.updateMutex.Lock()
	var input = inputs.inputMap[inputAddr]
	inputs.updateMutex.Unlock()
	return input
}

// GetUpdatedInputs returns the list of discovered inputs that have been updated
// clear the update on return
func (inputs *InputList) GetUpdatedInputs(clearUpdates bool) []*iotc.InputDiscoveryMessage {
	var updateList []*iotc.InputDiscoveryMessage = make([]*iotc.InputDiscoveryMessage, 0)

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
func (inputs *InputList) UpdateInput(input *iotc.InputDiscoveryMessage) {
	inputs.updateMutex.Lock()
	inputs.inputMap[input.Address] = input
	if inputs.updatedInputs == nil {
		inputs.updatedInputs = make(map[string]*iotc.InputDiscoveryMessage)
	}
	inputs.updatedInputs[input.Address] = input
	inputs.updateMutex.Unlock()
}

// MakeInputDiscoveryAddress creates the address for the input discovery
func MakeInputDiscoveryAddress(zone string, publisherID string, nodeID string, inputType string, instance string) string {
	address := fmt.Sprintf("%s/%s/%s/"+iotc.MessageTypeInputDiscovery+"/%s/%s",
		zone, publisherID, nodeID, inputType, instance)
	return address
}

// MakeInputSetAddress creates the address used to update an input value
func MakeInputSetAddress(zone string, publisherID string, nodeID string, ioType string, instance string) string {
	address := fmt.Sprintf("%s/%s/%s/"+iotc.MessageTypeSet+"/%s/%s",
		zone, publisherID, nodeID, ioType, instance)
	return address
}

// NewInput instance
// To add it to the inputlist use 'UpdateInput'
func NewInput(node *iotc.NodeDiscoveryMessage, inputType string, instance string) *iotc.InputDiscoveryMessage {
	address := MakeInputDiscoveryAddress(node.Zone, node.PublisherID, node.ID, inputType, instance)
	input := &iotc.InputDiscoveryMessage{
		Address:   address,
		Instance:  instance,
		InputType: inputType,
	}
	return input
}

// NewInputList creates a new instance for input management
func NewInputList() *InputList {
	inputs := InputList{
		inputMap:    make(map[string]*iotc.InputDiscoveryMessage),
		updateMutex: &sync.Mutex{},
	}
	return &inputs
}
