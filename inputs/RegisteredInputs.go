// Package inputs with managing and publishing of registered inputs
package inputs

import (
	"sync"
	"time"

	"github.com/iotdomain/iotdomain-go/types"
)

// RegisteredInputs manages registration of publisher inputs
// Generics would be nice as this overlaps with outputs, nodes, publishers
type RegisteredInputs struct {
	domain        string                                  // the domain of this publisher
	publisherID   string                                  // the registered publisher for the inputs
	inputMap      map[string]*types.InputDiscoveryMessage // registered inputs by address
	updatedInputs map[string]*types.InputDiscoveryMessage // inputs that have been rediscovered/updated since last publication
	updateMutex   *sync.Mutex                             // mutex for async handling of inputs
}

// GetAllInputs returns the list of inputs
func (regInputs *RegisteredInputs) GetAllInputs() []*types.InputDiscoveryMessage {
	regInputs.updateMutex.Lock()
	defer regInputs.updateMutex.Unlock()

	var inputList = make([]*types.InputDiscoveryMessage, 0)
	for _, output := range regInputs.inputMap {
		inputList = append(inputList, output)
	}
	return inputList
}

// GetInput returns the input of one of this publisher's nodes
// Returns nil if address has no known input
// address with node type and instance. The command will be ignored.
func (regInputs *RegisteredInputs) GetInput(
	nodeID string, inputType types.InputType, instance string) *types.InputDiscoveryMessage {
	inputAddr := MakeInputDiscoveryAddress(regInputs.domain, regInputs.publisherID, nodeID, inputType, instance)

	regInputs.updateMutex.Lock()
	defer regInputs.updateMutex.Unlock()
	var input = regInputs.inputMap[inputAddr]
	return input
}

// GetInputByAddress returns an input by its address
// inputAddr must contain the full input address, eg <zone>/<publisher>/<node>/"$input"/<type>/<instance>
// Returns nil if address has no known input
// This method is concurrent safe
func (regInputs *RegisteredInputs) GetInputByAddress(inputAddr string) *types.InputDiscoveryMessage {
	regInputs.updateMutex.Lock()
	defer regInputs.updateMutex.Unlock()
	var input = regInputs.inputMap[inputAddr]
	return input
}

// GetInputsByNode returns a list of all inputs of a given node
func (regInputs *RegisteredInputs) GetInputsByNode(nodeID string) []*types.InputDiscoveryMessage {
	inputList := make([]*types.InputDiscoveryMessage, 0)
	regInputs.updateMutex.Lock()
	defer regInputs.updateMutex.Unlock()
	for _, input := range regInputs.inputMap {
		if input.NodeID == nodeID {
			inputList = append(inputList, input)
		}
	}
	return inputList
}

// GetUpdatedInputs returns the list of registered inputs that have been updated
// clear the update on return
func (regInputs *RegisteredInputs) GetUpdatedInputs(clearUpdates bool) []*types.InputDiscoveryMessage {
	var updateList []*types.InputDiscoveryMessage = make([]*types.InputDiscoveryMessage, 0)

	regInputs.updateMutex.Lock()
	defer regInputs.updateMutex.Unlock()
	if regInputs.updatedInputs != nil {
		for _, input := range regInputs.updatedInputs {
			updateList = append(updateList, input)
		}
		if clearUpdates {
			regInputs.updatedInputs = nil
		}
	}
	return updateList
}

// NewInput creates and registers a new input
func (regInputs *RegisteredInputs) NewInput(
	nodeID string, inputType types.InputType, instance string) *types.InputDiscoveryMessage {

	input := NewInput(regInputs.domain, regInputs.publisherID, nodeID, inputType, instance)
	regInputs.UpdateInput(input)
	return input
}

// SetAlias updates the address of all outputs with the given nodeID using the alias instead
func (regInputs *RegisteredInputs) SetAlias(nodeID string, alias string) {
	inputs := regInputs.GetInputsByNode(nodeID)
	for _, input := range inputs {
		oldAddress := input.Address
		newAddress := MakeInputDiscoveryAddress(
			regInputs.domain, regInputs.publisherID, alias, input.InputType, input.Instance)
		input.Address = newAddress
		input.NodeID = alias
		regInputs.updateMutex.Lock()
		delete(regInputs.inputMap, oldAddress)
		regInputs.updateMutex.Unlock()
		regInputs.UpdateInput(input)
	}
}

// UpdateInput replaces the registered input using the node.Address
func (regInputs *RegisteredInputs) UpdateInput(input *types.InputDiscoveryMessage) {

	regInputs.updateMutex.Lock()
	defer regInputs.updateMutex.Unlock()
	regInputs.inputMap[input.Address] = input
	if regInputs.updatedInputs == nil {
		regInputs.updatedInputs = make(map[string]*types.InputDiscoveryMessage)
	}
	input.Timestamp = time.Now().Format(types.TimeFormat)
	regInputs.updatedInputs[input.Address] = input
}

// NewInput instance for creating an input object for later adding.
// To add it to the inputlist use 'UpdateInput'
func NewInput(
	domain string, publisherID string, nodeID string, inputType types.InputType, instance string) *types.InputDiscoveryMessage {

	address := MakeInputDiscoveryAddress(domain, publisherID, nodeID, inputType, instance)
	// segments := strings.Split(nodeAddress, "/")
	input := &types.InputDiscoveryMessage{
		Address:     address,
		Attr:        make(types.NodeAttrMap),
		Config:      make(types.ConfigAttrMap),
		PublisherID: publisherID,
		NodeID:      nodeID,
		Instance:    instance,
		InputType:   inputType,
		Timestamp:   time.Now().Format(types.TimeFormat),
	}
	return input
}

// NewRegisteredInputs creates a new instance for registered input management
func NewRegisteredInputs(domain string, publisherID string) *RegisteredInputs {

	regInputs := &RegisteredInputs{
		domain:      domain,
		publisherID: publisherID,
		inputMap:    make(map[string]*types.InputDiscoveryMessage),
		updateMutex: &sync.Mutex{},
	}
	return regInputs
}
