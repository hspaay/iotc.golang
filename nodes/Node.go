package nodes

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/hspaay/iotc.golang/iotc"
)

// Node contains logic for using the data from the node discovery message
// type Node struct {
// 	iotc.NodeDiscoveryMessage `json:"node"`
// }

// type Node = iotc.NodeDiscoveryMessage

// // Clone returns a copy of the node with new Attr, Config and Status maps
// // Intended for updating the node in a concurrent safe manner in combination with UpdateNode()
// // This does clones map values. Any updates to the map must use new instances of the values
// func (node *Node) Clone() *iotc.NodeDiscoveryMessage {
// 	newNode := *node

// 	newNode.Attr = make(map[iotc.NodeAttr]string)
// 	for key, value := range node.Attr {
// 		newNode.Attr[key] = value
// 	}
// 	newNode.Config = make(map[iotc.NodeAttr]iotc.ConfigAttr)
// 	for key, value := range node.Config {
// 		newNode.Config[key] = value
// 	}
// 	newNode.Status = make(map[iotc.NodeStatus]string)
// 	for key, value := range node.Status {
// 		newNode.Status[key] = value
// 	}
// 	return &newNode
// }

// GetNodeAlias returns the node alias, or node ID if no alias is set
// This is a convenience function to
// func GetNodeAlias(node *iotc.NodeDiscoveryMessage) (alias string, hasAlias bool) {
// 	hasAlias = false
// 	alias = node.ID
// 	aliasConfig, attrExists := node.Config[iotc.NodeAttrAlias]
// 	if attrExists && aliasConfig.Value != "" {
// 		alias = aliasConfig.Value
// 		hasAlias = true

// 	}
// 	return alias, hasAlias
// }

// GetNodeConfigInt returns the node configuration value as an integer
// This retuns the 'default' value if no value is set
func GetNodeConfigInt(node *iotc.NodeDiscoveryMessage, attrName iotc.NodeAttr) (value int, err error) {
	valueStr, configExists := GetNodeConfigValue(node, attrName)
	if !configExists {
		return 0, errors.New("Configuration does not exist")
	}
	return strconv.Atoi(valueStr)
}

// GetNodeConfigValue returns the node configuration value
// This retuns the 'default' value if no value is set
func GetNodeConfigValue(node *iotc.NodeDiscoveryMessage, attrName iotc.NodeAttr) (value string, configExists bool) {
	config, configExists := node.Config[attrName]
	if !configExists {
		return "", configExists
	}
	if config.Value == "" {
		return config.Default, configExists
	}
	return config.Value, configExists
}

// SetNodeAttr is a convenience function to update multiple attributes of a configuration
// Intended to update read-only attributes that describe the node.
// Returns true if one or more attributes have changed
func SetNodeAttr(node *iotc.NodeDiscoveryMessage, attrParams map[iotc.NodeAttr]string) (changed bool) {
	changed = false
	for key, value := range attrParams {
		if node.Attr[key] != value {
			node.Attr[key] = value
			changed = true
		}
	}
	return changed
}

// SetNodeConfigValues applies an update to a node's configuration values
// - param is the map with key-value pairs of configuration values to update
// Returns true if one or more attributes have changed
func SetNodeConfigValues(node *iotc.NodeDiscoveryMessage, params map[iotc.NodeAttr]string) (changed bool) {
	changed = false
	for key, newValue := range params {
		config, configExists := node.Config[key]
		if !configExists {
			newConfig := iotc.ConfigAttr{Value: newValue}
			node.Config[key] = newConfig
			changed = true
		} else {
			if config.Value != newValue {
				config.Value = newValue
				node.Config[key] = config
				changed = true
			}
		}
	}
	return changed
}

// MakeNodeDiscoveryAddress for publishing
// zoneID of the zone the node lives in.
// publisherID of the publisher for this node, unique for the zone
// nodeID of the node itself, unique for the publisher
func MakeNodeDiscoveryAddress(zoneID string, publisherID string, nodeID string) string {
	address := fmt.Sprintf("%s/%s/%s/"+iotc.MessageTypeNodeDiscovery, zoneID, publisherID, nodeID)
	return address
}

// NewConfigAttr instance for holding node configuration
// id of the attribute. See also iotc.NodeAttr for standard IDs
// dataType of the value. See also iotc.DataType for standard types.
// description of the value for humans
// defaultValue to use as default configuration value
// returns a new Configuration Attribute instance. Use nodes.SetNodeConfig to update the node with this configuration
func NewConfigAttr(id iotc.NodeAttr, dataType iotc.DataType, description string, defaultValue string) *iotc.ConfigAttr {
	config := iotc.ConfigAttr{
		ID:          id,
		DataType:    dataType,
		Description: description,
		Default:     defaultValue,
	}
	return &config
}

// NewNode create a node instance.
// Use UpdateNode to add it to the publisher
func NewNode(zoneID string, publisherID string, nodeID string, nodeType iotc.NodeType) *iotc.NodeDiscoveryMessage {
	address := MakeNodeDiscoveryAddress(zoneID, publisherID, nodeID)
	return &iotc.NodeDiscoveryMessage{
		Address:     address,
		Attr:        map[iotc.NodeAttr]string{},
		Config:      map[iotc.NodeAttr]iotc.ConfigAttr{},
		ID:          nodeID,
		PublisherID: publisherID,
		Status:      make(map[iotc.NodeStatus]string),
		Type:        nodeType,
		Zone:        zoneID,
	}
}
