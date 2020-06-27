// Package publisher with facade functions for nodes, inputs and outputs that work using nodeIDs
// instead of use of full addresses on the internal Nodes, Inputs and Outputs collections.
// Mostly intended to reduce boilerplate code in managing nodes, inputs and outputs
package publisher

import (
	"crypto/ecdsa"

	"github.com/iotdomain/iotdomain-go/inputs"
	"github.com/iotdomain/iotdomain-go/nodes"
	"github.com/iotdomain/iotdomain-go/outputs"
	"github.com/iotdomain/iotdomain-go/types"
)

// GetConfigValue convenience function to get a configuration value
// This retuns the 'default' value if no value is set
// func GetConfigValue(configMap map[string]iotd.ConfigAttr, attrName string) string {
// 	config, configExists := configMap[attrName]
// 	if !configExists {
// 		return ""
// 	}
// 	if config.Value == "" {
// 		return config.Default
// 	}
// 	return config.Value
// }

// GetNodeConfigBool convenience function to get a node configuration value as a boolean
// This retuns the given default if no configuration value exists and no configuration default is set
func (publisher *Publisher) GetNodeConfigBool(
	nodeID string, attrName types.NodeAttr, defaultValue bool) (value bool, err error) {
	nodeAddr := publisher.MakeNodeDiscoveryAddress(nodeID)
	return publisher.Nodes.GetNodeConfigBool(nodeAddr, attrName, defaultValue)
}

// GetNodeConfigFloat convenience function to get a node configuration value as a float number
// This retuns the given default if no configuration value exists and no configuration default is set
func (publisher *Publisher) GetNodeConfigFloat(
	nodeID string, attrName types.NodeAttr, defaultValue float32) (value float32, err error) {
	nodeAddr := publisher.MakeNodeDiscoveryAddress(nodeID)
	return publisher.Nodes.GetNodeConfigFloat(nodeAddr, attrName, defaultValue)
}

// GetNodeConfigInt convenience function to get a node configuration value as an integer
// This retuns the given default if no configuration value exists and no configuration default is set
func (publisher *Publisher) GetNodeConfigInt(
	nodeID string, attrName types.NodeAttr, defaultValue int) (value int, err error) {
	nodeAddr := publisher.MakeNodeDiscoveryAddress(nodeID)
	return publisher.Nodes.GetNodeConfigInt(nodeAddr, attrName, defaultValue)
}

// GetNodeConfigString convenience function to get a node configuration value as a string
// This retuns the given default if no configuration value exists and no configuration default is set
func (publisher *Publisher) GetNodeConfigString(
	nodeID string, attrName types.NodeAttr, defaultValue string) (value string, err error) {
	nodeAddr := publisher.MakeNodeDiscoveryAddress(nodeID)
	return publisher.Nodes.GetNodeConfigString(nodeAddr, attrName, defaultValue)
}

// GetNodeByID returns a node from this publisher or nil if the id isn't found in this publisher
// This is a convenience function as publishers tend to do this quite often
func (publisher *Publisher) GetNodeByID(nodeID string) (node *types.NodeDiscoveryMessage) {
	node = publisher.Nodes.GetNodeByID(publisher.Domain(), publisher.PublisherID(), nodeID)
	return node
}

// GetNodeStatus returns a node's status attribute
// This is a convenience function. See NodeList.GetNodeStatus for details
func (publisher *Publisher) GetNodeStatus(nodeID string, attrName types.NodeStatus) (value string, exists bool) {
	node := publisher.Nodes.GetNodeByID(publisher.Domain(), publisher.PublisherID(), nodeID)
	if node == nil {
		return "", false
	}
	value, exists = node.Status[attrName]
	return value, exists
}

// GetOutputByType returns a node output object using node id and output type and instance
// This is a convenience function using the publisher's output list
func (publisher *Publisher) GetOutputByType(nodeID string, outputType types.OutputType, instance string) *types.OutputDiscoveryMessage {
	nodeAddr := nodes.MakeNodeDiscoveryAddress(publisher.Domain(), publisher.PublisherID(), nodeID)
	outputAddr := outputs.MakeOutputDiscoveryAddress(nodeAddr, outputType, instance)
	output := publisher.Outputs.GetOutputByAddress(outputAddr)
	return output
}

// GetPublisherKey returns the public key of the publisher contained in the given address
// The address must at least contain a domain, publisherId and message type
func (publisher *Publisher) GetPublisherKey(address string) *ecdsa.PublicKey {
	return publisher.domainPublishers.GetPublisherKey(address)
}

