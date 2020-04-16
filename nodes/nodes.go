// Package nodes with node management
package nodes

import (
	"fmt"
	"strings"
	"sync"

	"github.com/hspaay/iotconnect.golang/standard"
)

// NodeList with node management
// To facilitate concurrent read-write, nodes are treated as immutable.
// Any modifications made to a node must be made to a copy using CloneNode and stored
// using UpdateNode.
type NodeList struct {
	nodeMap      map[string]*standard.Node
	updateMutex  *sync.Mutex               // mutex for async updating of nodes
	updatedNodes map[string]*standard.Node // nodes by address that have been rediscovered/updated since last publication
}

// CloneNode returns a copy of the node with new Attr, Config and Status maps
// Intended for updating the node in a concurrency safe manner in combination with UpdateNode()
// This does not perform a deep copy of the  maps. Any updates to the map must use new instances of the values
func CloneNode(node *standard.Node) *standard.Node {
	newNode := *node

	newNode.Attr = make(standard.AttrMap)
	for key, value := range node.Attr {
		newNode.Attr[key] = value
	}
	newNode.Config = make(standard.ConfigAttrMap)
	for key, value := range node.Config {
		newNode.Config[key] = value
	}
	newNode.Status = make(standard.AttrMap)
	for key, value := range node.Status {
		newNode.Status[key] = value
	}
	return &newNode
}

// GetAllNodes returns a list of nodes
// This method is concurrent safe
func (nodes *NodeList) GetAllNodes() []*standard.Node {
	nodes.updateMutex.Lock()
	var nodeList = make([]*standard.Node, 0)
	for _, node := range nodes.nodeMap {
		nodeList = append(nodeList, node)
	}
	nodes.updateMutex.Unlock()
	return nodeList
}

// GetNodeAlias returns the node alias or node ID if no alias is set
func (nodes *NodeList) GetNodeAlias(node *standard.Node) (alias string, hasAlias bool) {
	hasAlias = false
	alias = node.ID
	aliasConfig, attrExists := node.Config[standard.AttrNameAlias]
	if attrExists && aliasConfig.Value != "" {
		alias = aliasConfig.Value
	}
	return alias, hasAlias
}

// GetNodeByAddress returns a node by its node address using the zone, publisherID and nodeID
// address must contain the zone, publisher and nodeID. Any other fields are ignored.
// Returns nil if address has no known node
// This method is concurrent safe
func (nodes *NodeList) GetNodeByAddress(address string) *standard.Node {
	nodes.updateMutex.Lock()
	var node = nodes.getNode(address)
	nodes.updateMutex.Unlock()
	return node
}

// GetNodeByID returns a node by its zone, publisher and node ID
// Returns nil if address has no known node
// This method is concurrent safe
func (nodes *NodeList) GetNodeByID(zone string, publisherID string, nodeID string) *standard.Node {
	nodeAddr := fmt.Sprintf("%s/%s/%s/%s", zone, publisherID, nodeID, standard.CommandNodeDiscovery)

	nodes.updateMutex.Lock()
	var node = nodes.nodeMap[nodeAddr]
	nodes.updateMutex.Unlock()
	return node
}

// GetUpdatedNodes returns the list of discovered nodes that have been updated
// clear the update on return
func (nodes *NodeList) GetUpdatedNodes(clearUpdates bool) []*standard.Node {
	var updateList []*standard.Node = make([]*standard.Node, 0)

	nodes.updateMutex.Lock()
	if nodes.updatedNodes != nil {
		for _, node := range nodes.updatedNodes {
			updateList = append(updateList, node)
		}
		if clearUpdates {
			nodes.updatedNodes = nil
		}
	}
	nodes.updateMutex.Unlock()
	return updateList
}

// UpdateNode replaces a node or adds a new node based on node.Address
// node is a new instance which will replace the existing instance if it exists or adds it if it
// doesn't exist based on the node.Address. The node is also added to the list of updated nodes.
func (nodes *NodeList) UpdateNode(node *standard.Node) {
	nodes.updateMutex.Lock()
	nodes.updateNode(node)
	nodes.updateMutex.Unlock()
}

