// Package publisher with facade functions for nodes, inputs and outputs that work using nodeIDs
// instead of use of full addresses on the internal Nodes, Inputs and Outputs collections.
// Mostly intended to reduce boilerplate code in managing nodes, inputs and outputs
package publisher

import (
	"crypto/ecdsa"
	"errors"
	"fmt"

	"github.com/iotdomain/iotdomain-go/inputs"
	"github.com/iotdomain/iotdomain-go/nodes"
	"github.com/iotdomain/iotdomain-go/outputs"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/sirupsen/logrus"
)

// CreateInput creates a new node input that handle set commands and add it to the registered inputs
//  If an input of the given nodeID, type and instance already exist it will be replaced. This returns the new input
func (pub *Publisher) CreateInput(nodeID string, inputType types.InputType, instance string,
	handler func(inputAddress string, sender string, value string)) *types.InputDiscoveryMessage {
	input := pub.inputFromSetCommands.CreateInput(nodeID, inputType, instance, handler)
	return input
}

// CreateInputFromFile sends a file or folder to an input when it is modified
// The input handler is triggered with a message containing the path as value
func (pub *Publisher) CreateInputFromFile(nodeID string, inputType types.InputType, instance string,
	path string, handler func(inputAddress string, sender string, value string)) {
	// pub.fileInputs.LinkFileToInput(path, input)
	input := pub.registeredInputs.CreateInput(nodeID, inputType, instance, handler)
	// pub.fileInputs.SubscribeToFile(path, )
	_ = input
}

// CreateInputFromHTTP periodically polls an http address and sends the response to an input when it is modified
func (pub *Publisher) CreateInputFromHTTP(nodeID string, inputType types.InputType, instance string,
	url string, intervalSec int, handler func(inputAddress string, sender string, value string)) {
	input := pub.registeredInputs.CreateInput(nodeID, inputType, instance, handler)
	// pub.httpInputs.LinkToInput(path, path)
	_ = input
}

// CreateInputFromOutput subscribes to an output and triggers the input when a new value is received
func (pub *Publisher) CreateInputFromOutput(nodeID string, inputType types.InputType, instance string,
	outputAddress string, handler func(inputAddress string, sender string, value string)) {

	input := pub.registeredInputs.CreateInput(nodeID, inputType, instance, handler) // pub.httpInputs.LinkToInput(path, path)
	_ = input
}

// CreateNode creates a new node and add it to this publisher's discovered nodes
// returns the new node instance
func (pub *Publisher) CreateNode(nodeID string, nodeType types.NodeType) *types.NodeDiscoveryMessage {
	node := pub.registeredNodes.CreateNode(nodeID, nodeType)
	return node
}

// CreateOutput creates a new node output adds it to this publisher outputs list
// returns the output object to allow for easy updates
func (pub *Publisher) CreateOutput(nodeID string, outputType types.OutputType, instance string) *types.OutputDiscoveryMessage {
	output := pub.registeredOutputs.CreateOutput(nodeID, outputType, instance)
	return output
}

// GetDomainInput returns a discovered domain input
func (pub *Publisher) GetDomainInput(address string) *types.InputDiscoveryMessage {
	return pub.domainInputs.GetInputByAddress(address)
}

// GetDomainInputs returns all discovered domain inputs
func (pub *Publisher) GetDomainInputs() []*types.InputDiscoveryMessage {
	return pub.domainInputs.GetAllInputs()
}

// GetDomainNode returns a discovered domain node by its address
func (pub *Publisher) GetDomainNode(address string) *types.NodeDiscoveryMessage {
	return pub.domainNodes.GetNodeByAddress(address)
}

// GetDomainNodes returns all discovered domain nodes
func (pub *Publisher) GetDomainNodes() []*types.NodeDiscoveryMessage {
	return pub.domainNodes.GetAllNodes()
}

// GetDomainOutput returns a discovered domain output by its address
func (pub *Publisher) GetDomainOutput(address string) *types.OutputDiscoveryMessage {
	return pub.domainOutputs.GetOutputByAddress(address)
}

// GetDomainOutputs returns all discovered domain outputs
func (pub *Publisher) GetDomainOutputs() []*types.OutputDiscoveryMessage {
	return pub.domainOutputs.GetAllOutputs()
}

// GetDomainPublishers returns all discovered domain publishers
func (pub *Publisher) GetDomainPublishers() []*types.PublisherIdentityMessage {
	return pub.domainPublishers.GetAllPublishers()
}

// GetInput Get a registered input by node ID
func (pub *Publisher) GetInput(nodeID string, inputType types.InputType, instance string) *types.InputDiscoveryMessage {
	return pub.registeredInputs.GetInput(nodeID, inputType, instance)
}

// GetInputByAddress returns a registered input by its full address
func (pub *Publisher) GetInputByAddress(address string) *types.InputDiscoveryMessage {
	return pub.registeredInputs.GetInputByAddress(address)
}

