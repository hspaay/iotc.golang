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
	nodeMap      map[string]*iotc.NodeDiscoveryMessage // node discovery address
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
	// Shallow copy of the config list
	newNode.Config = node.Config

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

// GetNodeByAddress returns a node by its node address using the domain, publisherID and nodeID
// address must contain the domain, publisherID and nodeID. Any other fields are ignored.
// Returns nil if address has no known node
func (nodes *NodeList) GetNodeByAddress(address string) *iotc.NodeDiscoveryMessage {
	nodes.updateMutex.Lock()
	defer nodes.updateMutex.Unlock()

	var node = nodes.getNode(address)
	return node
}

// GetNodeAttr returns a node attribute value
func (nodes *NodeList) GetNodeAttr(address string, attrName iotc.NodeAttr) string {
	nodes.updateMutex.Lock()
	defer nodes.updateMutex.Unlock()
	var node = nodes.getNode(address)
	attrValue, _ := node.Attr[attrName]
	return attrValue
}

// GetNodeConfigBool returns the node configuration value as a boolean
// address starts with the node's address
// This retuns the provided default value if no value is set or no default is configured, or the value is not an integer
// An error is returned when the node or configuration doesn't exist
func (nodes *NodeList) GetNodeConfigBool(
	address string, attrName iotc.NodeAttr, defaultValue bool) (value bool, err error) {

	valueStr, err := nodes.GetNodeConfigString(address, attrName, "")
	if err != nil {
		return defaultValue, err
	}
	if valueStr == "" {
		return defaultValue, nil
	}
	value, err = strconv.ParseBool(valueStr)
	if err != nil {
		msg := fmt.Sprintf("NodeList.GetNodeConfigBool: Node '%s' configuration '%s' is not a boolean: %s", address, attrName, err)
		return defaultValue, errors.New(msg)
	}
	return value, nil
}

// GetNodeConfigFloat returns the node configuration value as an floating point number
// address starts with the node's address
// This retuns the provided default value if no value is set or no default is configured, or the value is not an integer
// An error is returned when the node or configuration doesn't exist or is not an integer
func (nodes *NodeList) GetNodeConfigFloat(
	address string, attrName iotc.NodeAttr, defaultValue float32) (value float32, err error) {

	valueStr, err := nodes.GetNodeConfigString(address, attrName, "")
	if err != nil {
		return defaultValue, err
	}
	if valueStr == "" {
		return defaultValue, nil
	}
	value64, err := strconv.ParseFloat(valueStr, 32)
	value = float32(value64)
	if err != nil {
		msg := fmt.Sprintf("NodeList.GetNodeConfigFloat: Node '%s' configuration '%s' is not a float: %s", address, attrName, err)
		return defaultValue, errors.New(msg)
	}
	return value, nil
}

// GetNodeConfigInt returns the node configuration value as an integer
// address starts with the node's address
// This retuns the provided default value if no value is set or no default is configured, or the value is not an integer
// An error is returned when the node or configuration doesn't exist or is not an integer
func (nodes *NodeList) GetNodeConfigInt(
	address string, attrName iotc.NodeAttr, defaultValue int) (value int, err error) {

	valueStr, err := nodes.GetNodeConfigString(address, attrName, "")
	if err != nil {
		return defaultValue, err
	}
	if valueStr == "" {
		return defaultValue, nil
	}
	value, err = strconv.Atoi(valueStr)
	if err != nil {
		msg := fmt.Sprintf("NodeList.GetNodeConfigInt: Node '%s' configuration '%s' is not an integer: %s", address, attrName, err)
		return defaultValue, errors.New(msg)
	}
	return value, nil
}

// GetNodeConfigString returns the attribute value of a node in this list
// address must starts with the node's address: domain/publisher/nodeid. Any suffix is ignored.
// This retuns the provided default value if no value is set and no default is configured.
// An error is returned when the node or configuration doesn't exist.
func (nodes *NodeList) GetNodeConfigString(
	address string, attrName iotc.NodeAttr, defaultValue string) (value string, err error) {

	nodes.updateMutex.Lock()
	defer nodes.updateMutex.Unlock()
	// in case of error, always return defaultValue
	node := nodes.getNode(address)
	if node == nil {
		msg := fmt.Sprintf("NodeList.GetNodeConfigString: Node '%s' not found", address)
		return defaultValue, errors.New(msg)
	}
	config, configExists := node.Config[attrName]
	if !configExists {
		msg := fmt.Sprintf("NodeList.GetNodeConfigString: Node '%s' configuration '%s' does not exist", address, attrName)
		return defaultValue, errors.New(msg)
	}
	// if no value is known, use the configuration default
	attrValue, exists := node.Attr[attrName]
	if !exists || attrValue == "" {
		attrValue = config.Default
	}
	// if still no value is known, use the provided default
	if attrValue == "" {
		return defaultValue, nil
	}
	return attrValue, nil
}

