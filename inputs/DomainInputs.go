// Package inputs with managing of discovered inputs
package inputs

import (
	"fmt"
	"strings"
	"sync"

	"github.com/iotdomain/iotdomain-go/lib"
	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/iotdomain/iotdomain-go/types"
)

// DomainInputs for managing discovered inputs.
type DomainInputs struct {
	// getPublisherKey func(address string) *ecdsa.PublicKey // get publisher key for signature verification
	inputMap      map[string]*types.InputDiscoveryMessage
	messageSigner *messaging.MessageSigner // subscription to input discovery messages
	updateMutex   *sync.Mutex              // mutex for async updating of inputs
}

// AddInput adds or replaces the input.
// If the input doesn't exist, it will be added, otherwise replaced.
func (domainInputs *DomainInputs) AddInput(input *types.InputDiscoveryMessage) {
	domainInputs.updateMutex.Lock()
	defer domainInputs.updateMutex.Unlock()
	domainInputs.inputMap[input.Address] = input
}

// GetAllInputs returns a new list with the inputs from this collection
func (domainInputs *DomainInputs) GetAllInputs() []*types.InputDiscoveryMessage {
	domainInputs.updateMutex.Lock()
	defer domainInputs.updateMutex.Unlock()

	var inputList = make([]*types.InputDiscoveryMessage, 0)
	for _, input := range domainInputs.inputMap {
		inputList = append(inputList, input)
	}
	return inputList
}

// GetNodeInputs returns all inputs of a node
// Returns nil if the node has no known input
func (domainInputs *DomainInputs) GetNodeInputs(nodeAddress string) []*types.InputDiscoveryMessage {
	var inputList = make([]*types.InputDiscoveryMessage, 0)
	segments := strings.Split(nodeAddress, "/")
	nodePrefix := strings.Join(segments[:2], "/")

	domainInputs.updateMutex.Lock()
	defer domainInputs.updateMutex.Unlock()
	for _, input := range domainInputs.inputMap {
		if strings.HasPrefix(input.Address, nodePrefix) {
			inputList = append(inputList, input)
		}
	}
	return inputList
}

// GetInputByAddress returns an input by its address
// inputAddr must contain the full input address, eg <zone>/<publisher>/<node>/"$input"/<type>/<instance>
// Returns nil if address has no known input
func (domainInputs *DomainInputs) GetInputByAddress(inputAddr string) *types.InputDiscoveryMessage {
	domainInputs.updateMutex.Lock()
	defer domainInputs.updateMutex.Unlock()
	var input = domainInputs.inputMap[inputAddr]
	return input
}

// RemoveInput removes an input using its address.
// If the input doesn't exist, this is ignored.
func (domainInputs *DomainInputs) RemoveInput(inputAddress string) {
	domainInputs.updateMutex.Lock()
	defer domainInputs.updateMutex.Unlock()
	delete(domainInputs.inputMap, inputAddress)
}

// Start subscribing to input discovery
func (domainInputs *DomainInputs) Start() {
	// subscription address for all inputs domain/publisher/node/type/instance/$input
	// TODO: Only subscribe to selected publishers
	addr := MakeInputDiscoveryAddress("+", "+", "+", "+", "+")
	domainInputs.messageSigner.Subscribe(addr, domainInputs.handleDiscoverInput)
}

// Stop polling for inputs
func (domainInputs *DomainInputs) Stop() {
	addr := MakeInputDiscoveryAddress("+", "+", "+", "+", "+")
	domainInputs.messageSigner.Unsubscribe(addr, domainInputs.handleDiscoverInput)
}

// handleDiscoverInput updates the domain input list with discovered inputs
// This verifies that the input discovery message is properly signed by its publisher
func (domainInputs *DomainInputs) handleDiscoverInput(address string, message string) error {
	var discoMsg types.InputDiscoveryMessage

	// verify the message signature and get the payload
	_, err := domainInputs.messageSigner.VerifySignedMessage(message, &discoMsg)
	if err != nil {
		return lib.MakeErrorf("handleDiscoverInput: Failed verifying signature on address %s: %s", address, err)
	}
	segments := strings.Split(address, "/")
	discoMsg.PublisherID = segments[1]
	discoMsg.NodeID = segments[2]
	discoMsg.InputType = types.InputType(segments[3])
	discoMsg.Instance = segments[4]
	domainInputs.AddInput(&discoMsg)
	return nil
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
		inputMap:      make(map[string]*types.InputDiscoveryMessage),
		messageSigner: messageSigner,
		updateMutex:   &sync.Mutex{},
	}
	return &inputs
}