// GetInputs returns a list of all registered inputs
func (pub *Publisher) GetInputs() []*types.InputDiscoveryMessage {
	return pub.registeredInputs.GetAllInputs()
}

// GetIdentity returns the publisher public identity including public signing key
func (pub *Publisher) GetIdentity() *types.PublisherIdentityMessage {
	return &pub.fullIdentity.PublisherIdentityMessage
}

// GetIdentityKeys returns the private/public key pair of this publisher
func (pub *Publisher) GetIdentityKeys() *ecdsa.PrivateKey {
	return pub.identityPrivateKey
}

// GetNodeAttr returns a node attribute value
func (pub *Publisher) GetNodeAttr(nodeID string, attrName types.NodeAttr) string {
	return pub.registeredNodes.GetNodeAttr(nodeID, attrName)
}

// GetNodeByAddress returns a registered node by its full address
func (pub *Publisher) GetNodeByAddress(address string) *types.NodeDiscoveryMessage {
	return pub.registeredNodes.GetNodeByAddress(address)
}

// GetNodeByID returns a node from this publisher or nil if the id isn't found in this publisher
func (pub *Publisher) GetNodeByID(nodeID string) (node *types.NodeDiscoveryMessage) {
	node = pub.registeredNodes.GetNodeByID(nodeID)
	return node
}

// GetNodeConfigBool returns a node configuration value as a boolean
// This retuns the given default if no configuration value exists and no configuration default is set
func (pub *Publisher) GetNodeConfigBool(
	nodeID string, attrName types.NodeAttr, defaultValue bool) (value bool, err error) {
	return pub.registeredNodes.GetNodeConfigBool(nodeID, attrName, defaultValue)
}

// GetNodeConfigFloat returns a node configuration value as a float number.
// This retuns the given default if no configuration value exists and no configuration default is set
func (pub *Publisher) GetNodeConfigFloat(
	nodeID string, attrName types.NodeAttr, defaultValue float32) (value float32, err error) {
	return pub.registeredNodes.GetNodeConfigFloat(nodeID, attrName, defaultValue)
}

// GetNodeConfigInt returns a node configuration value as an integer
// This retuns the given default if no configuration value exists and no configuration default is set
func (pub *Publisher) GetNodeConfigInt(
	nodeID string, attrName types.NodeAttr, defaultValue int) (value int, err error) {
	return pub.registeredNodes.GetNodeConfigInt(nodeID, attrName, defaultValue)
}

// GetNodeConfigString returns a node configuration value as a string
// This retuns the given default if no configuration value exists and no configuration default is set
func (pub *Publisher) GetNodeConfigString(
	nodeID string, attrName types.NodeAttr, defaultValue string) (value string, err error) {
	return pub.registeredNodes.GetNodeConfigString(nodeID, attrName, defaultValue)
}

// GetNodes returns a list of all registered nodes
func (pub *Publisher) GetNodes() []*types.NodeDiscoveryMessage {
	return pub.registeredNodes.GetAllNodes()
}

// GetNodeStatus returns a status attribute of a registered node
func (pub *Publisher) GetNodeStatus(nodeID string, attrName types.NodeStatus) (value string, exists bool) {
	node := pub.registeredNodes.GetNodeByID(nodeID)
	if node == nil {
		return "", false
	}
	value, exists = node.Status[attrName]
	return value, exists
}

// GetPublisherKey returns the public key of the publisher contained in the given address
// The address must at least contain a domain and publisherId
func (pub *Publisher) GetPublisherKey(address string) *ecdsa.PublicKey {
	return pub.domainPublishers.GetPublisherKey(address)
}

// GetOutput get a registered output
func (pub *Publisher) GetOutput(nodeID string, outputType types.OutputType, instance string) *types.OutputDiscoveryMessage {
	return pub.registeredOutputs.GetOutput(nodeID, outputType, instance)
}

// GetOutputs returns a list of all registered outputs
func (pub *Publisher) GetOutputs() []*types.OutputDiscoveryMessage {
	return pub.registeredOutputs.GetAllOutputs()
}

// GetOutputValue returns the registered output's value object including timestamp
func (pub *Publisher) GetOutputValue(nodeID string, outputType types.OutputType, instance string) *types.OutputValue {
	return pub.registeredOutputValues.GetOutputValueByType(nodeID, outputType, instance)
}

// MakeNodeDiscoveryAddress makes the node discovery address using the publisher domain and publisherID
func (pub *Publisher) MakeNodeDiscoveryAddress(nodeID string) string {
	addr := nodes.MakeNodeDiscoveryAddress(pub.Domain(), pub.PublisherID(), nodeID)
	return addr
}

