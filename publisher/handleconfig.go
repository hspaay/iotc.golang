// Package publisher with handling of configuration commands
package publisher

import (
	"encoding/json"

	"github.com/hspaay/iotconnect.golang/messaging"
)

// handle an incoming a configuration command for one of our nodes. This:
// - check if the signature is valid
// - check if the node is valid
// - pass the configuration update to the adapter's callback set in Start()
// TODO: support for authorization per node
func (publisher *Publisher) handleNodeConfigCommand(address string, publication *messaging.Publication) {
	publisher.Logger.Infof("handleNodeConfig on address %s", address)
	// TODO: authorization check
	node := publisher.Nodes.GetNodeByAddress(address)
	if node == nil || publication.Message == nil {
		publisher.Logger.Infof("handleNodeConfig unknown node for address %s or missing message", address)
		return
	}
	var configureMessage messaging.NodeConfigureMessage
	err := json.Unmarshal([]byte(publication.Message), &configureMessage)
	if err != nil {
		publisher.Logger.Infof("Unable to unmarshal ConfigureMessage in %s", address)
		return
	}
	// Verify that the message comes from the sender using the sender's public key
	isValid := publisher.VerifyMessageSignature(configureMessage.Sender, publication.Message, publication.Signature)
	if !isValid {
		publisher.Logger.Warningf("Incoming configuration verification failed for sender: %s", configureMessage.Sender)
		return
	}
	params := configureMessage.Attr
	if publisher.onNodeConfigHandler != nil {
		// A handler can filter which configuration updates take place
		params = publisher.onNodeConfigHandler(node, params)
	}
	// process the requested configuration
	publisher.Nodes.SetNodeConfigValues(address, params)
}
