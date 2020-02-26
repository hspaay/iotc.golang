// Package publisher with handling commands for my node inputs
// TODO: support for authorization per node
package publisher

import (
	"crypto/ecdsa"
	"encoding/json"
	"iotzone/messenger"
	"iotzone/standard"
)

// VerifyMessageSignature Verify the message is signed by the sender
// The node of the sender must have been received for its public key
func (publisher *ThisPublisherState) VerifyMessageSignature(
	sender string, message json.RawMessage, base64signature string) bool {

	publisher.updateMutex.Lock()
	node := publisher.zonePublishers[sender]
	publisher.updateMutex.Unlock()

	if node == nil {
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
func (publisher *ThisPublisherState) handleNodeInput(address string, publication *messenger.Publication) {
	// Check that address is one of our inputs
	input := publisher.GetInput(address)
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
	if publisher.onSetMessage != nil {
		publisher.onSetMessage(input, &setMessage)
	}
}
