// Package publisher with facade functions for nodes, inputs and outputs that work using nodeIDs
// instead of use of full addresses on the internal Nodes, Inputs and Outputs collections.
// Mostly intended to reduce boilerplate code in managing nodes, inputs and outputs
package publisher

import (
	"crypto/ecdsa"

	"github.com/iotdomain/iotdomain-go/inputs"
	"github.com/iotdomain/iotdomain-go/lib"
	"github.com/iotdomain/iotdomain-go/nodes"
	"github.com/iotdomain/iotdomain-go/outputs"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/sirupsen/logrus"
)

// Address returns the publisher's identity address
func (pub *Publisher) Address() string {
	// identityAddr := nodes.MakePublisherIdentityAddress(pub.Domain(), pub.PublisherID())
	// return identityAddr
	return pub.registeredIdentity.GetAddress()
}

// CreateInput creates a new node input that handle set commands and add it to the registered inputs
//  If an input of the given nodeHWID, type and instance already exist it will be replaced. This returns the new input
func (pub *Publisher) CreateInput(nodeHWID string, inputType types.InputType, instance string,
	setCommandHandler func(input *types.InputDiscoveryMessage, sender string, value string)) *types.InputDiscoveryMessage {
	input := pub.inputFromSetCommands.CreateInput(nodeHWID, inputType, instance, setCommandHandler)
	return input
}

// CreateInputFromFile sends a file or folder to an input when it is modified - TODO
// The input handler is triggered with a message containing the path as value
func (pub *Publisher) CreateInputFromFile(
	nodeHWID string, inputType types.InputType, instance string, path string,
	handler func(input *types.InputDiscoveryMessage, sender string, value string)) *types.InputDiscoveryMessage {

	input := pub.inputFromFiles.CreateInput(nodeHWID, inputType, instance, path, handler)
	return input
}

// CreateInputFromHTTP periodically polls an http address and sends the response to an input when it is modified
func (pub *Publisher) CreateInputFromHTTP(
	nodeHWID string, inputType types.InputType, instance string, url string, login string, password string, intervalSec int,
	handler func(input *types.InputDiscoveryMessage, sender string, value string)) {

	input := pub.inputFromHTTP.CreateHTTPInput(
		nodeHWID, inputType, instance, url, login, password, intervalSec, handler)
	_ = input
}

// CreateInputFromOutput subscribes to an output and triggers the input when a new value is received
func (pub *Publisher) CreateInputFromOutput(
	nodeHWID string, inputType types.InputType, instance string, outputAddress string,
	handler func(input *types.InputDiscoveryMessage, sender string, value string)) {

	input := pub.inputFromOutputs.CreateInput(nodeHWID, inputType, instance, outputAddress, handler)

	_ = input
}

// CreateNode creates a new node and add it to this publisher's registered nodes
// returns the new node instance
func (pub *Publisher) CreateNode(nodeHWID string, nodeType types.NodeType) *types.NodeDiscoveryMessage {
	node := pub.registeredNodes.CreateNode(nodeHWID, nodeType)
	return node
}

// CreateOutput creates a new node output adds it to this publisher outputs list
// returns the output object to allow for easy updates
func (pub *Publisher) CreateOutput(nodeHWID string, outputType types.OutputType,
	instance string) *types.OutputDiscoveryMessage {
	output := pub.registeredOutputs.CreateOutput(nodeHWID, outputType, instance)
	return output
}

// DeleteNode deletes a node from the collection of registered nodes
func (pub *Publisher) DeleteNode(hwAddress string) {
	pub.registeredNodes.DeleteNode(hwAddress)
}

// Domain returns the publication domain
func (pub *Publisher) Domain() string {
	ident, _ := pub.registeredIdentity.GetFullIdentity()
	return ident.Domain
}

// // FullIdentity return a copy of this publisher's full identity
// func (pub *Publisher) FullIdentity() types.PublisherFullIdentity {
// 	ident, _ := pub.registeredIdentity.GetIdentity()
// 	return *ident
// }

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
	return pub.domainIdentities.GetAllPublishers()
}

// GetInputByNodeHWID Get a registered input by its node HWID
func (pub *Publisher) GetInputByNodeHWID(nodeHWID string, inputType types.InputType, instance string) *types.InputDiscoveryMessage {
	return pub.registeredInputs.GetInputByNodeHWID(nodeHWID, inputType, instance)
}

// GetInputByAddress returns a registered input by its full address
func (pub *Publisher) GetInputByAddress(address string) *types.InputDiscoveryMessage {
	return pub.registeredInputs.GetInputByAddress(address)
}

