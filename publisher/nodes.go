// Package publisher with node management functions
package publisher

import (
	"fmt"
	"strings"

	"github.com/hspaay/iotconnect.golang/standard"
)

// GetAllNodes returns a list of nodes from this publisher
// This method is concurrent safe
func (publisher *PublisherState) GetAllNodes() []*standard.Node {
	publisher.updateMutex.Lock()
	var nodeList = make([]*standard.Node, 0)
	for _, node := range publisher.nodes {
		nodeList = append(nodeList, node)
	}
	publisher.updateMutex.Unlock()
	return nodeList
}

// GetNodeAlias returns the node alias or node ID if no alias is set
func (publisher *PublisherState) GetNodeAlias(node *standard.Node) (alias string, hasAlias bool) {
	hasAlias = false
	alias = node.ID
	aliasConfig := node.Config[standard.AttrNameAlias]
	if aliasConfig != nil && aliasConfig.Value != "" {
		alias = aliasConfig.Value
	}
	return alias, hasAlias
}

// GetNodeByAddress returns a node from this publisher by the node address
// address of the node, only the zone, publisher and nodeID are used. Any command suffix is ignored
// Returns nil if address has no known node
// This method is concurrent safe
func (publisher *PublisherState) GetNodeByAddress(address string) *standard.Node {
	segments := strings.Split(address, "/")
	if len(segments) < 3 {
		return nil
	}
	segments[3] = standard.CommandNodeDiscovery
	nodeAddr := strings.Join(segments[:4], "/")
	publisher.updateMutex.Lock()
	var node = publisher.nodes[nodeAddr]
	publisher.updateMutex.Unlock()
	return node
}

// GetNodeByID returns a discovered node from this publisher by its node ID
// Returns nil if address has no known node
// This method is concurrent safe
func (publisher *PublisherState) GetNodeByID(id string) *standard.Node {
	nodeAddr := fmt.Sprintf("%s/%s/%s/%s", publisher.zone, publisher.publisherID, id, standard.CommandNodeDiscovery)

	publisher.updateMutex.Lock()
	var node = publisher.nodes[nodeAddr]
	publisher.updateMutex.Unlock()
	return node
}

// getNodeOutputs returns a list of outputs for the given node
// This method is not concurrent safe and should only be used in a locked section
func (publisher *PublisherState) getNodeOutputs(node *standard.Node) []*standard.InOutput {
	outputs := []*standard.InOutput{}
	for _, output := range publisher.outputs {
		if output.NodeID == node.ID {
			outputs = append(outputs, output)
		}
	}
	return outputs
}

// getNode returns a discovered node by the node address
// address of the node, only the zone, publisher and nodeID are used. Any command suffix is ignored
// This method is not concurrent safe and should only be used in a locked section
// Returns nil if address has no known node
func (publisher *PublisherState) getNode(address string) *standard.Node {
	segments := strings.Split(address, "/")
	if len(segments) < 3 {
		return nil
	}
	segments[3] = standard.CommandNodeDiscovery
	nodeAddr := strings.Join(segments[:4], "/")

	var node = publisher.nodes[nodeAddr]
	return node
}
