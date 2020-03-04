// Package publisher with handling of node inputs
// TODO: support for authorization per node
package publisher

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hspaay/iotconnect.golang/messenger"
	"github.com/hspaay/iotconnect.golang/standard"
)

// GetInput returns the input of one of this publisher's nodes
// Returns nil if address has no known input
// address with node type and instance. The command will be ignored.
func (publisher *PublisherState) GetInput(
	node *standard.Node, outputType string, instance string) *standard.InOutput {
	// segments := strings.Split(address, "/")
	// segments[3] = standard.CommandInputDiscovery
	// inputAddr := strings.Join(segments, "/")
	inputAddr := fmt.Sprintf("%s/%s/%s/%s/%s/%s", node.Zone, node.PublisherID, node.ID,
		standard.CommandInputDiscovery, outputType, instance)

	publisher.updateMutex.Lock()
	var input = publisher.inputs[inputAddr]
	publisher.updateMutex.Unlock()
	return input
}

// VerifyMessageSignature Verify the message is signed by the sender
// The node of the sender must have been received for its public key
func (publisher *PublisherState) VerifyMessageSignature(
	sender string, message json.RawMessage, base64signature string) bool {

	publisher.updateMutex.Lock()
	node := publisher.zonePublishers[sender]
	publisher.updateMutex.Unlock()

	if node == nil {
		publisher.Logger.Warningf("VerifyMessageSignature unknown sender %s", sender)
		return false
	}
	var pubKey *ecdsa.PublicKey = standard.DecodePublicKey(node.Identity.PublicKeySigning)
	valid := standard.VerifyEcdsaSignature(message, base64signature, pubKey)
	return valid
}

// handle an incoming a set command for one of our nodes. This:
// - check if the signature is valid
// - check if the node is valid
// - pass the input value update to the adapter's callback method set in Start()
func (publisher *PublisherState) handleNodeInput(address string, publication *messenger.Publication) {
	// Check that address is one of our inputs
	segments := strings.Split(address, "/")
	segments[3] = standard.CommandInputDiscovery
	inputAddr := strings.Join(segments, "/")

	publisher.updateMutex.Lock()
	var input = publisher.inputs[inputAddr]
	publisher.updateMutex.Unlock()

	if input == nil || publication.Message == nil {
		publisher.Logger.Infof("handleNodeInput unknown input for address %s or missing message", address)
		return
	}
	// Decode the message into a SetMessage type
	var setMessage standard.SetMessage
	err := json.Unmarshal([]byte(publication.Message), &setMessage)
	if err != nil {
		publisher.Logger.Infof("Unable to unmarshal SetMessage in %s", address)
		return
	}
	// Verify that the message comes from the sender using the sender's public key
	isValid := publisher.VerifyMessageSignature(setMessage.Sender, publication.Message, publication.Signature)
	if !isValid {
		publisher.Logger.Warningf("Incoming message verification failed for sender: %s", setMessage.Sender)
		return
	}
	if publisher.onSetInput != nil {
		publisher.onSetInput(input, &setMessage)
	}
}