// GetNodeByID returns a node by its domain, publisher and node ID
// Returns nil if address has no known node
func (nodes *NodeList) GetNodeByID(domain string, publisherID string, nodeID string) *iotc.NodeDiscoveryMessage {
	nodeAddr := fmt.Sprintf("%s/%s/%s/%s", domain, publisherID, nodeID, iotc.MessageTypeNodeDiscovery)

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

// SetErrorStatus sets the node RunState to the given status with a lasterror message
// Use NodeRunStateError for errors and NodeRunStateReady to clear error
// This only updates the node if the status or lastError message changes
func (nodes *NodeList) SetErrorStatus(address string, runState string, errorMsg string) (changed bool) {
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
			newNode.Status[iotc.NodeStatusLastError] = errorMsg
			changed = true
		}

		if node.Status[iotc.NodeStatusRunState] != runState {
			changed = true
			newNode.Status[iotc.NodeStatusRunState] = runState
		}
		// Don't unnecesarily republish the node if the status doesnt change
		if changed {
			nodes.updateNode(newNode)
		}
	}
	return changed
}

// NewNode creates a node instance and adds it to the list.
// If the node exists it will remain unchanged
// This returns the node discovery address
func (nodes *NodeList) NewNode(domain string, publisherID string, nodeID string, nodeType iotc.NodeType) string {
	nodes.updateMutex.Lock()
	defer nodes.updateMutex.Unlock()
	addr := MakeNodeAddress(domain, publisherID, nodeID, iotc.MessageTypeNodeDiscovery)
	existingNode := nodes.getNode(addr)
	if existingNode == nil {
		newNode := NewNode(domain, publisherID, nodeID, nodeType)
		nodes.updateNode(newNode)
	}
	return addr
}

// NewNodeConfig creates a new node configuration instance and adds it to the node.
// If the configuration already exists, its dataType, description and defaultValue are updated
// nodeAddr is the address of the node to update
// attrName is the configuration attribute name. See also iotc.NodeAttr for standard IDs
// dataType of the value. See also iotc.DataType for standard types.
// description of the value for humans
// defaultValue to use as default configuration value
// returns a new Configuration Attribute instance.
func (nodes *NodeList) NewNodeConfig(nodeAddr string, attrName iotc.NodeAttr, dataType iotc.DataType, description string, defaultValue string) *iotc.ConfigAttr {
	nodes.updateMutex.Lock()
	defer nodes.updateMutex.Unlock()

	node := nodes.getNode(nodeAddr)
	if node != nil {
		config, configExists := node.Config[attrName]
		// update existing config or create a new one
		if !configExists {
			config = iotc.ConfigAttr{
				DataType:    dataType,
				Description: description,
				Default:     defaultValue,
			}
		} else {
			config.DataType = dataType
			config.Default = defaultValue
			config.Description = description
		}
		newNode := nodes.Clone(node)
		newNode.Config[attrName] = config
		nodes.updateNode(newNode)
		return &config
	}
	return nil
}

// SetNodeAttr updates node's attributes and publishes the updated node.
// Node is marked as modified for publication only if one of the attrParams has changes
//
// Use when additional node attributes has been discovered.
//  address of the node to update
//  param is the map with key-value pairs of attribute values to update
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

// SetNodeConfigValues applies an update to a node's existing configuration.
// Nodes are immutable. If one or more configuration values have changed then a new node is created and
// published and the old node instance is discarded.
//  address is the node discovery address
//  param is the map with key-value pairs of configuration values to update
// returns true if configuration changes, false if configuration doesn't exist
func (nodes *NodeList) SetNodeConfigValues(address string, params map[iotc.NodeAttr]string) (changed bool) {
	nodes.updateMutex.Lock()
	defer nodes.updateMutex.Unlock()

	node := nodes.getNode(address)
	if node == nil || params == nil {
		return false
	}
	newNode := nodes.Clone(node)

	changed = false
	for key, newValue := range params {
		_, configExists := node.Config[key]
		if !configExists {
			// ignore invalid configuration
		} else {
			// update attribute with the new value
			// TODO: datatype check
			oldValue, attrExists := node.Attr[key]
			if !attrExists || oldValue != newValue {
				newNode.Attr[key] = newValue
				changed = true
			}
		}
	}

	if changed {
		nodes.updateNode(newNode)
	}
	return changed
}

// SetNodeStatus updates one or more node's status attributes.
// Nodes are immutable. If one or more status values have changed then a new node is created and
// published. The old node instance is discarded.
//  address of the node to update
//  statusAttr is the map with key-value pairs of updated node statusses
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

// UpdateNode replaces a node or adds a new node based on node.Address.
//
// Intended to support Node immutability by making changes to a copy of a node and replacing
// the existing node with the updated node
// The updated node will be published
func (nodes *NodeList) UpdateNode(node *iotc.NodeDiscoveryMessage) {
	nodes.updateMutex.Lock()
	defer nodes.updateMutex.Unlock()
	nodes.updateNode(node)
}

