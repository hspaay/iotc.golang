// Package nodes with domain node management
package nodes

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/iotdomain/iotdomain-go/lib"
	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/iotdomain/iotdomain-go/types"
)

// DomainNodes manages nodes discovered on the domain
type DomainNodes struct {
	c lib.DomainCollection //

	// nodeMap       map[string]*types.NodeDiscoveryMessage // registered nodes by node address
	// messageSigner *messaging.MessageSigner               // subscription to input discovery messages
	// updateMutex   *sync.Mutex                            // mutex for async updating of nodes
}

// AddNode adds or replaces a discovered node
func (domainNodes *DomainNodes) AddNode(node *types.NodeDiscoveryMessage) {
	domainNodes.c.Add(node.Address, node)
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

// RemoveNode removes a node using its address.
// If the node doesn't exist, this is ignored.
func (domainNodes *DomainNodes) RemoveNode(address string) {
	domainNodes.c.Remove(address)
}

// Start subscribing to node discovery
func (domainNodes *DomainNodes) Start() {
	// subscription address for all inputs domain/publisher/node/$node
	// TODO: Only subscribe to selected publishers
	address := MakeNodeDiscoveryAddress("+", "+", "+")
	domainNodes.c.MessageSigner.Subscribe(address, domainNodes.handleDiscoverNode)
}

// Stop polling for nodes
func (domainNodes *DomainNodes) Stop() {
	address := MakeNodeDiscoveryAddress("+", "+", "+")
	domainNodes.c.MessageSigner.Unsubscribe(address, domainNodes.handleDiscoverNode)
}

// getNode returns a node by its discovery address
// address must contain the domain, publisher and nodeID. Any other fields are ignored.
// Returns nil if address has no known node
// func (domainNodes *DomainNodes) getNode(address string) *types.NodeDiscoveryMessage {

// 	var node = domainNodes.c.Get(address, "", "")
// 	if node == nil {
// 		return nil
// 	}
// 	return node.(*types.NodeDiscoveryMessage)
// }

// handleDiscoverNode adds discovered domain nodes to the collection
func (domainNodes *DomainNodes) handleDiscoverNode(address string, message string) error {
	var discoMsg types.NodeDiscoveryMessage

	err := domainNodes.c.HandleDiscovery(address, message, &discoMsg)
	return err
}

// MakeNodeDiscoveryAddress generates the address of a node: domain/publisherID/nodeID/$node.
// Intended for lookup of nodes in the node list.
// func (domainNodes *DomainNodes) MakeNodeDiscoveryAddress(domain string, publisherID string, nodeID string) string {
// 	address := fmt.Sprintf("%s/%s/%s/%s", domain, publisherID, nodeID, types.MessageTypeNodeDiscovery)
// 	return address
// }

// NewDomainNodes creates a new instance for domain node management.
//  messageSigner is used to receive signed node discovery messages
func NewDomainNodes(messageSigner *messaging.MessageSigner) *DomainNodes {
	domainNodes := DomainNodes{
		c: lib.NewDomainCollection(messageSigner, reflect.TypeOf(&types.NodeDiscoveryMessage{})),
	}
	return &domainNodes
}
