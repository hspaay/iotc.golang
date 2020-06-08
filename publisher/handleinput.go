// Package publisher with handling of input commands
package publisher

import (
	"encoding/json"
	"strings"

	"github.com/hspaay/iotc.golang/iotc"
)

// handle an incoming a set command for input of one of our nodes. This:
// - check if the signature is valid
// - check if the node is valid
// - pass the input value update to the adapter's onNodeInputHandler callback
func (publisher *Publisher) handleNodeInput(address string, message []byte) {
	// Check that address is one of our inputs
	segments := strings.Split(address, "/")
	// a full address is required
	if len(segments) < 6 {
		return
	}
	// zone/pub/node/inputtype/instance/$input
	segments[5] = iotc.MessageTypeInputDiscovery
	inputAddr := strings.Join(segments, "/")

	input := publisher.Inputs.GetInputByAddress(inputAddr)

	if input == nil || message == nil {
		publisher.logger.Infof("handleNodeInput unknown input for address %s or missing message", address)
		return
	}
	// Decode the message into a SetMessage type
	payload, err := publisher.signer.Verify(message)
	if err != nil {
		publisher.logger.Warningf("handleNodeConfig Invalid message: %s", err)
		return
	}

	var setMessage iotc.SetInputMessage
	err = json.Unmarshal(payload, &setMessage)
	if err != nil {
		publisher.logger.Infof("Unable to unmarshal SetMessage in %s", address)
		return
	}
	// Verify that the message comes from the sender using the sender's public key
	// isValid := publisher.VerifyMessageSignature(setMessage.Sender, message, publication.Signature)
	// if !isValid {
	// 	publisher.Logger.Warningf("Incoming message verification failed for sender: %s", setMessage.Sender)
	// 	return
	// }
	if publisher.onNodeInputHandler != nil {
		publisher.onNodeInputHandler(input, &setMessage)
	}
}
