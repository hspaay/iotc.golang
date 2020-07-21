// Package outputs with registered outputs from the local publisher
package outputs

import (
	"strings"
	"sync"
	"time"

	"github.com/iotdomain/iotdomain-go/types"
)

// RegisteredOutputs manages registration of publisher outputs
type RegisteredOutputs struct {
	domain         string                                   // the domain of this publisher
	publisherID    string                                   // the registered publisher for the inputs
	outputMap      map[string]*types.OutputDiscoveryMessage // output discovery address - object map
	updatedOutputs map[string]*types.OutputDiscoveryMessage // address of outputs that have been rediscovered/updated since last publication
	updateMutex    *sync.Mutex                              // mutex for async updating of outputs
}

// CreateOutput creates and registers a new output. If the output already exists, it is replaced.
func (regOutputs *RegisteredOutputs) CreateOutput(
	nodeID string, outputType types.OutputType, instance string) *types.OutputDiscoveryMessage {
	output := NewOutput(regOutputs.domain, regOutputs.publisherID, nodeID, outputType, instance)

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
	for _, output := range regOutputs.outputMap {
		outputList = append(outputList, output)
	}
	return outputList
}

// GetNodeOutputs returns all outputs for the given node in this list
// This method is concurrent safe
func (regOutputs *RegisteredOutputs) GetNodeOutputs(nodeAddress string) []*types.OutputDiscoveryMessage {
	nodeOutputs := []*types.OutputDiscoveryMessage{}
	segments := strings.Split(nodeAddress, "/")
	prefix := strings.Join(segments[:3], "/")

	for _, output := range regOutputs.outputMap {
		if strings.HasPrefix(output.Address, prefix) {
			nodeOutputs = append(nodeOutputs, output)
		}
	}
	return nodeOutputs
}

// GetOutput returns one of this publisher's registered outputs
// This method is concurrent safe
// Returns nil if address has no known output
func (regOutputs *RegisteredOutputs) GetOutput(
	nodeID string, outputType types.OutputType, instance string) *types.OutputDiscoveryMessage {
	outputAddr := MakeOutputDiscoveryAddress(regOutputs.domain, regOutputs.publisherID, nodeID, outputType, instance)

	regOutputs.updateMutex.Lock()
	defer regOutputs.updateMutex.Unlock()
	var output = regOutputs.outputMap[outputAddr]
	return output
}

// GetOutputByAddress returns an output by its address
// outputAddr must contain the full output address, eg <zone>/<publisher>/<node>/"$output"/<type>/<instance>
// Returns nil if address has no known output
// This method is concurrent safe
func (regOutputs *RegisteredOutputs) GetOutputByAddress(outputAddr string) *types.OutputDiscoveryMessage {
	regOutputs.updateMutex.Lock()
	var output = regOutputs.outputMap[outputAddr]
	regOutputs.updateMutex.Unlock()
	return output
}

// GetOutputsByNode returns a list of all outputs of a given node
func (regOutputs *RegisteredOutputs) GetOutputsByNode(nodeID string) []*types.OutputDiscoveryMessage {
	outputList := make([]*types.OutputDiscoveryMessage, 0)
	regOutputs.updateMutex.Lock()
	defer regOutputs.updateMutex.Unlock()
	for _, output := range regOutputs.outputMap {
		if output.NodeID == nodeID {
			outputList = append(outputList, output)
		}
	}
	return outputList
}

// GetUpdatedOutputs returns the list of discovered outputs that have been updated
// clear the update on return
func (regOutputs *RegisteredOutputs) GetUpdatedOutputs(clearUpdates bool) []*types.OutputDiscoveryMessage {
	var updateList []*types.OutputDiscoveryMessage = make([]*types.OutputDiscoveryMessage, 0)

	regOutputs.updateMutex.Lock()
	if regOutputs.updatedOutputs != nil {
		for _, output := range regOutputs.updatedOutputs {
			updateList = append(updateList, output)
		}
		if clearUpdates {
			regOutputs.updatedOutputs = nil
		}
	}
	regOutputs.updateMutex.Unlock()
	return updateList
}

// SetAlias updates the address of all inputs of the given nodeID using the alias instead
func (regOutputs *RegisteredOutputs) SetAlias(nodeID string, alias string) {
	nodeOutputs := regOutputs.GetOutputsByNode(nodeID)
	for _, output := range nodeOutputs {
		oldAddress := output.Address
		newAddress := MakeOutputDiscoveryAddress(
			regOutputs.domain, regOutputs.publisherID, alias, output.OutputType, output.Instance)
		output.Address = newAddress
		output.NodeID = alias

		regOutputs.updateMutex.Lock()
		defer regOutputs.updateMutex.Unlock()
		delete(regOutputs.outputMap, oldAddress)
		regOutputs.updateOutput(output)
	}
}

// UpdateOutput replaces the output and updates its timestamp.
// For internal use only. Use within locked section.
func (regOutputs *RegisteredOutputs) updateOutput(output *types.OutputDiscoveryMessage) {
	regOutputs.outputMap[output.Address] = output
	if regOutputs.updatedOutputs == nil {
		regOutputs.updatedOutputs = make(map[string]*types.OutputDiscoveryMessage)
	}
	output.Timestamp = time.Now().Format(types.TimeFormat)

	regOutputs.updatedOutputs[output.Address] = output
}

// NewOutput creates a new output for the given node address.
// It is not immediately added to allow for further updates of the ouput definition.
// To add it to the list use 'UpdateOutput'
func NewOutput(domain string, publisherID string, nodeID string, outputType types.OutputType, instance string) *types.OutputDiscoveryMessage {
	address := MakeOutputDiscoveryAddress(domain, publisherID, nodeID, outputType, instance)

	output := &types.OutputDiscoveryMessage{
		Address:     address,
		PublisherID: publisherID,
		NodeID:      nodeID,
		OutputType:  outputType,
		Instance:    instance,
		Timestamp:   time.Now().Format(types.TimeFormat),
	}
	return output
}

// NewRegisteredOutputs creates a new instance for registered output management
func NewRegisteredOutputs(domain string, publisherID string) *RegisteredOutputs {
	regOutputs := RegisteredOutputs{
		domain:      domain,
		publisherID: publisherID,
		outputMap:   make(map[string]*types.OutputDiscoveryMessage),
		updateMutex: &sync.Mutex{},
	}
	return &regOutputs
}
