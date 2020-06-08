// Package publisher with handling of configuration commands
package publisher

import (
	"encoding/json"

	"github.com/hspaay/iotc.golang/iotc"
)

// handle an incoming a configuration command for one of our nodes. This:
// - check if the signature is valid
// - check if the node is valid
// - pass the configuration update to the adapter's callback set in Start()
// - save node configuration if persistence is set
// TODO: support for authorization per node
func (publisher *Publisher) handleNodeConfigCommand(address string, message []byte) {
	publisher.logger.Infof("handleNodeConfig on address %s", address)
	// TODO: authorization check
	node := publisher.Nodes.GetNodeByAddress(address)
	if node == nil || message == nil {
		publisher.logger.Infof("handleNodeConfig unknown node for address %s or missing message", address)
		return
	}
	payload, err := publisher.signer.Verify(message)
	if err != nil {
		publisher.logger.Warningf("handleNodeConfig Invalid message: %s", err)
		return
	}

	var configureMessage iotc.NodeConfigureMessage
	err = json.Unmarshal(payload, &configureMessage)
	if err != nil {
		publisher.logger.Infof("Unable to unmarshal ConfigureMessage in %s", address)
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
