// Package nodes with node management
package nodes

import (
	"fmt"
	"strings"
	"sync"

	"github.com/hspaay/iotc.golang/iotc"
)

// NodeList for concurrency safe node management using Copy on Write.
//  To serialize the node list use GetAllNodes and UpdateNodes
// Nodes are immutable. Any modifications made are applied to a new instance. The old node instance
// is discarded and replaced with the new instance.
// To make changes to a node directly, always Clone the node first and use UpdateNode to apply the change.
type NodeList struct {
	// don't access directly. This is only accessible for serialization
	nodeMap      map[string]*Node
	updateMutex  *sync.Mutex      // mutex for async updating of nodes
	updatedNodes map[string]*Node // nodes by address that have been rediscovered/updated since last publication
}

// GetAllNodes returns a list of nodes
func (nodes *NodeList) GetAllNodes() []*Node {
	nodes.updateMutex.Lock()
	defer nodes.updateMutex.Unlock()

	var nodeList = make([]*Node, 0)
	for _, node := range nodes.nodeMap {
		nodeList = append(nodeList, node)
	}
	return nodeList
}

// GetNodeByAddress returns a node by its node address using the zone, publisherID and nodeID
// address must contain the zone, publisher and nodeID. Any other fields are ignored.
// Returns nil if address has no known node
func (nodes *NodeList) GetNodeByAddress(address string) *Node {
	nodes.updateMutex.Lock()
	defer nodes.updateMutex.Unlock()

	var node = nodes.getNode(address)
	return node
}

// GetNodeByID returns a node by its zone, publisher and node ID
// Returns nil if address has no known node
func (nodes *NodeList) GetNodeByID(zone string, publisherID string, nodeID string) *Node {
	nodeAddr := fmt.Sprintf("%s/%s/%s/%s", zone, publisherID, nodeID, iotc.MessageTypeNodeDiscovery)

	nodes.updateMutex.Lock()
	defer nodes.updateMutex.Unlock()

	var node = nodes.nodeMap[nodeAddr]
	return node
}

// GetUpdatedNodes returns the list of nodes that have been updated
// clearUpdates clears the list of updates. Intended for publishing only updated nodes.
func (nodes *NodeList) GetUpdatedNodes(clearUpdates bool) []*Node {
	var updateList []*Node = make([]*Node, 0)

	nodes.updateMutex.Lock()
	defer nodes.updateMutex.Unlock()

	if nodes.updatedNodes != nil {
		for _, node := range nodes.updatedNodes {
			updateList = append(updateList, node)
		}
		if clearUpdates {
			nodes.updatedNodes = nil
		}
	}
	return updateList
}

// SetErrorStatus sets the node RunState to error with a message in the node status NodeStateLastError
// Use SetRunState to clear the runstate.
func (nodes *NodeList) SetErrorStatus(node *Node, errorMsg string) (changed bool) {
	if node != nil {
		// newNode.SetErrorState(errorMsg)
		statusUpdate := map[iotc.NodeStatus]string{
			iotc.NodeStatusLastError: errorMsg,
		}
		changed = nodes.SetNodeStatus(node, statusUpdate)

		if node.RunState != iotc.NodeRunStateError {
			nodes.updateMutex.Lock()
			defer nodes.updateMutex.Unlock()
			newNode := node.Clone()

			changed = true
			newNode.RunState = iotc.NodeRunStateError
			nodes.updateNode(newNode)
		}
	}
	return changed
}

// SetNodeAttr updates node's attributes and publishes the updated node.
// Node is marked as modified for publication only if one of the attrParams has changes
// Use when additional node attributes has been discovered.
// - address of the node to update
// - param is the map with key-value pairs of attribute values to update
func (nodes *NodeList) SetNodeAttr(address string, attrParams map[iotc.NodeAttr]string) {
	nodes.updateMutex.Lock()
	defer nodes.updateMutex.Unlock()
	node := nodes.getNode(address)
	newNode := node.Clone()
	changed := newNode.SetNodeAttr(attrParams)
	if changed {
		nodes.updateNode(newNode)
	}
}

// SetNodeConfig updates a node's configuration and publishes the updated node
// Nodes are immutable. A new node is created and published and the old node instance is discarded.
// Use this when a device configuration has been identified, or when the config value updates.
// - node is the node to update
// - config is the config struct with description and value
// Returns a new node instance
func (nodes *NodeList) SetNodeConfig(address string, configAttr *iotc.ConfigAttr) {
	nodes.updateMutex.Lock()
	defer nodes.updateMutex.Unlock()
	node := nodes.getNode(address)
	if node == nil {
		return
	}
	newNode := node.Clone()
	newNode.Config[configAttr.ID] = *configAttr
	nodes.updateNode(newNode)
}

