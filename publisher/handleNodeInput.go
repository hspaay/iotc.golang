// Package publisher with handling of input commands for this publisher
package publisher

import (
	"strings"

	"github.com/iotdomain/iotdomain-go/messenger"
	"github.com/iotdomain/iotdomain-go/types"
)

// handle an incoming a set command for input of one of our nodes. This:
// - checks if the signature is valid
// - checks if the node is valid
// - pass the input value update to the adapter's onNodeInputHandler callback
func (publisher *Publisher) handleNodeInput(address string, message string) {
	var setMessage types.SetInputMessage

	// Check that address is one of our inputs
	segments := strings.Split(address, "/")
	// a full address is required
	if len(segments) < 6 {
		return
	}
	// domain/pub/node/inputtype/instance/$input
	segments[5] = types.MessageTypeInputDiscovery
	inputAddr := strings.Join(segments, "/")
	input := publisher.Inputs.GetInputByAddress(inputAddr)

	if input == nil || message == "" {
		publisher.logger.Infof("handleNodeInput unknown input for address %s or missing message", address)
		return
	}

	// Expect the message to be encrypted
	isEncrypted, dmessage, err := messenger.DecryptMessage(message, publisher.identityPrivateKey)
	if !isEncrypted {
		publisher.logger.Infof("handleNodeInput: message to input '%s' is not encrypted.", address)
		// this could be fine, just warning
	} else if err != nil {
		publisher.logger.Warnf("handleNodeInput: decryption failed of message to input '%s'. Message discarded.", address)
		return
	}

	// Verify the message using the public key of the sender
	isSigned, err := messenger.VerifySender(dmessage, &setMessage, publisher.domainPublishers.GetPublisherKey)
	if !isSigned {
		if publisher.signingMethod != SigningMethodNone {
			// all inputs must use signed messages
			publisher.logger.Warnf("handleNodeInput: message to input '%s' is not signed. Message discarded.", address)
			return
		}
	} else if err != nil {
		// signing failed, discard the message
		publisher.logger.Warnf("handleNodeInput: signature verification failed for message to input %s: %s. Message discarded.", address, err)
		return
	}

	publisher.logger.Infof("handleNodeInput input command on address %s. isEncrypted=%t, isSigned=%t", address, isEncrypted, isSigned)

	if publisher.onNodeInputHandler != nil {
		publisher.onNodeInputHandler(input, &setMessage)
	}
}
