// Package outputs with managing of discovered outputs
package outputs

import (
	"fmt"
	"reflect"

	"github.com/iotdomain/iotdomain-go/lib"
	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/iotdomain/iotdomain-go/types"
)

// DomainOutputs for managing discovered outputs
type DomainOutputs struct {
	c             lib.DomainCollection //
	messageSigner *messaging.MessageSigner
}

// AddOutput adds or replaces the output
func (domainOutputs *DomainOutputs) AddOutput(output *types.OutputDiscoveryMessage) {
	domainOutputs.c.Add(output.Address, output)
}

// GetAllOutputs returns a new list with the outputs from this collection
func (domainOutputs *DomainOutputs) GetAllOutputs() []*types.OutputDiscoveryMessage {
	allOutputs := make([]*types.OutputDiscoveryMessage, 0)
	domainOutputs.c.GetAll(&allOutputs)
	return allOutputs
}

// GetNodeOutputs returns all outputs of a node
// Returns nil if the node has no known input
func (domainOutputs *DomainOutputs) GetNodeOutputs(nodeAddress string) []*types.OutputDiscoveryMessage {
	var outputList = make([]*types.OutputDiscoveryMessage, 0)
	domainOutputs.c.GetByAddressPrefix(nodeAddress, &outputList)
	return outputList
}

// GetOutput returns the output of one of this publisher's nodes. Note this requires the
// full node Address with domain/publisher/nodeID.
// Returns nil if address has no known output
func (domainOutputs *DomainOutputs) GetOutput(
	nodeAddress string, outputType types.OutputType, instance string) *types.OutputDiscoveryMessage {

	var outputObject = domainOutputs.c.Get(nodeAddress, string(outputType), instance)
	if outputObject == nil {
		return nil
	}
	return outputObject.(*types.OutputDiscoveryMessage)
}

// GetOutputByAddress returns an output by its address
// outputAddr must contain the full domain output address, eg <zone>/<publisher>/<node>/"$output"/<type>/<instance>
// Returns nil if address has no known output
func (domainOutputs *DomainOutputs) GetOutputByAddress(address string) *types.OutputDiscoveryMessage {
	var outputObject = domainOutputs.c.GetByAddress(address)
	if outputObject == nil {
		return nil
	}
	return outputObject.(*types.OutputDiscoveryMessage)
}

// RemoveOutput removes an output using its address.
// If the output doesn't exist, this is ignored.
func (domainOutputs *DomainOutputs) RemoveOutput(address string) {
	domainOutputs.c.Remove(address)
}

// Start subscribing to output discovery
func (domainOutputs *DomainOutputs) Start() {
	// subscription address for all outputs domain/publisher/node/type/instance/$output
	// TODO: Only subscribe to selected publishers
	address := MakeOutputDiscoveryAddress("+", "+", "+", "+", "+")
	domainOutputs.messageSigner.Subscribe(address, domainOutputs.handleDiscoverOutput)
}

// Stop polling for outputs
func (domainOutputs *DomainOutputs) Stop() {
	address := MakeOutputDiscoveryAddress("+", "+", "+", "+", "+")
	domainOutputs.messageSigner.Unsubscribe(address, domainOutputs.handleDiscoverOutput)
}

// handleDiscoverOutput updates the domain output list with discovered outputs
// This verifies that the output discovery message is properly signed by its publisher
func (domainOutputs *DomainOutputs) handleDiscoverOutput(address string, message string) error {
	var discoMsg types.OutputDiscoveryMessage

	err := domainOutputs.c.HandleDiscovery(address, message, &discoMsg)
	return err
}

// MakeOutputDiscoveryAddress creates the address for the output discovery
func MakeOutputDiscoveryAddress(domain string, publisherID string, nodeID string, outputType types.OutputType, instance string) string {
	address := fmt.Sprintf("%s/%s/%s"+"/%s/%s/"+types.MessageTypeOutputDiscovery,
		domain, publisherID, nodeID, outputType, instance)
	return address
}

// NewDomainOutputs creates a new instance for handling of discovered domain outputs
func NewDomainOutputs(messageSigner *messaging.MessageSigner) *DomainOutputs {
	return &DomainOutputs{
		c:             lib.NewDomainCollection(reflect.TypeOf(&types.OutputDiscoveryMessage{}), messageSigner.GetPublicKey),
		messageSigner: messageSigner,
	}
}