// GetInputByID returns a registered input by its inputID
func (pub *Publisher) GetInputByID(inputID string) *types.InputDiscoveryMessage {
	return pub.registeredInputs.GetInputByID(inputID)
}

// GetInputs returns a list of all registered inputs
func (pub *Publisher) GetInputs() []*types.InputDiscoveryMessage {
	return pub.registeredInputs.GetAllInputs()
}

// GetIdentity returns the publisher public identity including public signing key
func (pub *Publisher) GetIdentity() *types.PublisherIdentityMessage {
	ident, _ := pub.registeredIdentity.GetFullIdentity()
	return &ident.PublisherIdentityMessage
}

// GetIdentityKeys returns the private/public key pair of this publisher
func (pub *Publisher) GetIdentityKeys() *ecdsa.PrivateKey {
	_, privKey := pub.registeredIdentity.GetFullIdentity()
	return privKey
}

// GetNodeAttr returns a node attribute value
func (pub *Publisher) GetNodeAttr(nodeHWID string, attrName types.NodeAttr) string {
	return pub.registeredNodes.GetNodeAttr(nodeHWID, attrName)
}

// GetNodeByAddress returns a registered node by its full address using its nodeID
func (pub *Publisher) GetNodeByAddress(address string) *types.NodeDiscoveryMessage {
	return pub.registeredNodes.GetNodeByAddress(address)
}

// GetNodeByHWID returns a node from this publisher or nil if the nodeHWID isn't found
func (pub *Publisher) GetNodeByHWID(nodeHWID string) (node *types.NodeDiscoveryMessage) {
	node = pub.registeredNodes.GetNodeByHWID(nodeHWID)
	return node
}

// GetNodeByNodeID returns a node from this publisher or nil if the node id isn't found
func (pub *Publisher) GetNodeByNodeID(nodeID string) (node *types.NodeDiscoveryMessage) {
	node = pub.registeredNodes.GetNodeByNodeID(nodeID)
	return node
}

// GetNodeConfigBool returns a node configuration value as a boolean
// This retuns the given default if no configuration value exists and no configuration default is set
func (pub *Publisher) GetNodeConfigBool(
	nodeHWID string, attrName types.NodeAttr, defaultValue bool) (value bool, err error) {
	return pub.registeredNodes.GetNodeConfigBool(nodeHWID, attrName, defaultValue)
}

// GetNodeConfigFloat returns a node configuration value as a float number.
// This retuns the given default if no configuration value exists and no configuration default is set
func (pub *Publisher) GetNodeConfigFloat(
	nodeHWID string, attrName types.NodeAttr, defaultValue float32) (value float32, err error) {
	return pub.registeredNodes.GetNodeConfigFloat(nodeHWID, attrName, defaultValue)
}

// GetNodeConfigInt returns a node configuration value as an integer
// This retuns the given default if no configuration value exists and no configuration default is set
func (pub *Publisher) GetNodeConfigInt(
	nodeHWID string, attrName types.NodeAttr, defaultValue int) (value int, err error) {
	return pub.registeredNodes.GetNodeConfigInt(nodeHWID, attrName, defaultValue)
}

// GetNodeConfigString returns a node configuration value as a string
// This retuns the given default if no configuration value exists and no configuration default is set
func (pub *Publisher) GetNodeConfigString(
	nodeHWID string, attrName types.NodeAttr, defaultValue string) (value string, err error) {
	return pub.registeredNodes.GetNodeConfigString(nodeHWID, attrName, defaultValue)
}

// GetNodes returns a list of all registered nodes
func (pub *Publisher) GetNodes() []*types.NodeDiscoveryMessage {
	return pub.registeredNodes.GetAllNodes()
}

// GetNodeStatus returns a status attribute of a registered node
func (pub *Publisher) GetNodeStatus(nodeHWID string, attrName types.NodeStatus) (value string, exists bool) {
	node := pub.registeredNodes.GetNodeByHWID(nodeHWID)
	if node == nil {
		return "", false
	}
	value, exists = node.Status[attrName]
	return value, exists
}

// GetOutputByNodeHWID get a registered output by node HWID
func (pub *Publisher) GetOutputByNodeHWID(nodeHWID string, outputType types.OutputType, instance string) *types.OutputDiscoveryMessage {
	return pub.registeredOutputs.GetOutputByNodeHWID(nodeHWID, outputType, instance)
}

