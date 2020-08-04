// Package inputs with managing and publishing of registered inputs
package inputs

import (
	"sync"
	"time"

	"github.com/iotdomain/iotdomain-go/lib"
	"github.com/iotdomain/iotdomain-go/types"
)

// // InputSubscription with handler of subscriber to input updates
// type InputSubscription struct {
// 	handler      func(inputAddress string, sender string, value string) // the notification handler of this input
// 	inputAddress string                                                 // the input belonging to this subscription
// 	source       string                                                 // the source of the input, eg, the input address, http address or file path
// }

// RegisteredInputs manages registration of publisher inputs
// Generics would be nice as this overlaps with outputs, nodes, publishers
// The inputID used in the inputMap consist of deviceID.inputType.instance
type RegisteredInputs struct {
	domain          string                                  // the domain of this publisher
	publisherID     string                                  // the registered publisher for the inputs
	addressMap      map[string]string                       // lookup inputID by publication address
	inputsByID      map[string]*types.InputDiscoveryMessage // lookup input by inputID
	updatedInputIDs map[string]string                       // inputIDs of inputs that have been rediscovered/updated
	updateMutex     *sync.Mutex                             // mutex for async handling of inputs
	// notification handlers by inputID
	handlers map[string]func(input *types.InputDiscoveryMessage, sender string, value string)
}

// CreateInput creates and registers a new input with optional handler for input trigger
func (regInputs *RegisteredInputs) CreateInput(
	deviceID string, inputType types.InputType, instance string,
	handler func(input *types.InputDiscoveryMessage, sender string, value string)) *types.InputDiscoveryMessage {

	return regInputs.CreateInputWithSource(deviceID, inputType, instance, "", handler)
}

// CreateInputWithSource creates and registers a new input that takes its input value from a given source
// Replaces the existing input if it already exist.
func (regInputs *RegisteredInputs) CreateInputWithSource(
	deviceID string, inputType types.InputType, instance string, source string,
	handler func(input *types.InputDiscoveryMessage, sender string, value string)) *types.InputDiscoveryMessage {

	regInputs.updateMutex.Lock()
	defer regInputs.updateMutex.Unlock()

	input := NewInput(regInputs.domain, regInputs.publisherID, deviceID, inputType, instance)
	input.Source = source

	regInputs.updateInput(input, handler)
	return input
}

// DeleteInput unregisters the input
func (regInputs *RegisteredInputs) DeleteInput(inputID string) {

	regInputs.updateMutex.Lock()
	defer regInputs.updateMutex.Unlock()

	// inputAddr := MakeInputDiscoveryAddress(regInputs.domain, regInputs.publisherID, nodeID, inputType, instance)
	delete(regInputs.inputsByID, inputID)
	delete(regInputs.handlers, inputID)
	if regInputs.updatedInputIDs == nil {
		regInputs.updatedInputIDs = make(map[string]string)
	}
	// "" updates mean that the input is deleted
	regInputs.updatedInputIDs[inputID] = ""
}

// GetAllInputs returns the list of inputs
func (regInputs *RegisteredInputs) GetAllInputs() []*types.InputDiscoveryMessage {
	regInputs.updateMutex.Lock()
	defer regInputs.updateMutex.Unlock()

	var inputList = make([]*types.InputDiscoveryMessage, 0)
	for _, output := range regInputs.inputsByID {
		inputList = append(inputList, output)
	}
	return inputList
}

// GetInputByAddress returns an input by its publication address
// Returns nil if address has no known input
func (regInputs *RegisteredInputs) GetInputByAddress(inputAddr string) *types.InputDiscoveryMessage {
	regInputs.updateMutex.Lock()
	defer regInputs.updateMutex.Unlock()
	inputID := regInputs.addressMap[inputAddr]
	input := regInputs.inputsByID[inputID]
	return input
}

// GetInputByDevice returns an input by deviceID, input type and instance
// Returns nil if the device has no such input
func (regInputs *RegisteredInputs) GetInputByDevice(
	deviceID string, inputType types.InputType, instance string) *types.InputDiscoveryMessage {

	inputID := MakeInputID(deviceID, inputType, instance)
	return regInputs.GetInputByID(inputID)
}

// GetInputsByDeviceID returns a list of all inputs of a given deviceID
func (regInputs *RegisteredInputs) GetInputsByDeviceID(deviceID string) []*types.InputDiscoveryMessage {
	inputList := make([]*types.InputDiscoveryMessage, 0)
	regInputs.updateMutex.Lock()
	defer regInputs.updateMutex.Unlock()
	for _, input := range regInputs.inputsByID {
		if input.DeviceID == deviceID {
			inputList = append(inputList, input)
		}
	}
	return inputList
}

// GetInputByID returns an input by its input ID (device.type.instance)
// Returns nil if there is no known input
func (regInputs *RegisteredInputs) GetInputByID(inputID string) *types.InputDiscoveryMessage {
	regInputs.updateMutex.Lock()
	defer regInputs.updateMutex.Unlock()
	var input = regInputs.inputsByID[inputID]
	return input
}

