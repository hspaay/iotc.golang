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
	"github.com/sirupsen/logrus"
)

// GetDomainIntput returns a discovered domain input
func (publisher *Publisher) GetDomainInput(address string) *types.InputDiscoveryMessage {
	return publisher.domainInputs.GetInputByAddress(address)
}

// GetDomainInputs returns all discovered domain inputs
func (publisher *Publisher) GetDomainInputs() []*types.InputDiscoveryMessage {
	return publisher.domainInputs.GetAllInputs()
}

// GetDomainNode returns a discovered domain node by its address
func (publisher *Publisher) GetDomainNode(address string) *types.NodeDiscoveryMessage {
	return publisher.domainNodes.GetNodeByAddress(address)
}

// GetDomainNodes returns all discovered domain nodes
func (publisher *Publisher) GetDomainNodes() []*types.NodeDiscoveryMessage {
	return publisher.domainNodes.GetAllNodes()
}

// GetDomainOutput returns a discovered domain output by its address
func (publisher *Publisher) GetDomainOutput(address string) *types.OutputDiscoveryMessage {
	return publisher.domainOutputs.GetOutputByAddress(address)
}

// GetDomainOutputs returns all discovered domain outputs
func (publisher *Publisher) GetDomainOutputs() []*types.OutputDiscoveryMessage {
	return publisher.domainOutputs.GetAllOutputs()
}

// GetDomainPublishers returns all discovered domain publishers
func (publisher *Publisher) GetDomainPublishers() []*types.PublisherIdentityMessage {
	return publisher.domainPublishers.GetAllPublishers()
}

// GetInput Get a registered input by node ID
func (publisher *Publisher) GetInput(nodeID string, inputType types.InputType, instance string) *types.InputDiscoveryMessage {
	return publisher.registeredInputs.GetInput(nodeID, inputType, instance)
}

// GetInputByAddress returns a registered input by its full address
func (publisher *Publisher) GetInputByAddress(address string) *types.InputDiscoveryMessage {
	return publisher.registeredInputs.GetInputByAddress(address)
}

// GetInputs returns a list of all registered inputs
func (publisher *Publisher) GetInputs() []*types.InputDiscoveryMessage {
	return publisher.registeredInputs.GetAllInputs()
}

// GetIdentity returns the publisher public identity including public signing key
func (publisher *Publisher) GetIdentity() *types.PublisherIdentityMessage {
	return &publisher.fullIdentity.PublisherIdentityMessage
}

// GetIdentityKeys returns the private/public key pair of this publisher
func (publisher *Publisher) GetIdentityKeys() *ecdsa.PrivateKey {
	return publisher.identityPrivateKey
}

// GetNodeAttr returns a node attribute value
func (publisher *Publisher) GetNodeAttr(nodeID string, attrName types.NodeAttr) string {
	return publisher.registeredNodes.GetNodeAttr(nodeID, attrName)
}

// GetNodeByAddress returns a registered node by its full address
func (publisher *Publisher) GetNodeByAddress(address string) *types.NodeDiscoveryMessage {
	return publisher.registeredNodes.GetNodeByAddress(address)
}

// GetNodeByID returns a node from this publisher or nil if the id isn't found in this publisher
func (publisher *Publisher) GetNodeByID(nodeID string) (node *types.NodeDiscoveryMessage) {
	node = publisher.registeredNodes.GetNodeByID(nodeID)
	return node
}

// GetNodeConfigBool returns a node configuration value as a boolean
// This retuns the given default if no configuration value exists and no configuration default is set
func (publisher *Publisher) GetNodeConfigBool(
	nodeID string, attrName types.NodeAttr, defaultValue bool) (value bool, err error) {
	return publisher.registeredNodes.GetNodeConfigBool(nodeID, attrName, defaultValue)
}

// GetNodeConfigFloat returns a node configuration value as a float number.
// This retuns the given default if no configuration value exists and no configuration default is set
func (publisher *Publisher) GetNodeConfigFloat(
	nodeID string, attrName types.NodeAttr, defaultValue float32) (value float32, err error) {
	return publisher.registeredNodes.GetNodeConfigFloat(nodeID, attrName, defaultValue)
}

