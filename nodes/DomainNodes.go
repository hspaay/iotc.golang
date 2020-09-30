// Package nodes with domain node management
package nodes

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"

	"github.com/iotdomain/iotdomain-go/lib"
	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/sirupsen/logrus"
)

// DomainNodes manages nodes discovered on the domain
type DomainNodes struct {
	c             lib.DomainCollection     //
	messageSigner *messaging.MessageSigner // subscription to input discovery messages
}

// AddNode adds or replaces a discovered node
func (domainNodes *DomainNodes) AddNode(node *types.NodeDiscoveryMessage) {
	domainNodes.c.Update(node.Address, node)
}

// GetAllNodes returns a list of all discovered nodes of the domain
func (domainNodes *DomainNodes) GetAllNodes() []*types.NodeDiscoveryMessage {
	allNodes := make([]*types.NodeDiscoveryMessage, 0)
	domainNodes.c.GetAll(&allNodes)
	return allNodes
}

// GetPublisherNodes returns a list of all nodes of a publisher
// publisherAddress contains the domain/publisherID[/$identity]
func (domainNodes *DomainNodes) GetPublisherNodes(publisherAddress string) []*types.NodeDiscoveryMessage {
	var nodeList = make([]*types.NodeDiscoveryMessage, 0)
	domainNodes.c.GetByAddressPrefix(publisherAddress, &nodeList)
	return nodeList
}

// GetNodeByAddress returns a node by its  address using the domain, publisherID and nodeID
// address must contain the domain, publisherID and nodeID. Any other fields are ignored.
// Returns nil if address has no known node
func (domainNodes *DomainNodes) GetNodeByAddress(address string) *types.NodeDiscoveryMessage {
	var nodeObject = domainNodes.c.GetByAddress(address)
	if nodeObject == nil {
		return nil
	}
	return nodeObject.(*types.NodeDiscoveryMessage)
}

// GetNodeAttr returns a node attribute value
func (domainNodes *DomainNodes) GetNodeAttr(address string, attrName types.NodeAttr) string {
	var node = domainNodes.c.GetByAddress(address)
	if node == nil {
		return ""
	}
	attrValue, _ := node.(*types.NodeDiscoveryMessage).Attr[attrName]
	return attrValue
}

// GetNodeConfigValue returns the attribute value of a node in this list
// This returns the provided default value if no value is set and no default is configured.
// An error is returned when the node or configuration doesn't exist.
func (domainNodes *DomainNodes) GetNodeConfigValue(
	address string, attrName types.NodeAttr, defaultValue string) (value string, err error) {

	// in case of error, always return defaultValue
	node := domainNodes.GetNodeByAddress(address)
	if node == nil {
		msg := fmt.Sprintf("NodeList.GetNodeConfigValue: Node '%s' not found", address)
		return defaultValue, errors.New(msg)
	}
	config, configExists := node.Config[attrName]
	if !configExists {
		msg := fmt.Sprintf("NodeList.GetNodeConfigValue: Node '%s' configuration '%s' does not exist", address, attrName)
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

// LoadNodes loads saved discovered nodes from file
// Existing nodes are retained but replaced if contained in the file
func (domainNodes *DomainNodes) LoadNodes(filename string) error {
	nodeList := make([]*types.NodeDiscoveryMessage, 0)

	jsonNodes, err := ioutil.ReadFile(filename)
	if err != nil {
		return lib.MakeErrorf("LoadNodes: Unable to open file %s: %s", filename, err)
	}
	err = json.Unmarshal(jsonNodes, &nodeList)
	if err != nil {
		return lib.MakeErrorf("LoadNodes: Error parsing JSON node file %s: %v", filename, err)
	}
	logrus.Infof("LoadIdentities: Identities loaded successfully from %s", filename)
	for _, node := range nodeList {
		domainNodes.AddNode(node)
	}
	return nil

}

// RemoveNode removes a node using its address.
// If the node doesn't exist, this is ignored.
func (domainNodes *DomainNodes) RemoveNode(address string) {
	domainNodes.c.Remove(address)
}

// SaveNodes saves previously discovered nodes to file
func (domainNodes *DomainNodes) SaveNodes(filename string) error {
	collection := domainNodes.GetAllNodes()
	jsonText, err := json.MarshalIndent(collection, "", "  ")
	if err != nil {
		return lib.MakeErrorf("SaveNodes: Error Marshalling JSON collection '%s': %v", filename, err)
	}
	err = ioutil.WriteFile(filename, jsonText, 0664)
	if err != nil {
		return lib.MakeErrorf("SaveNodes: Error saving collection to JSON file %s: %v", filename, err)
	}
	logrus.Infof("SaveNodes: Collection saved successfully to JSON file %s", filename)
	return nil
}

// Subscribe to nodes discovery of the given domain publisher.
func (domainNodes *DomainNodes) Subscribe(domain string, publisherID string) {
	// subscription address  domain/publisher/+/$node
	address := MakeNodeDiscoveryAddress(domain, publisherID, "+")
	domainNodes.messageSigner.Subscribe(address, domainNodes.handleDiscoverNode)
}

// Unsubscribe from publisher
func (domainNodes *DomainNodes) Unsubscribe(domain string, publisherID string) {
	address := MakeNodeDiscoveryAddress(domain, publisherID, "+")
	domainNodes.messageSigner.Unsubscribe(address, domainNodes.handleDiscoverNode)
}

// handleDiscoverNode adds discovered domain nodes to the collection
func (domainNodes *DomainNodes) handleDiscoverNode(address string, message string) error {
	var discoMsg types.NodeDiscoveryMessage

	err := domainNodes.c.HandleDiscovery(address, message, &discoMsg)
	return err
}

// NewDomainNodes creates a new instance for domain node management.
//  messageSigner is used to receive signed node discovery messages
func NewDomainNodes(messageSigner *messaging.MessageSigner) *DomainNodes {
	domainCollection := lib.NewDomainCollection(
		reflect.TypeOf(&types.NodeDiscoveryMessage{}), messageSigner.GetPublicKey)

	domainNodes := DomainNodes{
		c:             domainCollection,
		messageSigner: messageSigner,
	}
	return &domainNodes
}
