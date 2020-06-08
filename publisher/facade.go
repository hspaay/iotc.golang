// Package publisher with facade functions for nodes, inputs and outputs that work using nodeIDs
// instead of use of full addresses on the internal Nodes, Inputs and Outputs collections.
// Mostly intended to reduce boilerplate code in managing nodes, inputs and outputs
package publisher

import (
	"github.com/hspaay/iotc.golang/iotc"
	"github.com/hspaay/iotc.golang/nodes"
)

// GetConfigValue convenience function to get a configuration value
// This retuns the 'default' value if no value is set
// func GetConfigValue(configMap map[string]iotc.ConfigAttr, attrName string) string {
// 	config, configExists := configMap[attrName]
// 	if !configExists {
// 		return ""
// 	}
// 	if config.Value == "" {
// 		return config.Default
// 	}
// 	return config.Value
// }

// GetNodeConfigInt convenience function to get a node configuration value as an integer
// This retuns the 'default' value if no value is set
func (publisher *Publisher) GetNodeConfigInt(nodeID string, attrName iotc.NodeAttr) (value int, err error) {
	nodeAddr := publisher.MakeNodeDiscoveryAddress(nodeID)
	value, err = publisher.Nodes.GetNodeConfigInt(nodeAddr, attrName)
	return value, err
}

// GetNodeConfigValue convenience function to get a node configuration value
// nodeID is the node to read from
// attrName identifies the configuration attribute to get
// retuns the 'defaultValue' if no value is set
func (publisher *Publisher) GetNodeConfigValue(
	nodeID string, attrName iotc.NodeAttr, defaultValue string) (value string, exists bool) {

	nodeAddr := publisher.MakeNodeDiscoveryAddress(nodeID)
	value, exists = publisher.Nodes.GetNodeConfigValue(nodeAddr, attrName)
	if value == "" {
		value = defaultValue
	}
	return value, exists
}

// GetNodeByID returns a node from this publisher or nil if the id isn't found in this publisher
// This is a convenience function as publishers tend to do this quite often
func (publisher *Publisher) GetNodeByID(nodeID string) (node *iotc.NodeDiscoveryMessage) {
	node = publisher.Nodes.GetNodeByID(publisher.domain, publisher.publisherID, nodeID)
	return node
}

// GetNodeStatus returns a node's status attribute
// This is a convenience function. See NodeList.GetNodeStatus for details
func (publisher *Publisher) GetNodeStatus(nodeID string, attrName iotc.NodeStatus) (value string, exists bool) {
	node := publisher.Nodes.GetNodeByID(publisher.domain, publisher.publisherID, nodeID)
	if node == nil {
		return "", false
	}
	value, exists = node.Status[attrName]
	return value, exists
}

// GetOutputByType returns a node output object using node id and output type and instance
// This is a convenience function using the publisher's output list
func (publisher *Publisher) GetOutputByType(nodeID string, outputType string, instance string) *iotc.OutputDiscoveryMessage {
	nodeAddr := nodes.MakeNodeDiscoveryAddress(publisher.domain, publisher.publisherID, nodeID)
	outputAddr := nodes.MakeOutputDiscoveryAddress(nodeAddr, outputType, instance)
	output := publisher.Outputs.GetOutputByAddress(outputAddr)
	return output
}

// MakeNodeDiscoveryAddress makes the node discovery address using the publisher domain and publisherID
func (publisher *Publisher) MakeNodeDiscoveryAddress(nodeID string) string {
	addr := nodes.MakeNodeDiscoveryAddress(publisher.domain, publisher.publisherID, nodeID)
	return addr
}

// NewNode creates a new node and add it to this publisher's discovered nodes
// This is a convenience function that uses the publisher domain and id to create a node in its node list.
// returns the node's address
func (publisher *Publisher) NewNode(nodeID string, nodeType iotc.NodeType) string {
	addr := publisher.Nodes.NewNode(publisher.domain, publisher.publisherID, nodeID, nodeType)
	return addr
}

// NewNodeConfig creates a new node configuration for a node of this publisher and update the node
// If the configuration already exists, its dataType, description and defaultValue are updated but
// the value is retained.
// See NodeList.NewNodeConfig for more details
// Returns the node config object which can be used with UpdateNodeConfig
func (publisher *Publisher) NewNodeConfig(
	nodeID string,
	attrName iotc.NodeAttr,
	dataType iotc.DataType,
	description string,
	defaultValue string) *iotc.ConfigAttr {

	nodeAddr := nodes.MakeNodeDiscoveryAddress(publisher.domain, publisher.publisherID, nodeID)
	config := publisher.Nodes.NewNodeConfig(nodeAddr, attrName, dataType, description, defaultValue)
	return config
}

// NewInput creates a new node input and adds it to this publisher inputs list
// returns the input to allow for easy update
func (publisher *Publisher) NewInput(nodeID string, inputType string, instance string) *iotc.InputDiscoveryMessage {
	nodeAddr := nodes.MakeNodeDiscoveryAddress(publisher.domain, publisher.publisherID, nodeID)
	input := nodes.NewInput(nodeAddr, inputType, instance)
	publisher.Inputs.UpdateInput(input)
	return input
}

// NewOutput creates a new node output adds it to this publisher outputs list
// This is a convenience function for the publisher.Outputs list
// returns the output object to allow for easy updates
func (publisher *Publisher) NewOutput(nodeID string, outputType string, instance string) *iotc.OutputDiscoveryMessage {
	nodeAddr := nodes.MakeNodeDiscoveryAddress(publisher.domain, publisher.publisherID, nodeID)
	output := nodes.NewOutput(nodeAddr, outputType, instance)
	publisher.Outputs.UpdateOutput(output)
	return output
}

// SetNodeAttr sets one or more attributes of the node
// This only updates the node if the status or lastError message changes
func (publisher *Publisher) SetNodeAttr(nodeID string, attrParams map[iotc.NodeAttr]string) (changed bool) {
	nodeAddr := nodes.MakeNodeDiscoveryAddress(publisher.domain, publisher.publisherID, nodeID)
	return publisher.Nodes.SetNodeAttr(nodeAddr, attrParams)
}

// SetNodeStatus sets one or more status attributes of the node
// This only updates the node if the status or lastError message changes
func (publisher *Publisher) SetNodeStatus(nodeID string, status map[iotc.NodeStatus]string) (changed bool) {
	nodeAddr := nodes.MakeNodeDiscoveryAddress(publisher.domain, publisher.publisherID, nodeID)
	return publisher.Nodes.SetNodeStatus(nodeAddr, status)
}

// SetNodeErrorStatus sets the node RunState to the given status with a lasterror message
// Use NodeRunStateError for errors and NodeRunStateReady to clear error
// This only updates the node if the status or lastError message changes
func (publisher *Publisher) SetNodeErrorStatus(nodeID string, status string, lastError string) {
	nodeAddr := nodes.MakeNodeDiscoveryAddress(publisher.domain, publisher.publisherID, nodeID)
	publisher.Nodes.SetErrorStatus(nodeAddr, status, lastError)
}

// UpdateOutputValue adds the new node output value to the front of the value history
// See NodeList.UpdateOutputValue for more details
func (publisher *Publisher) UpdateOutputValue(nodeID string, outputType string, instance string, newValue string) bool {
	nodeAddr := nodes.MakeNodeDiscoveryAddress(publisher.domain, publisher.publisherID, nodeID)
	outputAddr := nodes.MakeOutputDiscoveryAddress(nodeAddr, outputType, instance)
	return publisher.OutputValues.UpdateOutputValue(outputAddr, newValue)
}