// GetInputsWithSource returns a list of inputs that have the given source
// The source is used for inputs that are files, http poll addresses or other outputs. It is not
// used with set input commands.
func (regInputs *RegisteredInputs) GetInputsWithSource(source string) []*types.InputDiscoveryMessage {
	inputList := make([]*types.InputDiscoveryMessage, 0)
	for _, input := range regInputs.inputsByID {
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
	if regInputs.updatedInputIDs != nil {
		for _, inputID := range regInputs.updatedInputIDs {
			input := regInputs.inputsByID[inputID]
			if input != nil {
				updateList = append(updateList, input)
			}
		}
		if clearUpdates {
			regInputs.updatedInputIDs = nil
		}
	}
	return updateList
}

// NotifyInputHandler passes a set input command to the input's handler to execute the request.
// The sender is the identity address of the publisher and can be used for authorization. It is
// empty for local inputs such as file watcher and http polling.
func (regInputs *RegisteredInputs) NotifyInputHandler(inputID string, sender string, value string) {

	handler := regInputs.handlers[inputID]
	input := regInputs.GetInputByID(inputID)
	if handler != nil {
		handler(input, sender, value)
	}
}

// SetAlias updates the publication address of all inputs with the given deviceID
func (regInputs *RegisteredInputs) SetAlias(deviceID string, alias string) {
	inputList := regInputs.GetInputsByDeviceID(deviceID)
	for _, input := range inputList {
		newAddress := MakeInputDiscoveryAddress(
			regInputs.domain, regInputs.publisherID, alias, input.InputType, input.Instance)
		input.Address = newAddress
		// input.NodeID = alias
		regInputs.updateMutex.Lock()
		regInputs.updateInput(input, nil)
		regInputs.updateMutex.Unlock()
	}
}

// UpdateInput replaces an existing input with the provided input.
// The input must already exist and be created using 'CreateInput', otherwise it returns an error
func (regInputs *RegisteredInputs) UpdateInput(input *types.InputDiscoveryMessage) error {

	regInputs.updateMutex.Lock()
	defer regInputs.updateMutex.Unlock()
	existingInput := regInputs.inputsByID[input.InputID]
	if existingInput == nil {
		return lib.MakeErrorf("UpdateInput: input '%s' does not exist", input.InputID)
	}
	regInputs.updateInput(input, nil)
	return nil
}

// updateInput replaces an existing input or adds the provided input.
// If the input doesn't exist it will be added. The input is also added to the updatedInputs map
// The handler for this input will be stored if provided. Use nil to retain the existing handler.
// Use within a locked section.
func (regInputs *RegisteredInputs) updateInput(input *types.InputDiscoveryMessage,
	handler func(input *types.InputDiscoveryMessage, sender string, value string)) {

	regInputs.inputsByID[input.InputID] = input
	regInputs.addressMap[input.Address] = input.InputID
	if handler != nil {
		regInputs.handlers[input.InputID] = handler
	}
	// track which inputs are updated
	if regInputs.updatedInputIDs == nil {
		regInputs.updatedInputIDs = make(map[string]string)
	}
	input.Timestamp = time.Now().Format(types.TimeFormat)
	regInputs.updatedInputIDs[input.InputID] = input.InputID
}

// MakeInputID creates the internal ID to identify the input of the owning deviceID or serviceID.
func MakeInputID(deviceID string, inputType types.InputType, instance string) string {
	inputID := deviceID + "." + string(inputType) + "." + instance
	return inputID
}

// NewInput instance for creating an input object for later adding.
// To add it to the inputlist use 'UpdateInput'
func NewInput(
	domain string, publisherID string, deviceID string, inputType types.InputType, instance string) *types.InputDiscoveryMessage {

	inputID := MakeInputID(deviceID, inputType, instance)
	address := MakeInputDiscoveryAddress(domain, publisherID, deviceID, inputType, instance)
	// segments := strings.Split(nodeAddress, "/")
	input := &types.InputDiscoveryMessage{
		Address:   address,
		Attr:      make(types.NodeAttrMap),
		Config:    make(types.ConfigAttrMap),
		Timestamp: time.Now().Format(types.TimeFormat),
		// internal use only
		InputID:     inputID,
		DeviceID:    deviceID,
		Instance:    instance,
		PublisherID: publisherID,
		InputType:   inputType,
	}
	return input
}

// NewRegisteredInputs creates a new instance for managing registered inputs
func NewRegisteredInputs(domain string, publisherID string) *RegisteredInputs {

	regInputs := &RegisteredInputs{
		domain:      domain,
		publisherID: publisherID,
		addressMap:  make(map[string]string),
		inputsByID:  make(map[string]*types.InputDiscoveryMessage),
		handlers:    make(map[string]func(input *types.InputDiscoveryMessage, sender string, newValue string)),
		updateMutex: &sync.Mutex{},
	}
	return regInputs
}
