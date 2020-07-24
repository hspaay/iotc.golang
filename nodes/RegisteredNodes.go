// Package nodes with node management
package nodes

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/iotdomain/iotdomain-go/types"
	"github.com/sirupsen/logrus"
)

// RegisteredNodes manages the publisher's node registration and publication for discovery
// Nodes are immutable. Any modifications made are applied to a new instance. The old node instance
// is discarded and replaced with the new instance.
// To make changes to a node directly, always Clone the node first and use UpdateNode to apply the change.
type RegisteredNodes struct {
	domain       string                                 // domain these nodes belong to
	publisherID  string                                 // ID of the publisher these nodes belong to
	deviceMap    map[string]*types.NodeDiscoveryMessage // registered nodes by device ID
	nodeMap      map[string]*types.NodeDiscoveryMessage // registered nodes by node ID
	updatedNodes map[string]*types.NodeDiscoveryMessage // updated nodes by node ID
	updateMutex  *sync.Mutex                            // mutex for async updating of nodes
}

// Clone returns a copy of the node with new Attr, Config and Status maps
// Intended for updating the node in a concurrent safe manner in combination with UpdateNode()
// This does clones map values. Any updates to the map must use new instances of the values
func (regNodes *RegisteredNodes) Clone(node *types.NodeDiscoveryMessage) *types.NodeDiscoveryMessage {
	newNode := *node

	newNode.Attr = make(map[types.NodeAttr]string)
	for key, value := range node.Attr {
		newNode.Attr[key] = value
	}
	// Shallow copy of the config list
	newNode.Config = node.Config

	newNode.Status = make(map[types.NodeStatus]string)
	for key, value := range node.Status {
		newNode.Status[key] = value
	}
	return &newNode
}

// CreateNode creates a node instance for a device or service and adds it to the list. If the node exists it will remain unchanged.
// This returns the node instance
func (regNodes *RegisteredNodes) CreateNode(deviceID string, nodeType types.NodeType) *types.NodeDiscoveryMessage {
	existingNode := regNodes.GetNodeByID(deviceID)
	if existingNode != nil {
		return existingNode
	}

	regNodes.updateMutex.Lock()
	defer regNodes.updateMutex.Unlock()

	newNode := NewNode(regNodes.domain, regNodes.publisherID, deviceID, nodeType)
	regNodes.updateNode(newNode)
	return newNode
}

// GetAllNodes returns a list of nodes
func (regNodes *RegisteredNodes) GetAllNodes() []*types.NodeDiscoveryMessage {
	regNodes.updateMutex.Lock()
	defer regNodes.updateMutex.Unlock()

	var nodeList = make([]*types.NodeDiscoveryMessage, 0)
	for _, node := range regNodes.nodeMap {
		nodeList = append(nodeList, node)
	}
	return nodeList
}

// GetNodeByAddress returns a node by its address using the nodeID from the address
// Returns nil if the nodeID is not registered
func (regNodes *RegisteredNodes) GetNodeByAddress(address string) *types.NodeDiscoveryMessage {
	regNodes.updateMutex.Lock()
	defer regNodes.updateMutex.Unlock()

	segments := strings.Split(address, "/")
	if len(segments) < 3 {
		return nil
	}
	var node = regNodes.nodeMap[segments[2]]
	return node
}

// GetNodeAttr returns a node attribute value
func (regNodes *RegisteredNodes) GetNodeAttr(nodeID string, attrName types.NodeAttr) string {
	regNodes.updateMutex.Lock()
	defer regNodes.updateMutex.Unlock()
	var node = regNodes.nodeMap[nodeID]
	if node == nil {
		return ""
	}
	attrValue, _ := node.Attr[attrName]
	return attrValue
}

// GetNodeConfigBool returns the node configuration value as a boolean
// address starts with the node's address
// This retuns the provided default value if no value is set or no default is configured, or the value is not an integer
// An error is returned when the node or configuration doesn't exist
func (regNodes *RegisteredNodes) GetNodeConfigBool(
	nodeID string, attrName types.NodeAttr, defaultValue bool) (value bool, err error) {

	valueStr, err := regNodes.GetNodeConfigString(nodeID, attrName, "")
	if err != nil {
		return defaultValue, err
	}
	if valueStr == "" {
		return defaultValue, nil
	}
	value, err = strconv.ParseBool(valueStr)
	if err != nil {
		msg := fmt.Sprintf("NodeList.GetNodeConfigBool: Node '%s' configuration '%s' is not a boolean: %s",
			nodeID, attrName, err)
		return defaultValue, errors.New(msg)
	}
	return value, nil
}