// PublishConfigureNode publishes a $configure command to a domain node
// Returns true if successful, false if the domain node publisher cannot be found or has no public key
// and the message is not sent.
func (pub *Publisher) PublishConfigureNode(domainNodeAddr string, attr types.NodeAttrMap) bool {
	destPubKey := pub.GetPublisherKey(domainNodeAddr)
	if destPubKey == nil {
		logrus.Warnf("PublishConfigureNode: no public key found to encrypt command for node %s. Message not sent.", domainNodeAddr)
		return false
	}
	nodes.PublishNodeConfigure(domainNodeAddr, attr, pub.Address(), pub.messageSigner, destPubKey)
	return true
}

// PublishRaw immediately publishes the given value of a node, output type and instance on the
// $raw output address. The content can be signed but is not encrypted.
// This is intended for publishing large values that should not be stored, for example images
func (pub *Publisher) PublishRaw(output *types.OutputDiscoveryMessage, sign bool, value string) {
	outputs.PublishOutputRaw(output, value, pub.messageSigner)
}

// PublishSetInput publishes a $set input command to the given domain input address
// Returns error if the message cannot be sent.
func (pub *Publisher) PublishSetInput(domainInputAddr string, value string) error {
	destPubKey := pub.GetPublisherKey(domainInputAddr)
	if destPubKey == nil {
		errText := fmt.Sprintf("PublishSetInput: no public key found to encrypt command for set input to %s. Message not sent.", domainInputAddr)
		logrus.Warnf(errText)
		return errors.New(errText)
	}
	err := inputs.PublishSetInput(domainInputAddr, value, pub.Address(), pub.messageSigner, destPubKey)
	return err
}

// UpdateNodeErrorStatus sets a registered node RunState to the given status with a lasterror message
// Use NodeRunStateError for errors and NodeRunStateReady to clear error
// This only updates the node if the status or lastError message changes
func (pub *Publisher) UpdateNodeErrorStatus(nodeID string, status string, lastError string) {
	pub.registeredNodes.UpdateErrorStatus(nodeID, status, lastError)
}

// UpdateNodeAttr updates one or more attributes of a registered node
// This only updates the node if the status or lastError message changes
func (pub *Publisher) UpdateNodeAttr(nodeID string, attrParams map[types.NodeAttr]string) (changed bool) {
	return pub.registeredNodes.UpdateNodeAttr(nodeID, attrParams)
}

// UpdateNodeConfig updates a registered node's configuration and publishes the updated node.
//  If a config already exists then its value is retained but its configuration parameters are replaced.
//  Nodes are immutable. A new node is created and published and the old node instance is discarded.
func (pub *Publisher) UpdateNodeConfig(nodeID string, attrName types.NodeAttr, configAttr *types.ConfigAttr) {
	pub.registeredNodes.UpdateNodeConfig(nodeID, attrName, configAttr)
}

// UpdateNodeConfigValues updates the configuration values for the given registered node. This takes a map of
// key-value pairs with the configuration attribute name and new value. Intended for updating the
// node configuration based on what the registered node reports.
func (pub *Publisher) UpdateNodeConfigValues(nodeID string, params types.NodeAttrMap) {
	pub.registeredNodes.UpdateNodeConfigValues(nodeID, params)
}

// UpdateNodeStatus updates one or more status attributes of a registered node
// This only updates the node if the status or lastError message changes
func (pub *Publisher) UpdateNodeStatus(nodeID string, status map[types.NodeStatus]string) (changed bool) {
	return pub.registeredNodes.UpdateNodeStatus(nodeID, status)
}

// UpdateInput replaces the existing registered input with a new instance. Intended to update an
// input attribute.
// func (pub *Publisher) UpdateInput(input *types.InputDiscoveryMessage) {
// 	pub.registeredInputs.UpdateInput(input)
// }

// UpdateNode replaces the existing registered node with a new instance. Intended to update a
// node attribute.
// func (pub *Publisher) UpdateNode(node *types.NodeDiscoveryMessage) {
// 	pub.registeredNodes.UpdateNode(node)
// }

// // UpdateOutput replaces a registered output with a new instance. Intended to update an
// // output attribute. If the output does not exist, this is ignored.
// func (pub *Publisher) UpdateOutput(output *types.OutputDiscoveryMessage) {
// 	pub.registeredOutputs.UpdateOutput(output)
// }

// UpdateOutputForecast replaces a forecast
func (pub *Publisher) UpdateOutputForecast(nodeID string, outputType types.OutputType, instance string, forecast outputs.OutputForecast) {
	pub.registeredForecastValues.UpdateForecast(nodeID, outputType, instance, forecast)
}

// UpdateOutputValue adds the registered node's output value to the front of the value history
func (pub *Publisher) UpdateOutputValue(nodeID string, outputType types.OutputType, instance string, newValue string) bool {
	outputAddr := outputs.MakeOutputDiscoveryAddress(pub.Domain(), pub.PublisherID(), nodeID, outputType, instance)
	return pub.registeredOutputValues.UpdateOutputValue(outputAddr, newValue)
}
