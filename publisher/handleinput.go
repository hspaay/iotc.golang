// Package publisher with handling of input commands
package publisher

import (
	"strings"

	"github.com/hspaay/iotc.golang/iotc"
	"github.com/hspaay/iotc.golang/messenger"
)

// handle an incoming a set command for input of one of our nodes. This:
// - checks if the signature is valid
// - checks if the node is valid
// - pass the input value update to the adapter's onNodeInputHandler callback
func (publisher *Publisher) handleNodeInput(address string, message string) {
	var setMessage iotc.SetInputMessage

	// Check that address is one of our inputs
	segments := strings.Split(address, "/")
	// a full address is required
	if len(segments) < 6 {
		return
	}
	// domain/pub/node/inputtype/instance/$input
	segments[5] = iotc.MessageTypeInputDiscovery
	inputAddr := strings.Join(segments, "/")

	input := publisher.Inputs.GetInputByAddress(inputAddr)

	if input == nil || message == "" {
		publisher.logger.Infof("handleNodeInput unknown input for address %s or missing message", address)
		return
	}

	// Verify the message using the public key of the sender
	isSigned, err := messenger.VerifySender(message, &setMessage, publisher.domainPublishers.GetPublisherSigningKey)
	if !isSigned {
		if publisher.signingMethod != SigningMethodNone {
			// all inputs must use signed messages
			publisher.logger.Warnf("handleNodeInput: message to input '%s' is not signed. Message discarded.", address)
			return
		}
	} else if err != nil {
		// signing failed, discard the message
		publisher.logger.Warnf("handleNodeInput: signature verification failed for message to input %s. Message discarded.", address)
		return
	}

	if publisher.onNodeInputHandler != nil {
		publisher.onNodeInputHandler(input, &setMessage)
	}
}