// GetNodeConfigFloat returns the node configuration value as an floating point number
// address starts with the node's address
// This retuns the provided default value if no value is set or no default is configured, or the value is not an integer
// An error is returned when the node or configuration doesn't exist or is not an integer
func (regNodes *RegisteredNodes) GetNodeConfigFloat(
	nodeID string, attrName types.NodeAttr, defaultValue float32) (value float32, err error) {

	valueStr, err := regNodes.GetNodeConfigString(nodeID, attrName, "")
	if err != nil {
		return defaultValue, err
	}
	if valueStr == "" {
		return defaultValue, nil
	}
	value64, err := strconv.ParseFloat(valueStr, 32)
	value = float32(value64)
	if err != nil {
		msg := fmt.Sprintf("NodeList.GetNodeConfigFloat: Node '%s' configuration '%s' is not a float: %s", nodeID, attrName, err)
		return defaultValue, errors.New(msg)
	}
	return value, nil
}

// GetNodeConfigInt returns the node configuration value as an integer
// This retuns the provided default value if no value is set or no default is configured, or the value is not an integer
// An error is returned when the node or configuration doesn't exist or is not an integer
func (regNodes *RegisteredNodes) GetNodeConfigInt(
	nodeID string, attrName types.NodeAttr, defaultValue int) (value int, err error) {

	valueStr, err := regNodes.GetNodeConfigString(nodeID, attrName, "")
	if err != nil {
		return defaultValue, err
	}
	if valueStr == "" {
		return defaultValue, nil
	}
	value, err = strconv.Atoi(valueStr)
	if err != nil {
		msg := fmt.Sprintf("NodeList.GetNodeConfigInt: Node '%s' configuration '%s' is not an integer: %s", nodeID, attrName, err)
		return defaultValue, errors.New(msg)
	}
	return value, nil
}