// MakeNodeDiscoveryAddress makes the node discovery address using the publisher domain and publisherID
func (publisher *Publisher) MakeNodeDiscoveryAddress(nodeID string) string {
	addr := nodes.MakeNodeDiscoveryAddress(publisher.Domain(), publisher.PublisherID(), nodeID)
	return addr
}

// NewNode creates a new node and add it to this publisher's discovered nodes
// This is a convenience function that uses the publisher domain and id to create a node in its node list.
// returns the node's address
func (publisher *Publisher) NewNode(nodeID string, nodeType types.NodeType) string {
	addr := publisher.Nodes.NewNode(publisher.Domain(), publisher.PublisherID(), nodeID, nodeType)
	return addr
}

// NewInput creates a new node input and adds it to this publisher inputs list
// returns the input to allow for easy update
func (publisher *Publisher) NewInput(nodeID string, inputType types.InputType, instance string) *types.InputDiscoveryMessage {
	nodeAddr := nodes.MakeNodeDiscoveryAddress(publisher.Domain(), publisher.PublisherID(), nodeID)
	input := inputs.NewInput(nodeAddr, inputType, instance)
	publisher.Inputs.UpdateInput(input)
	return input
}

// NewOutput creates a new node output adds it to this publisher outputs list
// This is a convenience function for the publisher.Outputs list
// returns the output object to allow for easy updates
func (publisher *Publisher) NewOutput(nodeID string, outputType types.OutputType, instance string) *types.OutputDiscoveryMessage {
	nodeAddr := nodes.MakeNodeDiscoveryAddress(publisher.Domain(), publisher.PublisherID(), nodeID)
	output := outputs.NewOutput(nodeAddr, outputType, instance)
	publisher.Outputs.UpdateOutput(output)
	return output
}

// PublishRaw immediately publishes the given value of a node, output type and instance on the
// $raw output address. The content can be signed but is not encrypted.
// This is intended for publishing large values that should not be stored, for example images
func (publisher *Publisher) PublishRaw(output *types.OutputDiscoveryMessage, sign bool, value []byte) {
	aliasAddress := publisher.getOutputAliasAddress(output.Address, types.MessageTypeRaw)
	publisher.publishSigned(aliasAddress, sign, string(value))
}

// SetNodeAttr sets one or more attributes of the node
// This only updates the node if the status or lastError message changes
func (publisher *Publisher) SetNodeAttr(nodeID string, attrParams map[types.NodeAttr]string) (changed bool) {
	nodeAddr := nodes.MakeNodeDiscoveryAddress(publisher.Domain(), publisher.PublisherID(), nodeID)
	return publisher.Nodes.SetNodeAttr(nodeAddr, attrParams)
}

// SetNodeStatus sets one or more status attributes of the node
// This only updates the node if the status or lastError message changes
func (publisher *Publisher) SetNodeStatus(nodeID string, status map[types.NodeStatus]string) (changed bool) {
	nodeAddr := nodes.MakeNodeDiscoveryAddress(publisher.Domain(), publisher.PublisherID(), nodeID)
	return publisher.Nodes.SetNodeStatus(nodeAddr, status)
}

// SetNodeErrorStatus sets the node RunState to the given status with a lasterror message
// Use NodeRunStateError for errors and NodeRunStateReady to clear error
// This only updates the node if the status or lastError message changes
func (publisher *Publisher) SetNodeErrorStatus(nodeID string, status string, lastError string) {
	nodeAddr := nodes.MakeNodeDiscoveryAddress(publisher.Domain(), publisher.PublisherID(), nodeID)
	publisher.Nodes.SetErrorStatus(nodeAddr, status, lastError)
}

// UpdateNodeConfig updates a node's configuration and publishes the updated node.
//
// If a config already exists then its value is retained but its configuration parameters are replaced.
// Nodes are immutable. A new node is created and published and the old node instance is discarded.
func (publisher *Publisher) UpdateNodeConfig(nodeID string, attrName types.NodeAttr, configAttr *types.ConfigAttr) {
	nodeAddr := nodes.MakeNodeDiscoveryAddress(publisher.Domain(), publisher.PublisherID(), nodeID)
	publisher.Nodes.UpdateNodeConfig(nodeAddr, attrName, configAttr)
}

// UpdateOutputValue adds the new node output value to the front of the value history
// See NodeList.UpdateOutputValue for more details
func (publisher *Publisher) UpdateOutputValue(nodeID string, outputType types.OutputType, instance string, newValue string) bool {
	nodeAddr := nodes.MakeNodeDiscoveryAddress(publisher.Domain(), publisher.PublisherID(), nodeID)
	outputAddr := outputs.MakeOutputDiscoveryAddress(nodeAddr, outputType, instance)
	return publisher.OutputValues.UpdateOutputValue(outputAddr, newValue)
}
