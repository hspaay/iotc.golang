// Package publisher with handling of input commands
package publisher

import (
	"encoding/json"
	"strings"

	"github.com/hspaay/iotc.golang/iotc"
	"github.com/hspaay/iotc.golang/messenger"
	"gopkg.in/square/go-jose.v2"
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
	// zone/pub/node/inputtype/instance/$input
	segments[5] = iotc.MessageTypeInputDiscovery
	inputAddr := strings.Join(segments, "/")

	input := publisher.Inputs.GetInputByAddress(inputAddr)

	if input == nil || message == "" {
		publisher.logger.Infof("handleNodeInput unknown input for address %s or missing message", address)
		return
	}

	payload := string(message)
	// decode the jws signature
	if publisher.signingMethod == SigningMethodJWS {
		// determine the sender public signing key so we can check the message signature
		jwsSignature, err := jose.ParseSigned(message)
		if err != nil {
			publisher.logger.Warningf("handleNodeInput: Not a JWS signed message on address %s: %s", address, err)
			return
		}
		payload = string(jwsSignature.UnsafePayloadWithoutVerification())

		err = json.Unmarshal([]byte(payload), &setMessage)
		if err != nil {
			publisher.logger.Infof("handleNodeInput: Unable to unmarshal message in %s", address)
			return
		}
		publicKey := publisher.domainPublishers.GetPublisherSigningKey(setMessage.Sender)

		_, err = messenger.VerifyJWSMessage(message, publicKey)
		if err != nil {
			publisher.logger.Warningf("handleNodeInput: verification failed for sender: %s", setMessage.Sender)
			return
		}
	} else {
		err := json.Unmarshal([]byte(payload), &setMessage)
		if err != nil {
			publisher.logger.Infof("handleNodeInput: Unable to unmarshal message in %s", address)
			return
		}
	}
	if publisher.onNodeInputHandler != nil {
		publisher.onNodeInputHandler(input, &setMessage)
	}
}