// GetOutputByID gets a registered output by its outputID
func (pub *Publisher) GetOutputByID(outputID string) *types.OutputDiscoveryMessage {
	return pub.registeredOutputs.GetOutputByID(outputID)
}

// GetOutputs returns a list of all registered outputs
func (pub *Publisher) GetOutputs() []*types.OutputDiscoveryMessage {
	return pub.registeredOutputs.GetAllOutputs()
}

// GetOutputValueByNodeHWID returns the registered output's value object including timestamp
func (pub *Publisher) GetOutputValueByNodeHWID(nodeHWID string, outputType types.OutputType, instance string) *types.OutputValue {
	return pub.registeredOutputValues.GetOutputValueByType(nodeHWID, outputType, instance)
}

// GetOutputValueByID returns the registered output's value object including timestamp
func (pub *Publisher) GetOutputValueByID(outputID string) *types.OutputValue {
	return pub.registeredOutputValues.GetOutputValueByID(outputID)
}

// GetPublisherKey returns the public key of the publisher contained in the given address
// The address must at least contain a domain and publisherId
func (pub *Publisher) GetPublisherKey(address string) *ecdsa.PublicKey {
	return pub.domainIdentities.GetPublisherKey(address)
}

// MakeNodeDiscoveryAddress makes the node discovery address using the publisher domain and publisherID
func (pub *Publisher) MakeNodeDiscoveryAddress(nodeID string) string {
	addr := nodes.MakeNodeDiscoveryAddress(pub.Domain(), pub.PublisherID(), nodeID)
	return addr
}

// PublisherID returns the publisher's ID
func (pub *Publisher) PublisherID() string {
	ident, _ := pub.registeredIdentity.GetFullIdentity()
	return ident.PublisherID
}

// PublishNodeConfigure publishes a $configure command to a domain node
// Returns true if successful, false if the domain node publisher cannot be found or has no public key
// and the message is not sent.
func (pub *Publisher) PublishNodeConfigure(domainNodeAddr string, attr types.NodeAttrMap) bool {
	destPubKey := pub.GetPublisherKey(domainNodeAddr)
	if destPubKey == nil {
		logrus.Warnf("PublishConfigureNode: no public key found to encrypt command for node %s. Message not sent.", domainNodeAddr)
		return false
	}
	nodes.PublishNodeConfigure(domainNodeAddr, attr, pub.Address(), pub.messageSigner, destPubKey)
	return true
}

// // PublishNodeAlias publishes a command to set a node's alias
// // The node's publisher must have been discovered
// func (pub *Publisher) PublishNodeAlias(nodeAddr string, alias string) {
// 	pubKey := pub.domainIdentities.GetPublisherKey(nodeAddr)
// 	nodes.PublishNodeAliasCommand(nodeAddr, alias, pub.Address(), pub.messageSigner, pubKey)
// }

// PublishRaw immediately publishes the given value of a node, output type and instance on the
// $raw output address. The content can be signed but is not encrypted.
// This is intended for publishing large values that should not be stored, for example images
func (pub *Publisher) PublishRaw(output *types.OutputDiscoveryMessage, sign bool, value string) {
	outputs.PublishOutputRaw(output, value, pub.messageSigner)
}

// PublishOutputEvent publishes all outputs of the node in a single event
func (pub *Publisher) PublishOutputEvent(node *types.NodeDiscoveryMessage) error {
	return PublishOutputEvent(node, pub.registeredOutputs, pub.registeredOutputValues, pub.messageSigner)
}

// PublishSetInput publishes a $setInput input command to the given input address
//  This requires that the publisher identity of the receiving input is known so the
// command can be encrypted.
// Returns error if the destination publisher is unknown and the message cannot be sent.
func (pub *Publisher) PublishSetInput(inputAddr string, value string) error {
	destPubKey := pub.GetPublisherKey(inputAddr)
	if destPubKey == nil {
		return lib.MakeErrorf("PublishSetInput: no public key found to encrypt command for set input to %s. Message not sent.", inputAddr)
	}
	err := inputs.PublishSetInput(inputAddr, value, pub.Address(), pub.messageSigner, destPubKey)
	return err
}

// PublishSetNodeID publishes a set node ID command to the given node address
//  This requires that the publisher identity of the receiving input is known so the
// command can be encrypted.
// Returns error if the destination publisher is unknown and the message cannot be sent.
func (pub *Publisher) PublishSetNodeID(nodeAddr string, newNodeID string) error {
	destPubKey := pub.GetPublisherKey(nodeAddr)
	if destPubKey == nil {
		return lib.MakeErrorf("PublishSetNodeID: no public key found to encrypt command for set input to %s"+
			". Message not sent.", nodeAddr)
	}
	err := nodes.PublishSetNodeID(nodeAddr, newNodeID, pub.Address(), pub.messageSigner, destPubKey)
	return err
}

