// Package outputs with registered outputs from the local publisher
package outputs

import (
	"sync"
	"time"

	"github.com/iotdomain/iotdomain-go/types"
)

// RegisteredOutputs manages registration of publisher outputs
type RegisteredOutputs struct {
	addressMap       map[string]string                        // lookup outputID by output publication address
	domain           string                                   // the domain of this publisher
	publisherID      string                                   // the registered publisher for the inputs
	outputsByID      map[string]*types.OutputDiscoveryMessage // lookup output by output ID
	updatedOutputIDs map[string]string                        // IDs of updated outputs
	updateMutex      *sync.Mutex                              // mutex for async updating of outputs
}

// CreateOutput creates and registers a new output. If the output already exists, it is replaced.
func (regOutputs *RegisteredOutputs) CreateOutput(
	deviceID string, outputType types.OutputType, instance string) *types.OutputDiscoveryMessage {

	output := NewOutput(regOutputs.domain, regOutputs.publisherID, deviceID, outputType, instance)

	regOutputs.updateMutex.Lock()
	defer regOutputs.updateMutex.Unlock()
	regOutputs.updateOutput(output)
	return output
}

// GetAllOutputs returns the list of outputs
func (regOutputs *RegisteredOutputs) GetAllOutputs() []*types.OutputDiscoveryMessage {
	regOutputs.updateMutex.Lock()
	defer regOutputs.updateMutex.Unlock()

	var outputList = make([]*types.OutputDiscoveryMessage, 0)
	for _, output := range regOutputs.outputsByID {
		outputList = append(outputList, output)
	}
	return outputList
}

// GetOutputByAddress returns an output by its address
// outputAddr must contain the full output address, eg <zone>/<publisher>/<node>/"$output"/<type>/<instance>
// Returns nil if address has no known output
func (regOutputs *RegisteredOutputs) GetOutputByAddress(outputAddr string) *types.OutputDiscoveryMessage {
	regOutputs.updateMutex.Lock()
	defer regOutputs.updateMutex.Unlock()
	var outputID = regOutputs.addressMap[outputAddr]
	output := regOutputs.outputsByID[outputID]
	return output
}

// GetOutputsByDeviceID returns a list of all outputs of a given device
func (regOutputs *RegisteredOutputs) GetOutputsByDeviceID(deviceID string) []*types.OutputDiscoveryMessage {
	outputList := make([]*types.OutputDiscoveryMessage, 0)
	regOutputs.updateMutex.Lock()
	defer regOutputs.updateMutex.Unlock()
	for _, output := range regOutputs.outputsByID {
		if output.DeviceID == deviceID {
			outputList = append(outputList, output)
		}
	}
	return outputList
}

// GetOutputByDevice returns one of this publisher's registered outputs
// This method is concurrent safe
// Returns nil if no known output
func (regOutputs *RegisteredOutputs) GetOutputByDevice(
	deviceID string, outputType types.OutputType, instance string) *types.OutputDiscoveryMessage {

	outputID := MakeOutputID(deviceID, outputType, instance)
	return regOutputs.GetOutputByID(outputID)
}

// GetOutputByID returns an output by its ID (device.type.instance)
// Returns nil if there is no known output
func (regOutputs *RegisteredOutputs) GetOutputByID(outputID string) *types.OutputDiscoveryMessage {
	regOutputs.updateMutex.Lock()
	defer regOutputs.updateMutex.Unlock()
	var output = regOutputs.outputsByID[outputID]
	return output
}

// GetUpdatedOutputs returns the list of discovered outputs that have been updated
// clear the update on return
func (regOutputs *RegisteredOutputs) GetUpdatedOutputs(clearUpdates bool) []*types.OutputDiscoveryMessage {
	var updateList []*types.OutputDiscoveryMessage = make([]*types.OutputDiscoveryMessage, 0)

	regOutputs.updateMutex.Lock()
	if regOutputs.updatedOutputIDs != nil {
		for _, outputID := range regOutputs.updatedOutputIDs {
			output := regOutputs.outputsByID[outputID]
			if output != nil {
				updateList = append(updateList, output)
			}
		}
		if clearUpdates {
			regOutputs.updatedOutputIDs = nil
		}
	}
	regOutputs.updateMutex.Unlock()
	return updateList
}

// SetAlias updates the address of all inputs with the given deviceID using the alias instead
func (regOutputs *RegisteredOutputs) SetAlias(deviceID string, alias string) {
	outputList := regOutputs.GetOutputsByDeviceID(deviceID)
	for _, output := range outputList {
		newAddress := MakeOutputDiscoveryAddress(
			regOutputs.domain, regOutputs.publisherID, alias, output.OutputType, output.Instance)
		output.Address = newAddress

		regOutputs.updateMutex.Lock()
		defer regOutputs.updateMutex.Unlock()
		regOutputs.updateOutput(output)
	}
}

// UpdateOutput replaces the output and updates its timestamp.
// For internal use only. Use within locked section.
func (regOutputs *RegisteredOutputs) updateOutput(output *types.OutputDiscoveryMessage) {

	regOutputs.outputsByID[output.OutputID] = output
	regOutputs.addressMap[output.Address] = output.OutputID

	if regOutputs.updatedOutputIDs == nil {
		regOutputs.updatedOutputIDs = make(map[string]string)
	}
	output.Timestamp = time.Now().Format(types.TimeFormat)
	regOutputs.updatedOutputIDs[output.OutputID] = output.OutputID
}

// MakeOutputID creates the internal ID to identify the output of the owning deviceID or serviceID.
func MakeOutputID(deviceID string, outputType types.OutputType, instance string) string {
	outputID := deviceID + "." + string(outputType) + "." + instance
	return outputID
}

// NewOutput creates a new output for the given device .
// It is not immediately added to allow for further updates of the ouput definition.
// To add it to the list use 'UpdateOutput'
func NewOutput(domain string, publisherID string, deviceID string, outputType types.OutputType, instance string) *types.OutputDiscoveryMessage {
	address := MakeOutputDiscoveryAddress(domain, publisherID, deviceID, outputType, instance)

	outputID := MakeOutputID(deviceID, outputType, instance)

	output := &types.OutputDiscoveryMessage{
		Address:   address,
		Timestamp: time.Now().Format(types.TimeFormat),
		// internal use only
		DeviceID:    deviceID,
		Instance:    instance,
		OutputID:    outputID,
		OutputType:  outputType,
		PublisherID: publisherID,
	}
	return output
}

// NewRegisteredOutputs creates a new instance for registered output management
func NewRegisteredOutputs(domain string, publisherID string) *RegisteredOutputs {
	regOutputs := RegisteredOutputs{
		domain:      domain,
		publisherID: publisherID,
		addressMap:  make(map[string]string),
		outputsByID: make(map[string]*types.OutputDiscoveryMessage),
		updateMutex: &sync.Mutex{},
	}
	return &regOutputs
}