// SetNodeRunState updates the node's runstate status
func (nodes *NodeList) SetNodeRunState(node *Node, runState iotc.NodeRunState) {
	nodes.updateMutex.Lock()
	defer nodes.updateMutex.Unlock()
	// node := nodes.getNode(address)
	if node == nil {
		return
	}
	changed := (node.RunState != runState)
	if changed {
		newNode := node.Clone()
		newNode.RunState = runState
		nodes.updateNode(newNode)
	}
}

// // SetNodeStatus updates one or more node's status attributes
// // Nodes are immutable. If one or more status values have changed then a new node is created and
// // published. The old node instance is discarded.
// // - address of the node to update
// // - param is the map with key-value pairs of node status
// func (nodes *NodeList) SetNodeStatus(address string, attrParams map[iotc.NodeStatus]string) {
// 	nodes.updateMutex.Lock()
// 	defer nodes.updateMutex.Unlock()
// 	node := nodes.getNode(address)
// 	if node == nil {
// 		return
// 	}
// 	newNode := node.Clone()
// 	changed := false
// 	for key, value := range attrParams {
// 		if newNode.Status[key] != value {
// 			newNode.Status[key] = value
// 			changed = true
// 		}
// 	}

// 	if changed {
// 		nodes.updateNode(newNode)
// 	}
// }

// SetNodeStatus updates one or more node's status attributes
// Nodes are immutable. If one or more status values have changed then a new node is created and
// published. The old node instance is discarded.
// - address of the node to update
// - statusAttr is the map with key-value pairs of updated node statusses
func (nodes *NodeList) SetNodeStatus(node *Node, statusAttr map[iotc.NodeStatus]string) (changed bool) {
	nodes.updateMutex.Lock()
	defer nodes.updateMutex.Unlock()
	// node := nodes.getNode(address)
	// if node == nil {
	// 	return
	// }
	newNode := node.Clone()
	changed = false
	for key, value := range statusAttr {
		if newNode.Status[key] != value {
			newNode.Status[key] = value
			changed = true
		}
	}

	if changed {
		nodes.updateNode(newNode)
	}
	return changed
}

// SetNodeConfigValues applies an update to a node's existing configuration
// Nodes are immutable. If one or more configuration values have changed then a new node is created and
// published and the old node instance is discarded.
// - address is the node discovery address
// - param is the map with key-value pairs of configuration values to update
func (nodes *NodeList) SetNodeConfigValues(address string, param map[iotc.NodeAttr]string) {
	nodes.updateMutex.Lock()
	defer nodes.updateMutex.Unlock()

	node := nodes.getNode(address)
	if node == nil {
		return
	}
	newNode := node.Clone()
	changed := newNode.SetNodeConfigValues(param)
	if changed {
		nodes.updateNode(newNode)
	}
}

// UpdateNode replaces a node or adds a new node based on node.Address
// Intended to support Node immutability by making changes to a copy of a node and replacing
// the existing node with the updated node
// The updated node will be published
func (nodes *NodeList) UpdateNode(node *Node) {
	nodes.updateMutex.Lock()
	defer nodes.updateMutex.Unlock()
	nodes.updateNode(node)
}

// UpdateNodes updates a list of nodes
// Intended to update the list with nodes from persistent storage
func (nodes *NodeList) UpdateNodes(updates []*Node) {
	nodes.updateMutex.Lock()
	defer nodes.updateMutex.Unlock()

	for _, node := range updates {
		// fill in missing fields
		if node.Attr == nil {
			node.Attr = map[iotc.NodeAttr]string{}
		}
		if node.Config == nil {
			node.Config = map[iotc.NodeAttr]iotc.ConfigAttr{}
		}
		if node.Status == nil {
			node.Status = make(map[iotc.NodeStatus]string)
		}
		nodes.updateNode(node)
	}
}

// getNode returns a node by its node address using the zone, publisherID and nodeID
// address must contain the zone, publisher and nodeID. Any other fields are ignored.
// Intended for use within a locked section for updating, eg lock - read - update - write - unlock
// Returns nil if address has no known node
func (nodes *NodeList) getNode(address string) *Node {
	segments := strings.Split(address, "/")
	if len(segments) < 3 {
		return nil
	}
	segments[3] = iotc.MessageTypeNodeDiscovery
	nodeAddr := strings.Join(segments[:4], "/")
	var node = nodes.nodeMap[nodeAddr]
	return node
}

// updateNode replaces a node and adds it to the list of updated nodes
// Intended for use within a locked section
func (nodes *NodeList) updateNode(node *Node) {
	nodes.nodeMap[node.Address] = node
	if nodes.updatedNodes == nil {
		nodes.updatedNodes = make(map[string]*Node)
	}
	nodes.updatedNodes[node.Address] = node
}

// NewNodeList creates a new instance for node management
func NewNodeList() *NodeList {
	nodes := NodeList{
		nodeMap:     make(map[string]*Node),
		updateMutex: &sync.Mutex{},
	}
	return &nodes
}
