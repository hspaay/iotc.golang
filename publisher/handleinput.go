// Package publisher with handling of input commands
package publisher

import (
	"encoding/json"
	"strings"

	"github.com/hspaay/iotconnect.golang/messaging"
)

// handle an incoming a set command for one of our nodes. This:
// - check if the signature is valid
// - check if the node is valid
// - pass the input value update to the adapter's onNodeInputHandler callback
func (publisher *Publisher) handleNodeInput(address string, publication *messaging.Publication) {
	// Check that address is one of our inputs
	segments := strings.Split(address, "/")
	segments[3] = messaging.MessageTypeInputDiscovery
	inputAddr := strings.Join(segments, "/")

	input := publisher.Inputs.GetInputByAddress(inputAddr)

	if input == nil || publication.Message == nil {
		publisher.Logger.Infof("handleNodeInput unknown input for address %s or missing message", address)
		return
	}
	// Decode the message into a SetMessage type
	var setMessage messaging.SetInputMessage
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
	if publisher.onNodeInputHandler != nil {
		publisher.onNodeInputHandler(input, &setMessage)
	}
}
