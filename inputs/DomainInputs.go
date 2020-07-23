// Package inputs with managing of discovered inputs
package inputs

import (
	"fmt"
	"reflect"

	"github.com/iotdomain/iotdomain-go/lib"
	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/iotdomain/iotdomain-go/types"
)

// DomainInputs for managing discovered inputs.
type DomainInputs struct {
	c lib.DomainCollection //
	// getPublisherKey func(address string) *ecdsa.PublicKey // get publisher key for signature verification
	// inputMap      map[string]*types.InputDiscoveryMessage
	// messageSigner *messaging.MessageSigner // subscription to input discovery messages
	// updateMutex   *sync.Mutex              // mutex for async updating of inputs
}

// AddInput adds or replaces the input.
func (domainInputs *DomainInputs) AddInput(input *types.InputDiscoveryMessage) {
	domainInputs.c.Add(input.Address, input)
}

// GetAllInputs returns a new list with the inputs from this collection
func (domainInputs *DomainInputs) GetAllInputs() []*types.InputDiscoveryMessage {
	allInputs := make([]*types.InputDiscoveryMessage, 0)
	domainInputs.c.GetAll(&allInputs)
	return allInputs
}

// GetNodeInputs returns all inputs of a node
// Returns nil if the node has no known input
func (domainInputs *DomainInputs) GetNodeInputs(nodeAddress string) []*types.InputDiscoveryMessage {
	var inputList = make([]*types.InputDiscoveryMessage, 0)
	domainInputs.c.GetNodeItems(nodeAddress, &inputList)
	return inputList
}

// GetInputByAddress returns an input by its address
// inputAddr must contain the full input address, eg <zone>/<publisher>/<node>/"$input"/<type>/<instance>
// Returns nil if address has no known input
func (domainInputs *DomainInputs) GetInputByAddress(inputAddr string) *types.InputDiscoveryMessage {
	var inputObject = domainInputs.c.GetByAddress(inputAddr)
	if inputObject == nil {
		return nil
	}
	return inputObject.(*types.InputDiscoveryMessage)
}

// RemoveInput removes an input using its address.
// If the input doesn't exist, this is ignored.
func (domainInputs *DomainInputs) RemoveInput(inputAddress string) {
	domainInputs.c.Remove(inputAddress)
}

// Start subscribing to input discovery
func (domainInputs *DomainInputs) Start() {
	// subscription address for all inputs domain/publisher/node/type/instance/$input
	// TODO: Only subscribe to selected publishers
	addr := MakeInputDiscoveryAddress("+", "+", "+", "+", "+")
	domainInputs.c.MessageSigner.Subscribe(addr, domainInputs.handleDiscoverInput)
}

// Stop polling for inputs
func (domainInputs *DomainInputs) Stop() {
	addr := MakeInputDiscoveryAddress("+", "+", "+", "+", "+")
	domainInputs.c.MessageSigner.Unsubscribe(addr, domainInputs.handleDiscoverInput)
}

// handleDiscoverInput updates the domain input list with discovered inputs
// This verifies that the input discovery message is properly signed by its publisher
func (domainInputs *DomainInputs) handleDiscoverInput(address string, message string) error {
	var discoMsg types.InputDiscoveryMessage

	err := domainInputs.c.HandleDiscovery(address, message, &discoMsg)
	return err
}

// MakeInputDiscoveryAddress creates the address for the input discovery
func MakeInputDiscoveryAddress(domain string, publisherID string, nodeID string, inputType types.InputType, instance string) string {
	address := fmt.Sprintf("%s/%s/%s"+"/%s/%s/"+types.MessageTypeInputDiscovery,
		domain, publisherID, nodeID, inputType, instance)
	return address
}

// NewDomainInputs creates a new instance for handling of discovered domain inputs
func NewDomainInputs(messageSigner *messaging.MessageSigner) *DomainInputs {

	inputs := DomainInputs{
		c: lib.NewDomainCollection(messageSigner, reflect.TypeOf(&types.InputDiscoveryMessage{})),
	}
	return &inputs
}
