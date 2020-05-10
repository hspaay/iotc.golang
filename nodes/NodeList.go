// Package nodes with node management
package nodes

import (
	"errors"
	"fmt"
	"strconv"
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
	nodeMap      map[string]*iotc.NodeDiscoveryMessage
	updateMutex  *sync.Mutex                           // mutex for async updating of nodes
	updatedNodes map[string]*iotc.NodeDiscoveryMessage // nodes by address that have been rediscovered/updated since last publication
}

// Clone returns a copy of the node with new Attr, Config and Status maps
// Intended for updating the node in a concurrent safe manner in combination with UpdateNode()
// This does clones map values. Any updates to the map must use new instances of the values
func (nodes *NodeList) Clone(node *iotc.NodeDiscoveryMessage) *iotc.NodeDiscoveryMessage {
	newNode := *node

	newNode.Attr = make(map[iotc.NodeAttr]string)
	for key, value := range node.Attr {
		newNode.Attr[key] = value
	}
	newNode.Config = make(map[iotc.NodeAttr]iotc.ConfigAttr)
	for key, value := range node.Config {
		newNode.Config[key] = value
	}
	newNode.Status = make(map[iotc.NodeStatus]string)
	for key, value := range node.Status {
		newNode.Status[key] = value
	}
	return &newNode
}

// GetAllNodes returns a list of nodes
func (nodes *NodeList) GetAllNodes() []*iotc.NodeDiscoveryMessage {
	nodes.updateMutex.Lock()
	defer nodes.updateMutex.Unlock()

	var nodeList = make([]*iotc.NodeDiscoveryMessage, 0)
	for _, node := range nodes.nodeMap {
		nodeList = append(nodeList, node)
	}
	return nodeList
}

// GetNodeByAddress returns a node by its node address using the zone, publisherID and nodeID
// address must contain the zone, publisher and nodeID. Any other fields are ignored.
// Returns nil if address has no known node
func (nodes *NodeList) GetNodeByAddress(address string) *iotc.NodeDiscoveryMessage {
	nodes.updateMutex.Lock()
	defer nodes.updateMutex.Unlock()

	var node = nodes.getNode(address)
	return node
}

// GetNodeConfigInt returns the node configuration value as an integer
// address starts with the node's address
// This retuns the 'default' value if no value is set
func (nodes *NodeList) GetNodeConfigInt(address string, attrName iotc.NodeAttr) (value int, err error) {
	valueStr, configExists := nodes.GetNodeConfigValue(address, attrName)
	if !configExists {
		return 0, errors.New("Configuration does not exist")
	}
	return strconv.Atoi(valueStr)
}

// GetNodeConfigValue returns the configuration value of a node in this list
// address starts with the node's address
// This retuns the 'default' value if no value is set
func (nodes *NodeList) GetNodeConfigValue(address string, attrName iotc.NodeAttr) (value string, configExists bool) {
	nodes.updateMutex.Lock()
	defer nodes.updateMutex.Unlock()

	node := nodes.getNode(address)
	if node == nil {
		return "", false
	}
	config, configExists := node.Config[attrName]
	if !configExists {
		return "", configExists
	}
	if config.Value == "" {
		return config.Default, configExists
	}
	return config.Value, configExists
}

// GetNodeByID returns a node by its zone, publisher and node ID
// Returns nil if address has no known node
func (nodes *NodeList) GetNodeByID(zone string, publisherID string, nodeID string) *iotc.NodeDiscoveryMessage {
	nodeAddr := fmt.Sprintf("%s/%s/%s/%s", zone, publisherID, nodeID, iotc.MessageTypeNodeDiscovery)

	nodes.updateMutex.Lock()
	defer nodes.updateMutex.Unlock()

	var node = nodes.nodeMap[nodeAddr]
	return node
}