// GetNodeConfigString returns the attribute value of a node in this list
// This retuns the provided default value if no value is set and no default is configured.
// An error is returned when the node or configuration doesn't exist.
func (regNodes *RegisteredNodes) GetNodeConfigString(
	nodeID string, attrName types.NodeAttr, defaultValue string) (value string, err error) {
	regNodes.updateMutex.Lock()
	defer regNodes.updateMutex.Unlock()

	var node = regNodes.nodeMap[nodeID]
	if node == nil {
		msg := fmt.Sprintf("NodeList.GetNodeConfigString: Node '%s' not found", nodeID)
		return defaultValue, errors.New(msg)
	}

	// in case of error, always return defaultValue

	config, configExists := node.Config[attrName]
	if !configExists {
		msg := fmt.Sprintf("NodeList.GetNodeConfigString: Node '%s' configuration '%s' does not exist", nodeID, attrName)
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

// GetNodeBydeviceID returns a registered node by its device ID
// Returns nil if deviceID does not exist
func (regNodes *RegisteredNodes) GetNodeBydeviceID(deviceID string) *types.NodeDiscoveryMessage {
	regNodes.updateMutex.Lock()
	defer regNodes.updateMutex.Unlock()

	var node = regNodes.deviceMap[deviceID]
	return node
}

// GetNodeByID returns a nodes from the publisher
// Returns nil if address has no known node
func (regNodes *RegisteredNodes) GetNodeByID(nodeID string) *types.NodeDiscoveryMessage {
	regNodes.updateMutex.Lock()
	defer regNodes.updateMutex.Unlock()

	var node = regNodes.nodeMap[nodeID]
	return node
}

// GetUpdatedNodes returns the list of nodes that have been updated
// clearUpdates clears the list of updates. Intended for publishing only updated nodes.
func (regNodes *RegisteredNodes) GetUpdatedNodes(clearUpdates bool) []*types.NodeDiscoveryMessage {
	var updateList []*types.NodeDiscoveryMessage = make([]*types.NodeDiscoveryMessage, 0)

	regNodes.updateMutex.Lock()
	defer regNodes.updateMutex.Unlock()

	if regNodes.updatedNodes != nil {
		for _, node := range regNodes.updatedNodes {
			updateList = append(updateList, node)
		}
		if clearUpdates {
			regNodes.updatedNodes = nil
		}
	}
	return updateList
}

// SetAlias changes the nodeID and address of the node with the given nodeID.
//  Use an empty alias to restore the nodeID to its device ID.
//  This creates a new node instance for the alias and marks it as updated for publication. The existing
// node publication remains unchanged.
//  Returns true if a new node is created using the alias, false if node not found or alias is already in use
func (regNodes *RegisteredNodes) SetAlias(nodeID string, aliasID string) bool {
	node := regNodes.GetNodeByID(nodeID)
	if node == nil {
		// nodeID not found
		return false
	}
	aliasNode := regNodes.Clone(node)
	// reset nodeID
	if aliasID == "" {
		aliasNode.NodeID = node.DeviceID
	} else {
		// The alias must not be an existing device
		existingDeviceNode := regNodes.GetNodeBydeviceID(aliasID)
		if existingDeviceNode != nil {
			return false
		}
		aliasNode.NodeID = aliasID
	}
	// TODO: the alias remains in existence with the last updated timestamp. should
	// this be removed?
	aliasNode.Address = MakeNodeDiscoveryAddress(regNodes.domain, regNodes.publisherID, aliasNode.NodeID)
	regNodes.updateNode(aliasNode)
	// delete(regNodes.nodeMap, nodeID)
	regNodes.updateNode(node) // last update of the alias
	return true
}

// UpdateErrorStatus sets the node RunState to the given status with a lasterror message
// Use NodeRunStateError for errors and NodeRunStateReady to clear error
// This only updates the node if the status or lastError message changes
func (regNodes *RegisteredNodes) UpdateErrorStatus(nodeID string, runState string, errorMsg string) (changed bool) {
	node := regNodes.GetNodeByID(nodeID)
	if node == nil {
		return false
	}

	regNodes.updateMutex.Lock()
	defer regNodes.updateMutex.Unlock()

	newNode := regNodes.Clone(node)
	changed = false
	if node.Status[types.NodeStatusLastError] != errorMsg {
		newNode.Status[types.NodeStatusLastError] = errorMsg
		changed = true
	}

	if node.Status[types.NodeStatusRunState] != runState {
		changed = true
		newNode.Status[types.NodeStatusRunState] = runState
	}
	// Don't unnecesarily republish the node if the status doesnt change
	if changed {
		regNodes.updateNode(newNode)
	}
	return changed
}

// NewNodeConfig creates a new node configuration instance and adds it to the node with the given ID.
// If the configuration already exists, its dataType, description and defaultValue are updated
//  attrName is the configuration attribute name. See also types.NodeAttr for standard IDs
//  dataType of the value. See also types.DataType for standard types.
//  description of the value for humans
//  defaultValue to use as default configuration value
// returns a new Configuration Attribute instance.
func (regNodes *RegisteredNodes) NewNodeConfig(nodeID string, attrName types.NodeAttr, dataType types.DataType, description string, defaultValue string) *types.ConfigAttr {
	node := regNodes.GetNodeByID(nodeID)
	if node == nil {
		return nil
	}
	regNodes.updateMutex.Lock()
	defer regNodes.updateMutex.Unlock()

	config, configExists := node.Config[attrName]
	// update existing config or create a new one
	if !configExists {
		config = types.ConfigAttr{
			DataType:    dataType,
			Description: description,
			Default:     defaultValue,
		}
	} else {
		config.DataType = dataType
		config.Default = defaultValue
		config.Description = description
	}
	newNode := regNodes.Clone(node)
	newNode.Config[attrName] = config
	regNodes.updateNode(newNode)
	return &config
}

// UpdateNodeAttr updates node's attributes and publishes the updated node.
// Node is marked as modified for publication only if one of the attrParams has changes
//
// Use when additional node attributes has been discovered.
//  nodeID of the node to update
//  param is the map with key-value pairs of attribute values to update
// returns true when node has changed, false if node doesn't exist or attributes haven't changed
func (regNodes *RegisteredNodes) UpdateNodeAttr(nodeID string, attrParams map[types.NodeAttr]string) (changed bool) {
	node := regNodes.GetNodeByID(nodeID)
	if node == nil {
		return false
	}

	regNodes.updateMutex.Lock()
	defer regNodes.updateMutex.Unlock()
	newNode := regNodes.Clone(node)

	changed = false
	for key, value := range attrParams {
		if newNode.Attr[key] != value {
			newNode.Attr[key] = value
			changed = true
		}
	}
	// changed := newNode.SetNodeAttr(attrParams)
	if changed {
		regNodes.updateNode(newNode)
	}
	return changed
}

// UpdateNodeConfigValues applies an update to a registered node configuration values.
// Nodes are immutable. If one or more configuration values have changed then a new node is created and
// published and the old node instance is discarded.
//  param is the map with key-value pairs of configuration values to update
// returns true if configuration changes, false if configuration doesn't exist
func (regNodes *RegisteredNodes) UpdateNodeConfigValues(nodeID string, params types.NodeAttrMap) (changed bool) {

	node := regNodes.GetNodeByID(nodeID)
	if node == nil || params == nil {
		return false
	}
	regNodes.updateMutex.Lock()
	defer regNodes.updateMutex.Unlock()
	newNode := regNodes.Clone(node)

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
		regNodes.updateNode(newNode)
	}
	return changed
}

// // UpdateNode replaces a node or adds a new node based on node.Address.
// //
// // Intended to support Node immutability by making changes to a copy of a node and replacing
// // the existing node with the updated node
// // The updated node will be published
// func (regNodes *RegisteredNodes) UpdateNode(node *types.NodeDiscoveryMessage) {
// 	regNodes.updateMutex.Lock()
// 	defer regNodes.updateMutex.Unlock()
// 	regNodes.updateNode(node)
// }

// UpdateNodeConfig updates a node's configuration and publishes the updated node.
//
// If a config already exists then its value is retained but its configuration parameters are replaced.
// Nodes are immutable. A new node is created and published and the old node instance is discarded.
func (regNodes *RegisteredNodes) UpdateNodeConfig(nodeID string, attrName types.NodeAttr, configAttr *types.ConfigAttr) {
	node := regNodes.GetNodeByID(nodeID)
	if node == nil || configAttr == nil {
		return
	}
	regNodes.updateMutex.Lock()
	defer regNodes.updateMutex.Unlock()

	newNode := regNodes.Clone(node)
	newNode.Config[attrName] = *configAttr
	regNodes.updateNode(newNode)
}

// UpdateNodes updates a list of nodes.
//
// Intended to update the list with nodes from persistent storage
func (regNodes *RegisteredNodes) UpdateNodes(updates []*types.NodeDiscoveryMessage) {
	regNodes.updateMutex.Lock()
	defer regNodes.updateMutex.Unlock()

	for _, node := range updates {
		// fill in missing fields
		if node.Attr == nil {
			node.Attr = map[types.NodeAttr]string{}
		}
		if node.Config == nil {
			node.Config = map[types.NodeAttr]types.ConfigAttr{}
		}
		if node.Status == nil {
			node.Status = make(map[types.NodeStatus]string)
		}
		regNodes.updateNode(node)
	}
}

// UpdateNodeStatus updates one or more node's status attributes.
// Nodes are immutable. If one or more status values have changed then a new node is created and
// published. The old node instance is discarded.
//  statusAttr is the map with key-value pairs of updated node statusses
func (regNodes *RegisteredNodes) UpdateNodeStatus(nodeID string, statusAttr map[types.NodeStatus]string) (changed bool) {

	node := regNodes.GetNodeByID(nodeID)
	if node == nil {
		return
	}

	regNodes.updateMutex.Lock()
	defer regNodes.updateMutex.Unlock()

	newNode := regNodes.Clone(node)
	changed = false
	for key, value := range statusAttr {
		if newNode.Status[key] != value {
			newNode.Status[key] = value
			changed = true
		}
	}

	if changed {
		regNodes.updateNode(newNode)
	}
	return changed
}

// updateNode replaces a node and adds it to the list of updated nodes.
//  Use within a locked section.
func (regNodes *RegisteredNodes) updateNode(node *types.NodeDiscoveryMessage) {
	if node == nil {
		return
	}
	regNodes.nodeMap[node.NodeID] = node
	regNodes.deviceMap[node.DeviceID] = node
	if regNodes.updatedNodes == nil {
		regNodes.updatedNodes = make(map[string]*types.NodeDiscoveryMessage)
	}
	node.Timestamp = time.Now().Format(types.TimeFormat)
	regNodes.updatedNodes[node.Address] = node
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
func MakeNodeDiscoveryAddress(domain string, publisherID string, nodeID string) string {
	return MakeNodeAddress(domain, publisherID, nodeID, types.MessageTypeNodeDiscovery)
}

// NewNodeConfig creates a new node configuration instance.
// Intended for updating additional attributes before updating the actual configuration
// Use UpdateNodeConfig to update the node with this configuration
//
// dataType of the value. See also types.DataType for standard types.
// description of the value for humans
// defaultValue to use as default configuration value
// returns a new Configuration Attribute instance.
func NewNodeConfig(dataType types.DataType, description string, defaultValue string) *types.ConfigAttr {
	config := types.ConfigAttr{
		DataType:    dataType,
		Description: description,
		Default:     defaultValue,
	}
	return &config
}

// NewNode returns a new instance of a node.
func NewNode(domain string, publisherID string, deviceID string, nodeType types.NodeType) *types.NodeDiscoveryMessage {
	if domain == "" || publisherID == "" || deviceID == "" || nodeType == "" {
		logrus.Errorf("NewNode: empty argument, one of domain (%s), publisherID (%s), deviceID (%s) or nodeType (%s) ",
			domain, publisherID, deviceID, nodeType)
		return nil
	}
	address := MakeNodeAddress(domain, publisherID, deviceID, types.MessageTypeNodeDiscovery)
	newNode := &types.NodeDiscoveryMessage{
		Address:     address,
		Attr:        map[types.NodeAttr]string{},
		Config:      map[types.NodeAttr]types.ConfigAttr{},
		DeviceID:    deviceID,
		NodeID:      deviceID,
		PublisherID: publisherID,
		Status:      make(map[types.NodeStatus]string),
		NodeType:    nodeType,
		Timestamp:   time.Now().Format(types.TimeFormat),
	}
	newNode.Config[types.NodeAttrName] = *NewNodeConfig(types.DataTypeString, "Human friendly node name", "")
	newNode.Config[types.NodeAttrPublishEvent] = *NewNodeConfig(types.DataTypeString, "Enable publishing outputs as event", "false")
	newNode.Config[types.NodeAttrPublishHistory] = *NewNodeConfig(types.DataTypeBool, "Enable publishing output history", "true")
	newNode.Config[types.NodeAttrPublishLatest] = *NewNodeConfig(types.DataTypeBool, "Enable publishing latest output", "true")
	newNode.Config[types.NodeAttrPublishRaw] = *NewNodeConfig(types.DataTypeBool, "Enable publishing raw outputs", "true")
	return newNode
}

// NewRegisteredNodes creates a new instance for node management.
func NewRegisteredNodes(domain string, publisherID string) *RegisteredNodes {
	nodes := RegisteredNodes{
		domain:       domain,
		publisherID:  publisherID,
		deviceMap:    make(map[string]*types.NodeDiscoveryMessage),
		nodeMap:      make(map[string]*types.NodeDiscoveryMessage),
		updatedNodes: make(map[string]*types.NodeDiscoveryMessage),
		updateMutex:  &sync.Mutex{},
	}
	return &nodes
}

// // SplitNodeAddress splits any given address into a node part, messageType, in/output type and instance
// // address is the address to split
// // returns address parts, returns empty string if
// func SplitNodeAddress(address string) (nodeAddress string, messageType types.MessageType, ioType string, instance string) {
// 	// domain/publisher/node[/mtype[/iotype/instance]]
// 	segments := strings.Split(address, "/")
// 	if len(segments) < 3 {
// 		return
// 	}
// 	nodeAddress = strings.Join(segments[:3], "/")
// 	if len(segments) > 3 {
// 		messageType = types.MessageType(segments[3])
// 	}
// 	if len(segments) > 4 {
// 		ioType = segments[4]
// 	}
// 	if len(segments) > 5 {
// 		instance = segments[5]
// 	}
// 	return
// }
