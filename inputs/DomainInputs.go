// Package inputs with managing of discovered inputs
package inputs

import (
	"crypto/ecdsa"
	"fmt"
	"strings"
	"sync"

	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/sirupsen/logrus"
)

// DomainInputs for managing discovered inputs.
type DomainInputs struct {
	getPublisherKey func(address string) *ecdsa.PublicKey // get publisher key for signature verification
	inputMap        map[string]*types.InputDiscoveryMessage
	messageSigner   *messaging.MessageSigner // subscription to input discovery messages
	updateMutex     *sync.Mutex              // mutex for async updating of inputs
}

// GetAllInputs returns a new list with the inputs from this collection
func (inputs *DomainInputs) GetAllInputs() []*types.InputDiscoveryMessage {
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
func (inputs *DomainInputs) GetInput(
	nodeAddress string, inputType types.InputType, instance string) *types.InputDiscoveryMessage {

	segments := strings.Split(nodeAddress, "/")
	inputAddr := MakeInputDiscoveryAddress(segments[0], segments[1], segments[2], inputType, instance)

	inputs.updateMutex.Lock()
	defer inputs.updateMutex.Unlock()
	var input = inputs.inputMap[inputAddr]
	return input
}

// GetInputByAddress returns an input by its address
// inputAddr must contain the full input address, eg <zone>/<publisher>/<node>/"$input"/<type>/<instance>
// Returns nil if address has no known input
func (inputs *DomainInputs) GetInputByAddress(inputAddr string) *types.InputDiscoveryMessage {
	inputs.updateMutex.Lock()
	defer inputs.updateMutex.Unlock()
	var input = inputs.inputMap[inputAddr]
	return input
}

// Start subscribing to input discovery
func (inputs *DomainInputs) Start() {
	// subscription address for all inputs domain/publisher/node/type/instance/$input
	// TODO: Only subscribe to selected publishers
	addr := MakeInputDiscoveryAddress("+", "+", "+", "+", "+")
	inputs.messageSigner.Subscribe(addr, inputs.handleDiscoverInput)
}

// Stop polling for inputs
func (inputs *DomainInputs) Stop() {
	addr := MakeInputDiscoveryAddress("+", "+", "+", "+", "+")
	inputs.messageSigner.Unsubscribe(addr, inputs.handleDiscoverInput)
}

// UpdateInput replaces the input using the node.Address
func (inputs *DomainInputs) UpdateInput(input *types.InputDiscoveryMessage) {
	inputs.updateMutex.Lock()
	defer inputs.updateMutex.Unlock()
	inputs.inputMap[input.Address] = input
}

// handleDiscoverInput updates the domain input list with discovered inputs
// This verifies that the input discovery message is properly signed by its publisher
func (inputs *DomainInputs) handleDiscoverInput(address string, message string) {
	var discoMsg types.InputDiscoveryMessage

	// verify the message signature and get the payload
	_, err := messaging.VerifySignature(message, &discoMsg, inputs.getPublisherKey)
	if err != nil {
		logrus.Warnf("handleDiscoverInput: Failed verifying signature on address %s: %s", address, err)
		return
	}
	segments := strings.Split(address, "/")
	discoMsg.PublisherID = segments[1]
	discoMsg.NodeID = segments[2]
	discoMsg.InputType = types.InputType(segments[3])
	discoMsg.Instance = segments[4]
	inputs.UpdateInput(&discoMsg)
}

// MakeInputDiscoveryAddress creates the address for the input discovery
func MakeInputDiscoveryAddress(domain string, publisherID string, nodeID string, inputType types.InputType, instance string) string {
	address := fmt.Sprintf("%s/%s/%s"+"/%s/%s/"+types.MessageTypeInputDiscovery,
		domain, publisherID, nodeID, inputType, instance)
	return address
}

// NewDomainInputs creates a new instance for handling of discovered domain inputs
func NewDomainInputs(
	getPublisherKey func(string) *ecdsa.PublicKey,
	messageSigner *messaging.MessageSigner) *DomainInputs {

	inputs := DomainInputs{
		inputMap:        make(map[string]*types.InputDiscoveryMessage),
		getPublisherKey: getPublisherKey,
		messageSigner:   messageSigner,
		updateMutex:     &sync.Mutex{},
	}
	return &inputs
}