// GetNodeConfigInt returns a node configuration value as an integer
// This retuns the given default if no configuration value exists and no configuration default is set
func (publisher *Publisher) GetNodeConfigInt(
	nodeID string, attrName types.NodeAttr, defaultValue int) (value int, err error) {
	return publisher.registeredNodes.GetNodeConfigInt(nodeID, attrName, defaultValue)
}

// GetNodeConfigString returns a node configuration value as a string
// This retuns the given default if no configuration value exists and no configuration default is set
func (publisher *Publisher) GetNodeConfigString(
	nodeID string, attrName types.NodeAttr, defaultValue string) (value string, err error) {
	return publisher.registeredNodes.GetNodeConfigString(nodeID, attrName, defaultValue)
}

// GetNodes returns a list of all registered nodes
func (publisher *Publisher) GetNodes() []*types.NodeDiscoveryMessage {
	return publisher.registeredNodes.GetAllNodes()
}

// GetNodeStatus returns a status attribute of a registered node
func (publisher *Publisher) GetNodeStatus(nodeID string, attrName types.NodeStatus) (value string, exists bool) {
	node := publisher.registeredNodes.GetNodeByID(nodeID)
	if node == nil {
		return "", false
	}
	value, exists = node.Status[attrName]
	return value, exists
}

// GetPublisherKey returns the public key of the publisher contained in the given address
// The address must at least contain a domain and publisherId
func (publisher *Publisher) GetPublisherKey(address string) *ecdsa.PublicKey {
	return publisher.domainPublishers.GetPublisherKey(address)
}

// GetOutput get a registered output
func (publisher *Publisher) GetOutput(nodeID string, outputType types.OutputType, instance string) *types.OutputDiscoveryMessage {
	return publisher.registeredOutputs.GetOutput(nodeID, outputType, instance)
}

// GetOutputs returns a list of all registered outputs
func (publisher *Publisher) GetOutputs() []*types.OutputDiscoveryMessage {
	return publisher.registeredOutputs.GetAllOutputs()
}

// GetOutputValue returns the registered output's value object including timestamp
func (publisher *Publisher) GetOutputValue(nodeID string, outputType types.OutputType, instance string) *types.OutputValue {
	return publisher.registeredOutputValues.GetOutputValueByType(nodeID, outputType, instance)
}

// MakeNodeDiscoveryAddress makes the node discovery address using the publisher domain and publisherID
func (publisher *Publisher) MakeNodeDiscoveryAddress(nodeID string) string {
	addr := nodes.MakeNodeDiscoveryAddress(publisher.Domain(), publisher.PublisherID(), nodeID)
	return addr
}

// NewInput creates a new node input and adds it to this publisher inputs list
// returns the input to allow for easy update
func (publisher *Publisher) NewInput(nodeID string, inputType types.InputType, instance string) *types.InputDiscoveryMessage {
	input := publisher.registeredInputs.NewInput(nodeID, inputType, instance)
	return input
}

// NewNode creates a new node and add it to this publisher's discovered nodes
// returns the new node instance
func (publisher *Publisher) NewNode(nodeID string, nodeType types.NodeType) *types.NodeDiscoveryMessage {
	node := publisher.registeredNodes.NewNode(nodeID, nodeType)
	return node
}

// NewOutput creates a new node output adds it to this publisher outputs list
// returns the output object to allow for easy updates
func (publisher *Publisher) NewOutput(nodeID string, outputType types.OutputType, instance string) *types.OutputDiscoveryMessage {
	output := publisher.registeredOutputs.NewOutput(nodeID, outputType, instance)
	return output
}

// PublishConfigureNode publishes a $configure command to a domain node
// Returns true if successful, false if the domain node publisher cannot be found or has no public key
// and the message is not sent.
func (publisher *Publisher) PublishConfigureNode(domainNodeAddr string, attr types.NodeAttrMap) bool {
	destPubKey := publisher.GetPublisherKey(domainNodeAddr)
	if destPubKey == nil {
		logrus.Warnf("PublishConfigureNode: no public key found to encrypt command for node %s. Message not sent.", domainNodeAddr)
		return false
	}
	nodes.PublishConfigureNode(domainNodeAddr, attr, publisher.Address(), publisher.messageSigner, destPubKey)
	return true
}

