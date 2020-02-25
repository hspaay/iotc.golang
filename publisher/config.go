// Package publisher handling node configuration
// - Update node configuration as it is discovered by the publisher, or one of its nodes
// - Handle incoming node configuration command
// -
// -
// Not thread-safe.
package publisher

import (
	"encoding/json"
	"iotzone/messenger"
	"iotzone/nodes"
)

// DiscoverNodeConfig adds configuration to the node
// node whose config has been discovered
// name of config, unique for the node
// config struct with configuration description and value
func (publisher *ThisPublisherState) DiscoverNodeConfig(
	node *nodes.Node, name string, config *nodes.ConfigAttr) {

	publisher.Logger.Info("DiscoverNodeConfig node: ", node.Address)

	publisher.updateMutex.Lock()
	node.Config[name] = config
	if publisher.updatedNodes == nil {
		publisher.updatedNodes = make(map[string]*nodes.Node)
	}
	publisher.updatedNodes[node.Address] = node

	if publisher.synchroneous {
		publisher.publishUpdates()
	}
	publisher.updateMutex.Unlock()
}

// UpdateNodeConfig updates a node's existing configuration
// Called when receiving a configuration update, after it has been processed by the adapter handler.
// Configuration updates that are send directly to the node should not be included as they only take
// effect after the node confirms that its configuration has changed, eg zwave callback.
// This will re-publish the node discovery.
func (publisher *ThisPublisherState) UpdateNodeConfig(address string, param map[string]string) {
	node := publisher.GetNode(address)

	var appliedParams map[string]string = param
	for key, value := range appliedParams {
		config := node.Config[key]
		if config == nil {
			config = &nodes.ConfigAttr{}
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
//
func (publisher *ThisPublisherState) handleNodeConfig(address string, publication *messenger.Publication) {
	// TODO: authorization check
	node := publisher.GetNode(address)
	if node == nil || publication.Message == "" {
		publisher.Logger.Infof("handleNodeConfig unknown node for address %s or missing message", address)
		return
	}
	var configureMessage ConfigureMessage
	err := json.Unmarshal([]byte(publication.Message), &configureMessage)
	if err != nil {
		publisher.Logger.Infof("Unable to unmarshal ConfigureMessage in %s", address)
		return
	}
	params := configureMessage.Config
	if publisher.onConfig != nil {
		// config handler returnst the parameters that must be applied directly
		params = publisher.onConfig(node, params)
	}
	// process the requested configuration
	publisher.UpdateNodeConfig(address, params)
}