// SetSigningOnOff turns signing of publications on or off.
//  The default is on (true)
func (pub *Publisher) SetSigningOnOff(onOff bool) {
	pub.messageSigner.SetSignMessages(onOff)
}

// Subscribe to receive nodes, inputs and outputs from the selected domain and/or publisher
// To subscribe to all domains or all publishers use "" as the domain or publisherID
func (pub *Publisher) Subscribe(domain string, publisherID string) {
	// subscription address for all outputs domain/publisher/node/type/instance/$output
	if domain == "" {
		domain = "+"
	}
	if publisherID == "" {
		publisherID = "+"
	}
	pub.domainNodes.Subscribe(domain, publisherID)
	pub.domainInputs.Subscribe(domain, publisherID)
	pub.domainOutputs.Subscribe(domain, publisherID)
}

// Unsubscribe from receiving nodes, inputs and outputs from the selected domain and/or publisher
// Use the same domain and publisherID as used in Subscribe
func (pub *Publisher) Unsubscribe(domain string, publisherID string) {
	// subscription address for all outputs domain/publisher/node/type/instance/$output
	if domain == "" {
		domain = "+"
	}
	if publisherID == "" {
		publisherID = "+"
	}
	pub.domainNodes.Unsubscribe(domain, publisherID)
	pub.domainInputs.Unsubscribe(domain, publisherID)
	pub.domainOutputs.Unsubscribe(domain, publisherID)
}

// UpdateNodeErrorStatus sets a registered node RunState to the given status with a lasterror message
// Use NodeRunStateError for errors and NodeRunStateReady to clear error
// This only updates the node if the status or lastError message changes
func (pub *Publisher) UpdateNodeErrorStatus(nodeHWID string, status string, lastError string) {
	pub.registeredNodes.UpdateErrorStatus(nodeHWID, status, lastError)
}

// UpdateNodeAttr updates one or more attributes of a registered node
// This only updates the node if the status or lastError message changes
func (pub *Publisher) UpdateNodeAttr(nodeHWID string, attrParams types.NodeAttrMap) (changed bool) {
	return pub.registeredNodes.UpdateNodeAttr(nodeHWID, attrParams)
}

// UpdateNodeConfig updates a registered node's configuration and publishes the updated node.
//  If a config already exists then its value is retained but its configuration parameters are replaced.
//  Nodes are immutable. A new node is created and published and the old node instance is discarded.
func (pub *Publisher) UpdateNodeConfig(nodeHWID string, attrName types.NodeAttr, configAttr *types.ConfigAttr) {
	pub.registeredNodes.UpdateNodeConfig(nodeHWID, attrName, configAttr)
}

// UpdateNodeConfigValues updates the configuration values for the given registered node. This takes a map of
// key-value pairs with the configuration attribute name and new value. Intended for updating the
// node configuration based on what the registered node reports.
func (pub *Publisher) UpdateNodeConfigValues(nodeHWID string, params types.NodeAttrMap) (changed bool) {
	return pub.registeredNodes.UpdateNodeConfigValues(nodeHWID, params)
}

// UpdateNodeStatus updates one or more status attributes of a registered node
// This only updates the node if the status changes
func (pub *Publisher) UpdateNodeStatus(nodeHWID string, status map[types.NodeStatus]string) (changed bool) {
	return pub.registeredNodes.UpdateNodeStatus(nodeHWID, status)
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

// UpdateOutput replaces a registered output with a new instance. Intended to update an
// output attribute. If the output does not exist, this is ignored.
func (pub *Publisher) UpdateOutput(output *types.OutputDiscoveryMessage) {
	pub.registeredOutputs.UpdateOutput(output)
}

// UpdateOutputForecast replaces a forecast
func (pub *Publisher) UpdateOutputForecast(outputID string, forecast outputs.OutputForecast) {
	pub.registeredForecastValues.UpdateForecast(outputID, forecast)
}

// UpdateOutputValue adds the registered node's output value to the front of the value history
func (pub *Publisher) UpdateOutputValue(nodeHWID string, outputType types.OutputType, instance string, newValue string) bool {
	outputID := outputs.MakeOutputID(nodeHWID, outputType, instance)
	return pub.registeredOutputValues.UpdateOutputValue(outputID, newValue)
}
