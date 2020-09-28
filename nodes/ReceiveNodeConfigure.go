// Package nodes with handling of node configuration commands
package nodes

import (
	"crypto/ecdsa"
	"sync"

	"github.com/iotdomain/iotdomain-go/lib"
	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/sirupsen/logrus"
)

// NodeConfigureHandler application handler when command to update a node's configuration is received
// This returns a new map with configuration values that can be applied immediately.
type NodeConfigureHandler func(nodeAddress string, params types.NodeAttrMap) types.NodeAttrMap

// ReceiveNodeConfigure with handling of node configure commands aimed at nodes managed by this publisher.
// This decrypts incoming messages determines the sender and verifies the signature with
// the sender public key.
type ReceiveNodeConfigure struct {
	domain               string                   // the domain of this publisher
	publisherID          string                   // the registered publisher for the inputs
	nodeConfigureHandler NodeConfigureHandler     // handler to pass the command to
	messageSigner        *messaging.MessageSigner // subscription and publication messenger
	privateKey           *ecdsa.PrivateKey        // private key for decrypting set command messages
	registeredNodes      *RegisteredNodes         // registered nodes of this publisher
	updateMutex          *sync.Mutex              // mutex for async handling of inputs
}

// SetConfigureNodeHandler set the handler for updating node inputs
func (nodeConfigure *ReceiveNodeConfigure) SetConfigureNodeHandler(
	handler func(nodeAddress string, params types.NodeAttrMap) types.NodeAttrMap) {
	nodeConfigure.nodeConfigureHandler = handler
}

// Start listening for configure commands
func (nodeConfigure *ReceiveNodeConfigure) Start() {
	nodeConfigure.updateMutex.Lock()
	defer nodeConfigure.updateMutex.Unlock()
	// subscribe to all configure commands for this publisher's nodes
	addr := MakeNodeConfigureAddress(nodeConfigure.domain, nodeConfigure.publisherID, "+")
	nodeConfigure.messageSigner.Subscribe(addr, nodeConfigure.receiveConfigureCommand)
}

// Stop listening for commands
func (nodeConfigure *ReceiveNodeConfigure) Stop() {
	nodeConfigure.updateMutex.Lock()
	defer nodeConfigure.updateMutex.Unlock()
	addr := MakeNodeConfigureAddress(nodeConfigure.domain, nodeConfigure.publisherID, "+")
	nodeConfigure.messageSigner.Unsubscribe(addr, nodeConfigure.receiveConfigureCommand)
}

// handle an incoming a configuration command for one of our nodes. This:
// - check if the message is encrypted
// - check if the signature is valid
// - check if the node is valid
// - pass the configuration update to the actual handler
// - save node configuration if persistence is set
// TODO: support for authorization per node
func (nodeConfigure *ReceiveNodeConfigure) receiveConfigureCommand(nodeAddress string, message string) error {
	var configureMessage types.NodeConfigureMessage

	isEncrypted, isSigned, err := nodeConfigure.messageSigner.DecodeMessage(message, &configureMessage)

	if !isEncrypted {
		return lib.MakeErrorf("receiveConfigureCommand: Configuration update of '%s' is not encrypted. Message discarded.", nodeAddress)
	} else if !isSigned {
		return lib.MakeErrorf("receiveConfigureCommand: Configuration update of '%s' is not signed. Message discarded.", nodeAddress)
	} else if err != nil {
		return lib.MakeErrorf("receiveConfigureCommand: Message to %s. Error %s'. Message discarded.", nodeAddress, err)
	}

	// TODO: authorization check
	node := nodeConfigure.registeredNodes.GetNodeByAddress(nodeAddress)
	if node == nil || message == "" {
		return lib.MakeErrorf("receiveConfigureCommand unknown node for address %s or missing message", nodeAddress)
	}
	logrus.Infof("receiveConfigureCommand configure command on address %s. isEncrypted=%t, isSigned=%t", nodeAddress, isEncrypted, isSigned)

	params := configureMessage.Attr
	if nodeConfigure.nodeConfigureHandler != nil {
		// A handler can filter which configuration updates take place
		params = nodeConfigure.nodeConfigureHandler(nodeAddress, params)
	}
	// process the requested configuration, or ignore if none are applicable
	if params != nil {
		nodeConfigure.registeredNodes.UpdateNodeConfigValues(node.HWID, params)
	}
	return nil
}

// NewReceiveNodeConfigure returns a new instance of handling of node configuration commands.
func NewReceiveNodeConfigure(
	domain string,
	publisherID string,
	configHandler NodeConfigureHandler,
	messageSigner *messaging.MessageSigner,
	registeredNodes *RegisteredNodes,
	privateKey *ecdsa.PrivateKey) *ReceiveNodeConfigure {
	sin := &ReceiveNodeConfigure{
		domain:               domain,
		messageSigner:        messageSigner,
		nodeConfigureHandler: configHandler,
		publisherID:          publisherID,
		registeredNodes:      registeredNodes,
		privateKey:           privateKey,
		updateMutex:          &sync.Mutex{},
	}
	return sin
}
