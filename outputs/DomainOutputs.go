// Package outputs with managing of discovered outputs
package outputs

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/sirupsen/logrus"
)

// DomainOutputs for managing discovered outputs
type DomainOutputs struct {
	outputMap     map[string]*types.OutputDiscoveryMessage
	messageSigner *messaging.MessageSigner // subscription to output discovery messages
	updateMutex   *sync.Mutex              // mutex for async updating of outputs
}

// GetAllOutputs returns a new list with the outputs from this collection
func (outputs *DomainOutputs) GetAllOutputs() []*types.OutputDiscoveryMessage {
	outputs.updateMutex.Lock()
	defer outputs.updateMutex.Unlock()

	var outputList = make([]*types.OutputDiscoveryMessage, 0)
	for _, output := range outputs.outputMap {
		outputList = append(outputList, output)
	}
	return outputList
}

// GetOutput returns the output of one of this publisher's nodes
// Returns nil if address has no known output
func (outputs *DomainOutputs) GetOutput(
	nodeAddress string, outputType types.OutputType, instance string) *types.OutputDiscoveryMessage {

	segments := strings.Split(nodeAddress, "/")
	outputAddr := MakeOutputDiscoveryAddress(segments[0], segments[1], segments[2], outputType, instance)

	outputs.updateMutex.Lock()
	defer outputs.updateMutex.Unlock()
	var output = outputs.outputMap[outputAddr]
	return output
}

// GetOutputByAddress returns an output by its address
// outputAddr must contain the full domain output address, eg <zone>/<publisher>/<node>/"$output"/<type>/<instance>
// Returns nil if address has no known output
func (outputs *DomainOutputs) GetOutputByAddress(outputAddr string) *types.OutputDiscoveryMessage {
	outputs.updateMutex.Lock()
	defer outputs.updateMutex.Unlock()
	var output = outputs.outputMap[outputAddr]
	return output
}

// Start subscribing to output discovery
func (outputs *DomainOutputs) Start() {
	// subscription address for all outputs domain/publisher/node/type/instance/$output
	// TODO: Only subscribe to selected publishers
	addr := MakeOutputDiscoveryAddress("+", "+", "+", "+", "+")
	outputs.messageSigner.Subscribe(addr, outputs.handleDiscoverOutput)
}

// Stop polling for outputs
func (outputs *DomainOutputs) Stop() {
	addr := MakeOutputDiscoveryAddress("+", "+", "+", "+", "+")
	outputs.messageSigner.Unsubscribe(addr, outputs.handleDiscoverOutput)
}

// UpdateOutput replaces the output using the node.Address
func (outputs *DomainOutputs) UpdateOutput(output *types.OutputDiscoveryMessage) {
	outputs.updateMutex.Lock()
	defer outputs.updateMutex.Unlock()
	outputs.outputMap[output.Address] = output
}

// handleDiscoverOutput updates the domain output list with discovered outputs
// This verifies that the output discovery message is properly signed by its publisher
func (outputs *DomainOutputs) handleDiscoverOutput(address string, message string) error {
	var discoMsg types.OutputDiscoveryMessage

	// verify the message signature and get the payload
	_, err := outputs.messageSigner.VerifySignedMessage(message, &discoMsg)
	if err != nil {
		errText := fmt.Sprintf("handleDiscoverOutput: Failed verifying signature on address %s: %s", address, err)
		logrus.Warn(errText)
		return errors.New(errText)
	}
	segments := strings.Split(address, "/")
	discoMsg.PublisherID = segments[1]
	discoMsg.NodeID = segments[2]
	discoMsg.OutputType = types.OutputType(segments[3])
	discoMsg.Instance = segments[4]
	outputs.UpdateOutput(&discoMsg)
	return nil
}

// MakeOutputDiscoveryAddress creates the address for the output discovery
func MakeOutputDiscoveryAddress(domain string, publisherID string, nodeID string, outputType types.OutputType, instance string) string {
	address := fmt.Sprintf("%s/%s/%s"+"/%s/%s/"+types.MessageTypeOutputDiscovery,
		domain, publisherID, nodeID, outputType, instance)
	return address
}

// NewDomainOutputs creates a new instance for handling of discovered domain outputs
func NewDomainOutputs(messageSigner *messaging.MessageSigner) *DomainOutputs {

	outputs := DomainOutputs{
		outputMap:     make(map[string]*types.OutputDiscoveryMessage),
		messageSigner: messageSigner,
		updateMutex:   &sync.Mutex{},
	}
	return &outputs
}
