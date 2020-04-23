package nodes

import (
	"fmt"

	"github.com/hspaay/iotconnect.golang/messaging"
)

// Node contains logic for using the data from the node discovery message
type Node struct {
	messaging.NodeDiscoveryMessage
}

// Clone returns a copy of the node with new Attr, Config and Status maps
// Intended for updating the node in a concurrent safe manner in combination with UpdateNode()
// This does clones map values. Any updates to the map must use new instances of the values
func (node *Node) Clone() *Node {
	newNode := *node

	newNode.Attr = make(map[messaging.NodeAttr]string)
	for key, value := range node.Attr {
		newNode.Attr[key] = value
	}
	newNode.Config = make(map[messaging.NodeAttr]messaging.ConfigAttr)
	for key, value := range node.Config {
		newNode.Config[key] = value
	}
	newNode.Status = make(map[messaging.NodeStatus]string)
	for key, value := range node.Status {
		newNode.Status[key] = value
	}
	return &newNode
}

// GetAlias returns the node alias, or node ID if no alias is set
func (node *Node) GetAlias() (alias string, hasAlias bool) {
	hasAlias = false
	alias = node.ID
	aliasConfig, attrExists := node.Config[messaging.NodeAttrAlias]
	if attrExists && aliasConfig.Value != "" {
		alias = aliasConfig.Value
		hasAlias = true

	}
	return alias, hasAlias
}

// GetConfigValue returns the node configuration value
// This retuns the 'default' value if no value is set
func (node *Node) GetConfigValue(attrName messaging.NodeAttr) (value string, configExists bool) {
	config, configExists := node.Config[attrName]
	if !configExists {
		return "", configExists
	}
	if config.Value == "" {
		return config.Default, configExists
	}
	return config.Value, configExists
}

// UpdateNodeAttr is a convenience function to update multiple attributes of a configuration
// Intended to update read-only attributes that describe the node.
// Returns true if one or more attributes have changed
func (node *Node) UpdateNodeAttr(attrParams map[messaging.NodeAttr]string) (changed bool) {
	changed = false
	for key, value := range attrParams {
		if node.Attr[key] != value {
			node.Attr[key] = value
			changed = true
		}
	}
	return changed
}

// UpdateNodeConfigValues applies an update to a node's configuration values
// - param is the map with key-value pairs of configuration values to update
// Returns true if one or more attributes have changed
func (node *Node) UpdateNodeConfigValues(params map[messaging.NodeAttr]string) (changed bool) {
	changed = false
	for key, newValue := range params {
		config, configExists := node.Config[key]
		if !configExists {
			newConfig := messaging.ConfigAttr{Value: newValue}
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

// UpdateNodeStatus is a convenience function to update multiple node status fields
// Returns true if one or more status values have changed
func (node *Node) UpdateNodeStatus(attrParams map[messaging.NodeStatus]string) (changed bool) {
	changed = false
	for key, value := range attrParams {
		if node.Status[key] != value {
			node.Status[key] = value
			changed = true
		}
	}
	return changed
}

// MakeNodeDiscoveryAddress for publishing
func MakeNodeDiscoveryAddress(zoneID string, publisherID string, nodeID string) string {
	address := fmt.Sprintf("%s/%s/%s/"+messaging.CommandNodeDiscovery, zoneID, publisherID, nodeID)
	return address
}

// NewConfigAttr instance for holding node configuration
func NewConfigAttr(id messaging.NodeAttr, dataType messaging.DataType, description string, value string) *messaging.ConfigAttr {
	config := messaging.ConfigAttr{
		ID:          id,
		DataType:    dataType,
		Description: description,
		Value:       value,
	}
	return &config
}

// NewNode instance
func NewNode(zoneID string, publisherID string, nodeID string) *Node {
	address := MakeNodeDiscoveryAddress(zoneID, publisherID, nodeID)
	return &Node{
		messaging.NodeDiscoveryMessage{
			Address:     address,
			Attr:        map[messaging.NodeAttr]string{},
			Config:      map[messaging.NodeAttr]messaging.ConfigAttr{},
			ID:          nodeID,
			PublisherID: publisherID,
			Status:      make(map[messaging.NodeStatus]string),
			Zone:        zoneID,
		},
	}
}