// PublishRaw immediately publishes the given value of a node, output type and instance on the
// $raw output address. The content can be signed but is not encrypted.
// This is intended for publishing large values that should not be stored, for example images
func (publisher *Publisher) PublishRaw(output *types.OutputDiscoveryMessage, sign bool, value string) {
	outputs.PublishOutputRaw(output, value, publisher.messageSigner)
}

// PublishSetInput publishes a $set input command to the given domain input address
// Returns true if successful, false if the input's publisher cannot be found and the message
// is not sent.
func (publisher *Publisher) PublishSetInput(domainInputAddr string, value string) bool {
	destPubKey := publisher.GetPublisherKey(domainInputAddr)
	if destPubKey == nil {
		logrus.Warnf("PublishSetInput: no public key found to encrypt command for set input to %s. Message not sent.", domainInputAddr)
		return false
	}
	inputs.PublishSetInput(domainInputAddr, value, publisher.Address(), publisher.messageSigner, destPubKey)
	return true
}

// UpdateNodeErrorStatus sets a registered node RunState to the given status with a lasterror message
// Use NodeRunStateError for errors and NodeRunStateReady to clear error
// This only updates the node if the status or lastError message changes
func (publisher *Publisher) UpdateNodeErrorStatus(nodeID string, status string, lastError string) {
	publisher.registeredNodes.UpdateErrorStatus(nodeID, status, lastError)
}

// UpdateNodeAttr updates one or more attributes of a registered node
// This only updates the node if the status or lastError message changes
func (publisher *Publisher) UpdateNodeAttr(nodeID string, attrParams map[types.NodeAttr]string) (changed bool) {
	return publisher.registeredNodes.UpdateNodeAttr(nodeID, attrParams)
}

// UpdateNodeConfig updates a registered node's configuration and publishes the updated node.
//  If a config already exists then its value is retained but its configuration parameters are replaced.
//  Nodes are immutable. A new node is created and published and the old node instance is discarded.
func (publisher *Publisher) UpdateNodeConfig(nodeID string, attrName types.NodeAttr, configAttr *types.ConfigAttr) {
	publisher.registeredNodes.UpdateNodeConfig(nodeID, attrName, configAttr)
}

// UpdateNodeConfigValues updates the configuration values for the given registered node. This takes a map of
// key-value pairs with the configuration attribute name and new value. Intended for updating the
// node configuration based on what the registered node reports.
func (publisher *Publisher) UpdateNodeConfigValues(nodeID string, params types.NodeAttrMap) {
	publisher.registeredNodes.UpdateNodeConfigValues(nodeID, params)
}

// UpdateNodeStatus updates one or more status attributes of a registered node
// This only updates the node if the status or lastError message changes
func (publisher *Publisher) UpdateNodeStatus(nodeID string, status map[types.NodeStatus]string) (changed bool) {
	return publisher.registeredNodes.UpdateNodeStatus(nodeID, status)
}

// UpdateInput replaces the existing registered input with a new instance. Intended to update an
// input attribute.
func (publisher *Publisher) UpdateInput(input *types.InputDiscoveryMessage) {
	publisher.registeredInputs.UpdateInput(input)
}

// UpdateNode replaces the existing registered node with a new instance. Intended to update a
// node attribute.
func (publisher *Publisher) UpdateNode(node *types.NodeDiscoveryMessage) {
	publisher.registeredNodes.UpdateNode(node)
}

// UpdateOutput replaces a registered output with a new instance. Intended to update an
// output attribute.
func (publisher *Publisher) UpdateOutput(output *types.OutputDiscoveryMessage) {
	publisher.registeredOutputs.UpdateOutput(output)
}

// UpdateOutputValue adds the registered node's output value to the front of the value history
func (publisher *Publisher) UpdateOutputValue(nodeID string, outputType types.OutputType, instance string, newValue string) bool {
	outputAddr := outputs.MakeOutputDiscoveryAddress(publisher.Domain(), publisher.PublisherID(), nodeID, outputType, instance)
	return publisher.registeredOutputValues.UpdateOutputValue(outputAddr, newValue)
}
