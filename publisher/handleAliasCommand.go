// Package publisher with handling of setting a node's alias
package publisher

import (
	"github.com/iotdomain/iotdomain-go/types"
)

// HandleAliasCommand handles the command to set the alias of a node. This updates the address
// of a node, its inputs and its outputs.
func (publisher *Publisher) HandleAliasCommand(address string, message *types.NodeAliasMessage) {
	node := publisher.registeredNodes.GetNodeByAddress(address)
	if node == nil {
		return
	}
	nodeID := node.NodeID
	publisher.registeredNodes.SetAlias(nodeID, message.Alias)
	publisher.registeredInputs.SetAlias(nodeID, message.Alias)
	publisher.registeredOutputs.SetAlias(nodeID, message.Alias)
}