// UpdateNodeAttr updates a node's attributes and publishes the updated node.
// Use when additional node attributes has been discovered.
// - node is the node to update
// - param is the map with key-value pairs of attribute values to update
// Returns a new node instance
func (nodes *NodeList) UpdateNodeAttr(address string, attrParams map[string]string) *standard.Node {
	nodes.updateMutex.Lock()
	node := nodes.getNode(address)
	newNode := CloneNode(node)

	for key, value := range attrParams {
		newNode.Attr[key] = value
	}
	nodes.updateNode(newNode)
	nodes.updateMutex.Unlock()
	return newNode
}

// UpdateNodeConfig updates a node's configuration and publishes the updated node
// Use this when a device configuration has been identified, or when the config value updates.
// - node is the node to update
// - config is the config struct with description and value
// Returns a new node instance
func (nodes *NodeList) UpdateNodeConfig(address string, config *standard.ConfigAttr) *standard.Node {
	nodes.updateMutex.Lock()
	node := nodes.getNode(address)
	newNode := CloneNode(node)

	newNode.Config[config.ID] = *config

	nodes.updateNode(newNode)
	nodes.updateMutex.Unlock()
	return newNode
}

// UpdateNodeConfigValues applies an update to a node's existing configuration and publish the updated node.
// To be called by the configuration update handler after receiving a $configure command. The handler
// must only apply configuration updates that can be applied directly and have been validated.
// Some configuration, like for example ZWave device configuration, must be set to the device first and
// not through this function. When the device accepts the configuration, use UpdateNodeConfig to apply it.
// This function is concurrency safe. A new node instance is created with the new config. The original
// node remains unchanged.
// - address is the node discovery address
// - param is the map with key-value pairs of configuration values to update
func (nodes *NodeList) UpdateNodeConfigValues(address string, param map[string]string) *standard.Node {
	nodes.updateMutex.Lock()
	node := nodes.getNode(address)
	newNode := CloneNode(node)

	var appliedParams map[string]string = param
	for key, value := range appliedParams {
		config, configExists := node.Config[key]
		if !configExists {
			newConfig := standard.ConfigAttr{Value: value}
			newNode.Config[key] = newConfig
		} else {
			newConfig := config // shallow copy of config before changing the value
			newConfig.Value = value
			newNode.Config[key] = newConfig
		}
	}
	nodes.updateNode(newNode)
	nodes.updateMutex.Unlock()
	return newNode
}

// UpdateNodeStatus updates a node's status attribute.
// - node is the node to update
// - param is the map with key-value pairs of status values to update
// Returns a new node instance
func (nodes *NodeList) UpdateNodeStatus(address string, statusParams map[string]string) *standard.Node {
	// Nodes are immutable
	nodes.updateMutex.Lock()
	node := nodes.getNode(address)
	newNode := CloneNode(node)
	for key, value := range statusParams {
		newNode.Status[key] = value
	}
	nodes.updateNode(newNode)
	nodes.updateMutex.Unlock()
	return newNode
}

// getNode returns a node by its node address using the zone, publisherID and nodeID
// address must contain the zone, publisher and nodeID. Any other fields are ignored.
// Intended for use within a locked section for updating, eg lock - read - update - write - unlock
// Returns nil if address has no known node
func (nodes *NodeList) getNode(address string) *standard.Node {
	segments := strings.Split(address, "/")
	if len(segments) < 3 {
		return nil
	}
	segments[3] = standard.CommandNodeDiscovery
	nodeAddr := strings.Join(segments[:4], "/")
	var node = nodes.nodeMap[nodeAddr]
	return node
}

// updateNode replaces a node and adds it to the list of updated nodes
// Intended for use within a locked section
func (nodes *NodeList) updateNode(node *standard.Node) {
	nodes.nodeMap[node.Address] = node
	if nodes.updatedNodes == nil {
		nodes.updatedNodes = make(map[string]*standard.Node)
	}
	nodes.updatedNodes[node.Address] = node
}

// NewNodeList creates a new instance for node management
func NewNodeList() *NodeList {
	nodes := NodeList{
		nodeMap:     make(map[string]*standard.Node),
		updateMutex: &sync.Mutex{},
	}
	return &nodes
}