// GetUpdatedNodes returns the list of nodes that have been updated
// clearUpdates clears the list of updates. Intended for publishing only updated nodes.
func (nodes *NodeList) GetUpdatedNodes(clearUpdates bool) []*iotc.NodeDiscoveryMessage {
	var updateList []*iotc.NodeDiscoveryMessage = make([]*iotc.NodeDiscoveryMessage, 0)

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
func (nodes *NodeList) SetErrorStatus(address string, errorMsg string) (changed bool) {
	nodes.updateMutex.Lock()
	defer nodes.updateMutex.Unlock()

	node := nodes.getNode(address)
	if node == nil {
		return false
	}
	if node != nil {
		newNode := nodes.Clone(node)
		changed = false
		if node.Status[iotc.NodeStatusLastError] != errorMsg {
			node.Status[iotc.NodeStatusLastError] = errorMsg
			changed = true
		}

		if node.RunState != iotc.NodeRunStateError {
			changed = true
			newNode.RunState = iotc.NodeRunStateError
		}
		if changed {
			nodes.updateNode(newNode)
		}
	}
	return changed
}

// NewNode creates a node instance and adds it to the list.
// If the node exists it will remain unchanged
// This returns the node address
func (nodes *NodeList) NewNode(zoneID string, publisherID string, nodeID string, nodeType iotc.NodeType) string {
	nodes.updateMutex.Lock()
	defer nodes.updateMutex.Unlock()
	addr := MakeNodeAddress(zoneID, publisherID, nodeID)
	existingNode := nodes.getNode(addr)
	if existingNode == nil {
		newNode := NewNode(zoneID, publisherID, nodeID, nodeType)
		nodes.updateNode(newNode)
	}
	return addr
}

// SetNodeAttr updates node's attributes and publishes the updated node.
// Node is marked as modified for publication only if one of the attrParams has changes
// Use when additional node attributes has been discovered.
// - address of the node to update
// - param is the map with key-value pairs of attribute values to update
// returns true when node has changed, false if node doesn't exist or attributes haven't changed
func (nodes *NodeList) SetNodeAttr(address string, attrParams map[iotc.NodeAttr]string) (changed bool) {
	nodes.updateMutex.Lock()
	defer nodes.updateMutex.Unlock()
	node := nodes.getNode(address)
	if node == nil {
		return false
	}
	newNode := nodes.Clone(node)

	changed = false
	for key, value := range attrParams {
		if newNode.Attr[key] != value {
			newNode.Attr[key] = value
			changed = true
		}
	}
	// changed := newNode.SetNodeAttr(attrParams)
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
// returns true if configuration changes
func (nodes *NodeList) SetNodeConfigValues(address string, params map[iotc.NodeAttr]string) (changed bool) {
	nodes.updateMutex.Lock()
	defer nodes.updateMutex.Unlock()

	node := nodes.getNode(address)
	if node == nil {
		return false
	}
	newNode := nodes.Clone(node)

	changed = false
	for key, newValue := range params {
		config, configExists := node.Config[key]
		if !configExists {
			newConfig := iotc.ConfigAttr{Value: newValue}
			newNode.Config[key] = newConfig
			changed = true
		} else {
			if config.Value != newValue {
				config.Value = newValue
				newNode.Config[key] = config
				changed = true
			}
		}
	}

	if changed {
		nodes.updateNode(newNode)
	}
	return changed
}

// SetNodeRunState updates the node's runstate status
func (nodes *NodeList) SetNodeRunState(address string, runState iotc.NodeRunState) (changed bool) {
	nodes.updateMutex.Lock()
	defer nodes.updateMutex.Unlock()

	node := nodes.getNode(address)
	if node == nil {
		return false
	}

	changed = (node.RunState != runState)
	if changed {
		newNode := nodes.Clone(node)
		newNode.RunState = runState
		nodes.updateNode(newNode)
	}
	return changed
}

// SetNodeStatus updates one or more node's status attributes
// Nodes are immutable. If one or more status values have changed then a new node is created and
// published. The old node instance is discarded.
// - address of the node to update
// - statusAttr is the map with key-value pairs of updated node statusses
func (nodes *NodeList) SetNodeStatus(address string, statusAttr map[iotc.NodeStatus]string) (changed bool) {
	nodes.updateMutex.Lock()
	defer nodes.updateMutex.Unlock()
	node := nodes.getNode(address)
	if node == nil {
		return
	}
	newNode := nodes.Clone(node)
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

// UpdateNode replaces a node or adds a new node based on node.Address
// Intended to support Node immutability by making changes to a copy of a node and replacing
// the existing node with the updated node
// The updated node will be published
func (nodes *NodeList) UpdateNode(node *iotc.NodeDiscoveryMessage) {
	nodes.updateMutex.Lock()
	defer nodes.updateMutex.Unlock()
	nodes.updateNode(node)
}

// UpdateNodeConfig updates a node's configuration and publishes the updated node
// If a config already exists then its value is retained.
// Nodes are immutable. A new node is created and published and the old node instance is discarded.
// Use this when a device configuration has been identified, or when the config value updates.
// - address of the node to update
// - config is the config struct with description and value
// Returns a new node instance
func (nodes *NodeList) UpdateNodeConfig(address string, configAttr *iotc.ConfigAttr) {
	nodes.updateMutex.Lock()
	defer nodes.updateMutex.Unlock()
	node := nodes.getNode(address)
	if node == nil {
		return
	}
	existingAttr, found := node.Config[configAttr.ID]
	newAttr := *configAttr
	if found {
		newAttr.Value = existingAttr.Value
	}
	newNode := nodes.Clone(node)
	newNode.Config[configAttr.ID] = newAttr
	nodes.updateNode(newNode)
}

// UpdateNodes updates a list of nodes
// Intended to update the list with nodes from persistent storage
func (nodes *NodeList) UpdateNodes(updates []*iotc.NodeDiscoveryMessage) {
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
func (nodes *NodeList) getNode(address string) *iotc.NodeDiscoveryMessage {
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
func (nodes *NodeList) updateNode(node *iotc.NodeDiscoveryMessage) {
	nodes.nodeMap[node.Address] = node
	if nodes.updatedNodes == nil {
		nodes.updatedNodes = make(map[string]*iotc.NodeDiscoveryMessage)
	}
	nodes.updatedNodes[node.Address] = node
}

// MakeNodeAddress generates the discovery address of a node
// zoneID of the zone the node lives in.
// publisherID of the publisher for this node, unique for the zone
// nodeID of the node itself, unique for the publisher
func MakeNodeAddress(zoneID string, publisherID string, nodeID string) string {
	address := fmt.Sprintf("%s/%s/%s/"+iotc.MessageTypeNodeDiscovery, zoneID, publisherID, nodeID)
	return address
}

// NewNodeConfig creates a new node configuration instance.
// Use SetNodeConfig to update the node with this configuration
//
// id is the configuration attribute ID. See also iotc.NodeAttr for standard IDs
// dataType of the value. See also iotc.DataType for standard types.
// description of the value for humans
// defaultValue to use as default configuration value
// returns a new Configuration Attribute instance.
func NewNodeConfig(id iotc.NodeAttr, dataType iotc.DataType, description string, defaultValue string) *iotc.ConfigAttr {
	config := iotc.ConfigAttr{
		ID:          id,
		DataType:    dataType,
		Description: description,
		Default:     defaultValue,
	}
	return &config
}

// NewNode returns a new instance of a node
func NewNode(zoneID string, publisherID string, nodeID string, nodeType iotc.NodeType) *iotc.NodeDiscoveryMessage {
	address := MakeNodeAddress(zoneID, publisherID, nodeID)
	newNode := &iotc.NodeDiscoveryMessage{
		Address:     address,
		Attr:        map[iotc.NodeAttr]string{},
		Config:      map[iotc.NodeAttr]iotc.ConfigAttr{},
		ID:          nodeID,
		PublisherID: publisherID,
		Status:      make(map[iotc.NodeStatus]string),
		Type:        nodeType,
		Zone:        zoneID,
	}
	return newNode
}

// NewNodeList creates a new instance for node management
func NewNodeList() *NodeList {
	nodes := NodeList{
		nodeMap:     make(map[string]*iotc.NodeDiscoveryMessage),
		updateMutex: &sync.Mutex{},
	}
	return &nodes
}
