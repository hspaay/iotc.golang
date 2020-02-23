// Package publisher handling node configuration
// - Update node configuration as it is discovered by the publisher, or one of its nodes
// - Handle incoming node configuration command
// -
// -
// Not thread-safe.
package publisher

import "myzone/nodes"

// ConfigureMessage with configuration parameters
type ConfigureMessage struct {
	Address   string        `json:"address"` // zone/publisher/node/$configure
	Config    nodes.AttrMap `json:"config"`
	Sender    string        `json:"sender"`
	Timestamp string        `json:"timestamp"`
}

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
// - pass the configuration update to the adapter's callback
//
// The message has already be unmarshalled and the signature verified by the messenger.
// address of the node to be configured
// message object with the configuration
func (publisher *ThisPublisherState) handleNodeConfig(address string, message interface{}) {
	// TODO: authorization check
	node := publisher.GetNode(address)
	if node == nil || message == nil {
		publisher.Logger.Infof("handleNodeConfig unknown node for address %s or missing message", address)
		return
	}
	configMessage := message.(ConfigureMessage)
	params := configMessage.Config
	if publisher.onConfig != nil {
		publisher.onConfig(node, params)
	} else {
		// process the requested configuration
		publisher.UpdateNodeConfig(address, params)
	}
}