// UpdateNodeConfig updates a node's configuration and publishes the updated node.
//
// If a config already exists then its value is retained but its configuration parameters are replaced.
// Nodes are immutable. A new node is created and published and the old node instance is discarded.
func (nodes *NodeList) UpdateNodeConfig(address string, attrName iotc.NodeAttr, configAttr *iotc.ConfigAttr) {
	nodes.updateMutex.Lock()
	defer nodes.updateMutex.Unlock()
	node := nodes.getNode(address)
	if node == nil {
		return
	}
	newNode := nodes.Clone(node)
	newNode.Config[attrName] = *configAttr
	nodes.updateNode(newNode)
}

// UpdateNodes updates a list of nodes.
//
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

// getNode returns a node by its node address using the domain, publisherID and nodeID.
//
// address must contain the domain, publisher and nodeID. Any other fields are ignored.
// Intended for use within a locked section for updating, eg lock - read - update - write - unlock
// Returns nil if address has no known node
func (nodes *NodeList) getNode(address string) *iotc.NodeDiscoveryMessage {
	segments := strings.Split(address, "/")
	if len(segments) <= 3 {
		return nil
	}
	segments[3] = iotc.MessageTypeNodeDiscovery
	nodeAddr := strings.Join(segments[:4], "/")
	var node = nodes.nodeMap[nodeAddr]
	return node
}

// updateNode replaces a node and adds it to the list of updated nodes.
//
// Use within a locked section.
func (nodes *NodeList) updateNode(node *iotc.NodeDiscoveryMessage) {
	nodes.nodeMap[node.Address] = node
	if nodes.updatedNodes == nil {
		nodes.updatedNodes = make(map[string]*iotc.NodeDiscoveryMessage)
	}
	nodes.updatedNodes[node.Address] = node
}

// MakeNodeAddress generates the address of a node: domain/publisherID/nodeID[/messageType].
//
// As per standard, the domain of the domain the node lives in; publisherID of the publisher for this node,
// unique for the domain; nodeID of the node itself, unique for the publisher; messageType is optional,
// use "" if it doesn't apply.
func MakeNodeAddress(domain string, publisherID string, nodeID string, messageType string) string {
	address := fmt.Sprintf("%s/%s/%s", domain, publisherID, nodeID)
	if messageType != "" {
		address = address + "/" + messageType
	}
	return address
}

// MakeNodeDiscoveryAddress generates the address of a node: domain/publisherID/nodeID/$node.
// Intended for lookup of nodes in the node list.
func MakeNodeDiscoveryAddress(domain string, publisherID string, nodeID string) string {
	address := fmt.Sprintf("%s/%s/%s/%s", domain, publisherID, nodeID, iotc.MessageTypeNodeDiscovery)
	return address
}

// NewNodeConfig creates a new node configuration instance.
// Intended for updating additional attributes before updating the actual configuration
// Use UpdateNodeConfig to update the node with this configuration
//
// dataType of the value. See also iotc.DataType for standard types.
// description of the value for humans
// defaultValue to use as default configuration value
// returns a new Configuration Attribute instance.
func NewNodeConfig(dataType iotc.DataType, description string, defaultValue string) *iotc.ConfigAttr {
	config := iotc.ConfigAttr{
		DataType:    dataType,
		Description: description,
		Default:     defaultValue,
	}
	return &config
}

// NewNode returns a new instance of a node.
// This node will have default configurations for name and alias.
func NewNode(domain string, publisherID string, nodeID string, nodeType iotc.NodeType) *iotc.NodeDiscoveryMessage {
	address := MakeNodeAddress(domain, publisherID, nodeID, iotc.MessageTypeNodeDiscovery)
	newNode := &iotc.NodeDiscoveryMessage{
		Address:     address,
		Attr:        map[iotc.NodeAttr]string{},
		Config:      map[iotc.NodeAttr]iotc.ConfigAttr{},
		NodeID:      nodeID,
		PublisherID: publisherID,
		Status:      make(map[iotc.NodeStatus]string),
		Type:        nodeType,
	}
	newNode.Config[iotc.NodeAttrAlias] = *NewNodeConfig(iotc.DataTypeString, "Alias node ID for inputs and outputs", "")
	newNode.Config[iotc.NodeAttrName] = *NewNodeConfig(iotc.DataTypeString, "Human friendly node name", "")
	return newNode
}

// NewNodeList creates a new instance for node management.
func NewNodeList() *NodeList {
	nodes := NodeList{
		nodeMap:     make(map[string]*iotc.NodeDiscoveryMessage),
		updateMutex: &sync.Mutex{},
	}
	return &nodes
}

// // SplitNodeAddress splits any given address into a node part, messageType, in/output type and instance
// // address is the address to split
// // returns address parts, returns empty string if
// func SplitNodeAddress(address string) (nodeAddress string, messageType iotc.MessageType, ioType string, instance string) {
// 	// domain/publisher/node[/mtype[/iotype/instance]]
// 	segments := strings.Split(address, "/")
// 	if len(segments) < 3 {
// 		return
// 	}
// 	nodeAddress = strings.Join(segments[:3], "/")
// 	if len(segments) > 3 {
// 		messageType = iotc.MessageType(segments[3])
// 	}
// 	if len(segments) > 4 {
// 		ioType = segments[4]
// 	}
// 	if len(segments) > 5 {
// 		instance = segments[5]
// 	}
// 	return
// }
