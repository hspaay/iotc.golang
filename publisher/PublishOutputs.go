package publisher

import (
	"strings"

	"github.com/iotdomain/iotdomain-go/persist"
	"github.com/iotdomain/iotdomain-go/types"
)

// PublishUpdatedOutputs publishes pending updates to discovered outputs
func (publisher *Publisher) PublishUpdatedOutputs() {
	updatedOutputs := publisher.Outputs.GetUpdatedOutputs(true)

	// publish updated output discovery
	for _, output := range updatedOutputs {
		aliasAddress := publisher.getOutputAliasAddress(output.Address, "")
		publisher.logger.Infof("Publisher.PublishUpdates: publish output discovery: %s", aliasAddress)
		publisher.publishObject(aliasAddress, true, output, nil)
	}
	if len(updatedOutputs) > 0 && publisher.cacheFolder != "" {
		allOutputs := publisher.Outputs.GetAllOutputs()
		persist.SaveOutputs(publisher.cacheFolder, publisher.PublisherID(), allOutputs)
	}
}

// Replace the address with the node's alias instead the node ID, and the message type with the given
//  message type for publication.
// If the node doesn't have an alias then its nodeId will be kept.
// messageType to substitute in the address. Use "" to keep the original message type (usually discovery message)
func (publisher *Publisher) getOutputAliasAddress(address string, messageType types.MessageType) string {
	node := publisher.Nodes.GetNodeByAddress(address)
	if node == nil {
		return address
	}
	alias, _ := publisher.Nodes.GetNodeConfigString(address, types.NodeAttrAlias, "")
	// alias, hasAlias := nodes.GetNodeAlias(node)
	// zone/pub/node/outtype/instance/messagetype
	parts := strings.Split(address, "/")
	if alias == "" {
		alias = parts[2]
	}
	parts[2] = alias
	if messageType != "" {
		parts[5] = string(messageType)
	}
	aliasAddr := strings.Join(parts, "/")
	return aliasAddr
}
