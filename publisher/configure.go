// Package publisher handling configuration of my nodes
// - Update node configuration as it is discovered by the publisher, or one of its nodes
// - Handle incoming node configuration command
//
// FIXME: Differentiate between node and service (adapter?) configuration
//        node configuration are applied to a device and service and do not apply until
//         the device accepted the configuration. Examples are calibration, report type
//         reported unit, reporting inteval, min/max limits for alerting.
//        service configuration relate to the managing the node and include attrs like
//         name, alias, keys, poll interval, enable/disable
//
// TODO: support for authorization per node
// Not thread-safe.
package publisher

import (
	"encoding/json"

	"github.com/hspaay/iotconnect.golang/messenger"
	"github.com/hspaay/iotconnect.golang/standard"
)

// handle an incoming a configuration command for one of our nodes. This:
// - check if the signature is valid
// - check if the node is valid
// - pass the configuration update to the adapter's callback set in Start()
// TODO: support for authorization per node
func (publisher *PublisherState) handleNodeConfigCommand(address string, publication *messenger.Publication) {
	// TODO: authorization check
	node := publisher.Nodes.GetNodeByAddress(address)
	if node == nil || publication.Message == nil {
		publisher.Logger.Infof("handleNodeConfig unknown node for address %s or missing message", address)
		return
	}
	var configureMessage standard.ConfigureMessage
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
	params := configureMessage.Config
	if publisher.onNodeConfigHandler != nil {
		// config handler returns the parameters that must be applied directly
		params = publisher.onNodeConfigHandler(node, params)
	}
	// process the requested configuration
	publisher.Nodes.UpdateNodeConfigValues(address, params)
}
