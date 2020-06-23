// Package publisher with handling of configuration commands
package publisher

import (
	"github.com/iotdomain/iotdomain-go/messenger"
	"github.com/iotdomain/iotdomain-go/types"
)

// handle an incoming a configuration command for one of our nodes. This:
// - check if the signature is valid
// - check if the node is valid
// - pass the configuration update to the adapter's callback set in Start()
// - save node configuration if persistence is set
// TODO: support for authorization per node
func (publisher *Publisher) handleNodeConfigCommand(address string, message string) {
	var configureMessage types.NodeConfigureMessage

	// Expect the message to be encrypted
	isEncrypted, dmessage, err := messenger.DecryptMessage(message, publisher.identityPrivateKey)
	if !isEncrypted {
		publisher.logger.Infof("handleNodeConfigCommand: message to '%s' is not encrypted.", address)
		// TODO: determine if encryption is required
		// this could be fine for now, just warning
	} else if err != nil {
		publisher.logger.Warnf("handleNodeConfigCommand: decryption failed of message to '%s'. Message discarded.", address)
		return
	}

	// Verify the message using the public key of the sender
	isSigned, err := messenger.VerifySender(dmessage, &configureMessage, publisher.domainPublishers.GetPublisherKey)
	if !isSigned {
		// all configuration commands must use signed messages
		publisher.logger.Warnf("handleNodeConfigCommand: message to input '%s' is not signed. Message discarded.", address)
		return
	} else if err != nil {
		// signing failed, discard the message
		publisher.logger.Warnf("handleNodeConfigCommand: signature verification failed for message to input %s. Message discarded.", address)
		return
	}

	// TODO: authorization check
	node := publisher.Nodes.GetNodeByAddress(address)
	if node == nil || message == "" {
		publisher.logger.Infof("handleNodeConfig unknown node for address %s or missing message", address)
		return
	}
	publisher.logger.Infof("handleNodeConfig configure command on address %s. isEncrypted=%t, isSigned=%t", address, isEncrypted, isSigned)

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
