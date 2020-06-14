// Package publisher with handling of configuration commands
package publisher

import (
	"encoding/json"

	"github.com/hspaay/iotc.golang/iotc"
	"github.com/hspaay/iotc.golang/messenger"
	"gopkg.in/square/go-jose.v2"
)

// handle an incoming a configuration command for one of our nodes. This:
// - check if the signature is valid
// - check if the node is valid
// - pass the configuration update to the adapter's callback set in Start()
// - save node configuration if persistence is set
// TODO: support for authorization per node
func (publisher *Publisher) handleNodeConfigCommand(address string, message string) {
	var configureMessage iotc.NodeConfigureMessage

	publisher.logger.Infof("handleNodeConfig on address %s", address)

	payload := string(message)

	// determine the sender public signing key so we can check the message signature
	// decode the jws signature
	if publisher.signingMethod == SigningMethodJWS {
		jwsSignature, err := jose.ParseSigned(message)
		if err != nil {
			publisher.logger.Warningf("handleNodeConfig: Not a JWS signed message on address %s: %s", address, err)
			return
		}

		payload = string(jwsSignature.UnsafePayloadWithoutVerification())

		err = json.Unmarshal([]byte(payload), &configureMessage)
		if err != nil {
			publisher.logger.Infof("Unable to unmarshal ConfigureMessage in %s", address)
			return
		}

		publicKey := publisher.domainPublishers.GetPublisherSigningKey(configureMessage.Sender)

		_, err = messenger.VerifyJWSMessage(message, publicKey)
		if err != nil {
			publisher.logger.Warningf("Incoming configuration verification failed for sender: %s", configureMessage.Sender)
			return
		}
	} else {
		err := json.Unmarshal([]byte(payload), &configureMessage)
		if err != nil {
			publisher.logger.Infof("Unable to unmarshal ConfigureMessage in %s", address)
			return
		}
	}

	// TODO: authorization check
	node := publisher.Nodes.GetNodeByAddress(address)
	if node == nil || message == "" {
		publisher.logger.Infof("handleNodeConfig unknown node for address %s or missing message", address)
		return
	}

	// Verify that the message comes from the sender using the sender's public key
	// isValid := publisher.VerifyMessageSignature(configureMessage.Sender, message, publication.Signature)
	// if !isValid {
	// 	publisher.Logger.Warningf("Incoming configuration verification failed for sender: %s", configureMessage.Sender)
	// 	return
	// }
	params := configureMessage.Attr
	if publisher.onNodeConfigHandler != nil {
		// A handler can filter which configuration updates take place
		params = publisher.onNodeConfigHandler(node, params)
	}
	// process the requested configuration, or ignore if none are applicable
	if params != nil {
		publisher.Nodes.SetNodeConfigValues(address, params)
	}
}
