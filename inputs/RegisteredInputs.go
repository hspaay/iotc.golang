// Package inputs with managing and publishing of registered inputs
package inputs

import (
	"errors"
	"sync"
	"time"

	"github.com/iotdomain/iotdomain-go/types"
)

// InputSubscription with handler of subscriber to input updates
type InputSubscription struct {
	handler      func(inputAddress string, sender string, value string) // the notification handler of this input
	inputAddress string                                                 // the input belonging to this subscription
	source       string                                                 // the source of the input, eg, the input address, http address or file path
}

// RegisteredInputs manages registration of publisher inputs
// Generics would be nice as this overlaps with outputs, nodes, publishers
type RegisteredInputs struct {
	domain        string                                  // the domain of this publisher
	publisherID   string                                  // the registered publisher for the inputs
	inputMap      map[string]*types.InputDiscoveryMessage // registered inputs by address
	updatedInputs map[string]*types.InputDiscoveryMessage // inputs that have been rediscovered/updated since last publication
	updateMutex   *sync.Mutex                             // mutex for async handling of inputs
	subscriptions map[string]*InputSubscription           // subscription to input by inputAddress
}

// CreateInput creates and registers a new input with optional handler for input trigger
func (regInputs *RegisteredInputs) CreateInput(
	nodeID string, inputType types.InputType, instance string,
	handler func(inputAddress string, sender string, value string)) *types.InputDiscoveryMessage {

	return regInputs.CreateInputWithSource(nodeID, inputType, instance, "", handler)
}

// CreateInputWithSource creates and registers a new input that takes its input value from a given source
// Replaces the existing input if it already exist.
func (regInputs *RegisteredInputs) CreateInputWithSource(
	nodeID string, inputType types.InputType, instance string, source string,
	handler func(inputAddress string, sender string, value string)) *types.InputDiscoveryMessage {

	regInputs.updateMutex.Lock()
	defer regInputs.updateMutex.Unlock()

	input := NewInput(regInputs.domain, regInputs.publisherID, nodeID, inputType, instance)
	input.Source = source

	regInputs.updateInput(input)

	sub := &InputSubscription{
		handler:      handler,
		inputAddress: input.Address,
		source:       input.Address,
	}
	regInputs.subscriptions[input.Address] = sub
	return input
}

// DeleteInput unregisters the input
func (regInputs *RegisteredInputs) DeleteInput(
	nodeID string, inputType types.InputType, instance string) {

	regInputs.updateMutex.Lock()
	defer regInputs.updateMutex.Unlock()

	inputAddr := MakeInputDiscoveryAddress(regInputs.domain, regInputs.publisherID, nodeID, inputType, instance)
	delete(regInputs.inputMap, inputAddr)
	delete(regInputs.subscriptions, inputAddr)
	if regInputs.updatedInputs == nil {
		regInputs.updatedInputs = make(map[string]*types.InputDiscoveryMessage)
	}
	// nil updates mean that the inputs are deleted
	regInputs.updatedInputs[inputAddr] = nil
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

// GetInputsWithSource returns a list of inputs that have the given source
func (regInputs *RegisteredInputs) GetInputsWithSource(source string) []*types.InputDiscoveryMessage {
	inputList := make([]*types.InputDiscoveryMessage, 0)
	for _, input := range regInputs.inputMap {
		if input.Source == source {
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

// SetAlias updates the address of all inputs with the given nodeID using the alias instead
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
		regInputs.updateInput(input)
		regInputs.updateMutex.Unlock()
	}
}

// UpdateInput replaces an existing input with the provided input.
// The input must already exist and be created using 'CreateInput', otherwise it returns an error
func (regInputs *RegisteredInputs) UpdateInput(input *types.InputDiscoveryMessage) error {

	regInputs.updateMutex.Lock()
	defer regInputs.updateMutex.Unlock()
	existingInput := regInputs.inputMap[input.Address]
	if existingInput == nil {
		return errors.New("UpdateInput but input " + input.Address + " doesn't exist.")
	}
	regInputs.updateInput(input)
	return nil
}

// updateInput replaces an existing input or adds the provided input.
// If the input doesn't exist it will be added. The input is also added to the updatedInputs map
func (regInputs *RegisteredInputs) updateInput(input *types.InputDiscoveryMessage) {

	regInputs.inputMap[input.Address] = input
	if regInputs.updatedInputs == nil {
		regInputs.updatedInputs = make(map[string]*types.InputDiscoveryMessage)
	}
	input.Timestamp = time.Now().Format(types.TimeFormat)
	regInputs.updatedInputs[input.Address] = input
}

// NotifyInputHandler passes the sender and value of an input command to the input's handler
// The sender is the identity address of the publisher and can be used for authorization. It is
// empty for locally triggered inputs such as file change and http polling.
func (regInputs *RegisteredInputs) NotifyInputHandler(inputAddress string, sender string, value string) {
	subscription := regInputs.subscriptions[inputAddress]
	if subscription != nil && subscription.handler != nil {
		subscription.handler(inputAddress, sender, value)
	}
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
		domain:        domain,
		publisherID:   publisherID,
		inputMap:      make(map[string]*types.InputDiscoveryMessage),
		subscriptions: make(map[string]*InputSubscription), // subscription to input by address
		updateMutex:   &sync.Mutex{},
	}
	return regInputs
}
