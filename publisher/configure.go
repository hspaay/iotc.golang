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

// DiscoverNodeConfig is called by the adapter to add or update a configuration attribute
// that was reported by the node.
// config struct with configuration description and value
func (publisher *PublisherState) DiscoverNodeConfig(
	node *standard.Node, config *standard.ConfigAttr) {

	publisher.Logger.Info("DiscoverNodeConfig node: ", node.Address)

	publisher.updateMutex.Lock()
	node.Config[config.ID] = config
	if publisher.updatedNodes == nil {
		publisher.updatedNodes = make(map[string]*standard.Node)
	}
	publisher.updatedNodes[node.Address] = node

	if publisher.synchroneous {
		publisher.publishDiscovery()
	}
	publisher.updateMutex.Unlock()
}

// UpdateNodeConfigValue applies an update to a node's existing configuration
// Intended for use by the handler that receives a $configure command. The handler
// must only apply configuration updates that are not handled by the node, like for example
// the name.
func (publisher *PublisherState) UpdateNodeConfigValue(address string, param map[string]string) {
	node := publisher.GetNodeByAddress(address)

	var appliedParams map[string]string = param
	for key, value := range appliedParams {
		config := node.Config[key]
		if config == nil {
			config = &standard.ConfigAttr{}
			// FIXME: this is not thread-safe
			node.Config[key] = config
		}
		config.Value = value
	}
	// re-discover the node for publication
	publisher.DiscoverNode(node)
}

// handle an incoming a configuration command for one of our nodes. This:
// - check if the signature is valid
// - check if the node is valid
// - pass the configuration update to the adapter's callback set in Start()
// TODO: support for authorization per node
func (publisher *PublisherState) handleNodeConfigCommand(address string, publication *messenger.Publication) {
	// TODO: authorization check
	node := publisher.GetNodeByAddress(address)
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
	if publisher.onConfig != nil {
		// config handler returnst the parameters that must be applied directly
		params = publisher.onConfig(node, params)
	}
	// process the requested configuration
	publisher.UpdateNodeConfigValue(address, params)
}
