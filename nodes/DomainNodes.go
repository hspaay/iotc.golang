// Package nodes with domain node management
package nodes

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/iotdomain/iotdomain-go/lib"
	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/iotdomain/iotdomain-go/types"
)

// DomainNodes manages nodes discovered on the domain
type DomainNodes struct {
	nodeMap       map[string]*types.NodeDiscoveryMessage // registered nodes by node address
	messageSigner *messaging.MessageSigner               // subscription to input discovery messages
	updateMutex   *sync.Mutex                            // mutex for async updating of nodes
}

// GetAllNodes returns a list of all discovered nodes of the domain
func (domainNodes *DomainNodes) GetAllNodes() []*types.NodeDiscoveryMessage {
	domainNodes.updateMutex.Lock()
	defer domainNodes.updateMutex.Unlock()

	var nodeList = make([]*types.NodeDiscoveryMessage, 0)
	for _, node := range domainNodes.nodeMap {
		nodeList = append(nodeList, node)
	}
	return nodeList
}

// GetPublisherNodes returns a list of all nodes of a publisher
func (domainNodes *DomainNodes) GetPublisherNodes(publisherID string) []*types.NodeDiscoveryMessage {
	domainNodes.updateMutex.Lock()
	defer domainNodes.updateMutex.Unlock()

	var nodeList = make([]*types.NodeDiscoveryMessage, 0)
	for _, node := range domainNodes.nodeMap {
		if node.PublisherID == publisherID {
			nodeList = append(nodeList, node)
		}
	}
	return nodeList
}

// GetNodeByAddress returns a node by its  address using the domain, publisherID and nodeID
// address must contain the domain, publisherID and nodeID. Any other fields are ignored.
// Returns nil if address has no known node
func (domainNodes *DomainNodes) GetNodeByAddress(address string) *types.NodeDiscoveryMessage {
	domainNodes.updateMutex.Lock()
	defer domainNodes.updateMutex.Unlock()

	var node = domainNodes.getNode(address)
	return node
}

// GetNodeAttr returns a node attribute value
func (domainNodes *DomainNodes) GetNodeAttr(address string, attrName types.NodeAttr) string {
	domainNodes.updateMutex.Lock()
	defer domainNodes.updateMutex.Unlock()
	var node = domainNodes.getNode(address)
	attrValue, _ := node.Attr[attrName]
	return attrValue
}

// GetNodeConfigValue returns the attribute value of a node in this list
// address must starts with the node's address: domain/publisher/nodeid. Any suffix is ignored.
// This retuns the provided default value if no value is set and no default is configured.
// An error is returned when the node or configuration doesn't exist.
func (domainNodes *DomainNodes) GetNodeConfigValue(
	address string, attrName types.NodeAttr, defaultValue string) (value string, err error) {

	domainNodes.updateMutex.Lock()
	defer domainNodes.updateMutex.Unlock()
	// in case of error, always return defaultValue
	node := domainNodes.getNode(address)
	if node == nil {
		msg := fmt.Sprintf("NodeList.GetNodeConfigString: Node '%s' not found", address)
		return defaultValue, errors.New(msg)
	}
	config, configExists := node.Config[attrName]
	if !configExists {
		msg := fmt.Sprintf("NodeList.GetNodeConfigString: Node '%s' configuration '%s' does not exist", address, attrName)
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

// Start subscribing to node discovery
func (domainNodes *DomainNodes) Start() {
	// subscription address for all inputs domain/publisher/node/$node
	// TODO: Only subscribe to selected publishers
	addr := MakeNodeDiscoveryAddress("+", "+", "+")
	domainNodes.messageSigner.Subscribe(addr, domainNodes.handleDiscoverNode)
}

// Stop polling for nodes
func (domainNodes *DomainNodes) Stop() {
	addr := MakeNodeDiscoveryAddress("+", "+", "+")
	domainNodes.messageSigner.Unsubscribe(addr, domainNodes.handleDiscoverNode)
}

// UpdateNode replaces a node or adds a new node based on node.Address.
//
// Intended to support Node immutability by making changes to a copy of a node and replacing
// the existing node with the updated node
// The updated node will be published
func (domainNodes *DomainNodes) UpdateNode(node *types.NodeDiscoveryMessage) {
	domainNodes.updateMutex.Lock()
	defer domainNodes.updateMutex.Unlock()
	domainNodes.nodeMap[node.Address] = node
}

// getNode returns a node by its discovery address
// address must contain the domain, publisher and nodeID. Any other fields are ignored.
// Returns nil if address has no known node
func (domainNodes *DomainNodes) getNode(address string) *types.NodeDiscoveryMessage {
	segments := strings.Split(address, "/")
	if len(segments) <= 3 {
		return nil
	}
	segments[3] = types.MessageTypeNodeDiscovery
	nodeAddr := strings.Join(segments[:4], "/")
	var node = domainNodes.nodeMap[nodeAddr]
	return node
}

// handleDiscoverNode adds discovered domain nodes to the collection
func (domainNodes *DomainNodes) handleDiscoverNode(address string, message string) error {
	var discoMsg types.NodeDiscoveryMessage

	// verify the message signature and get the payload
	_, err := domainNodes.messageSigner.VerifySignedMessage(message, &discoMsg)
	if err != nil {
		return lib.MakeErrorf("handleDiscoverNode: Failed verifying signature on address %s: %s", address, err)
	}
	segments := strings.Split(address, "/")
	discoMsg.PublisherID = segments[1]
	discoMsg.NodeID = segments[2]
	domainNodes.UpdateNode(&discoMsg)
	return nil
}

// MakeNodeDiscoveryAddress generates the address of a node: domain/publisherID/nodeID/$node.
// Intended for lookup of nodes in the node list.
func (domainNodes *DomainNodes) MakeNodeDiscoveryAddress(domain string, publisherID string, nodeID string) string {
	address := fmt.Sprintf("%s/%s/%s/%s", domain, publisherID, nodeID, types.MessageTypeNodeDiscovery)
	return address
}

// NewDomainNodes creates a new instance for domain node management.
//  messageSigner is used to receive signed node discovery messages
func NewDomainNodes(messageSigner *messaging.MessageSigner) *DomainNodes {
	domainNodes := DomainNodes{
		messageSigner: messageSigner,
		nodeMap:       make(map[string]*types.NodeDiscoveryMessage),
		updateMutex:   &sync.Mutex{},
	}
	return &domainNodes
}
