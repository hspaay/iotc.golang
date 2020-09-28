// Package nodes with receiving of the SetNodeId command
package nodes

import (
	"crypto/ecdsa"
	"fmt"
	"strings"
	"sync"

	"github.com/iotdomain/iotdomain-go/lib"
	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/sirupsen/logrus"
)

// SetNodeIDHandler callback when command to change a node ID
type SetNodeIDHandler func(nodeAddress string, message *types.SetNodeIDMessage)

// ReceiveSetNodeID listener
// This decrypts incoming messages, determines the sender and verifies the signature with
// the sender public key.
type ReceiveSetNodeID struct {
	domain        string                   // the domain of this publisher
	publisherID   string                   // the registered publisher for the inputs
	messageSigner *messaging.MessageSigner // subscription and publication messenger
	privateKey    *ecdsa.PrivateKey        // private key for decrypting set command messages
	handler       SetNodeIDHandler         // handler to pass the command to
	updateMutex   *sync.Mutex              // mutex for async handling of inputs
}

// SetNodeIDHandler set the handler for updating node IDs
func (setNodeID *ReceiveSetNodeID) SetNodeIDHandler(
	handler func(nodeAddress string, message *types.SetNodeIDMessage)) {

	setNodeID.handler = handler
}

// Start listening for set node ID commands
func (setNodeID *ReceiveSetNodeID) Start() {
	setNodeID.updateMutex.Lock()
	defer setNodeID.updateMutex.Unlock()
	addr := MakeSetNodeIDAddress(setNodeID.domain, setNodeID.publisherID, "+")
	setNodeID.messageSigner.Subscribe(addr, setNodeID.decodeSetNodeIDCommand)
}

// Stop listening for set node ID commands
func (setNodeID *ReceiveSetNodeID) Stop() {
	setNodeID.updateMutex.Lock()
	defer setNodeID.updateMutex.Unlock()
	addr := MakeSetNodeIDAddress(setNodeID.domain, setNodeID.publisherID, "+")
	setNodeID.messageSigner.Unsubscribe(addr, setNodeID.decodeSetNodeIDCommand)
}

// decodeSetNodeIDCommand decrypts and verifies the signature of an incoming set command.
// If successful this passes the set command to the setInputHandler callback
func (setNodeID *ReceiveSetNodeID) decodeSetNodeIDCommand(setAddress string, message string) error {
	var setNodeIDMessage types.SetNodeIDMessage

	// Check that address is one of our inputs
	segments := strings.Split(setAddress, "/")
	// a full address is required: domain/pub/node/$setNodeId
	if len(segments) < 4 {
		return lib.MakeErrorf("decodeSetNodeIDCommand: address '%s' is incomplete", setAddress)
	}
	// determine which node this message is for
	segments[3] = types.MessageTypeNodeDiscovery
	nodeAddr := strings.Join(segments, "/")

	isEncrypted, isSigned, err := setNodeID.messageSigner.DecodeMessage(message, &setNodeIDMessage)

	if !isEncrypted {
		return lib.MakeErrorf("decodeSetNodeIDCommand: Update of '%s' is not encrypted. Message discarded.", setAddress)
	} else if !isSigned {
		return lib.MakeErrorf("decodeSetNodeIDCommand: Update of '%s' is not signed. Message discarded.", setAddress)
	} else if err != nil {
		return lib.MakeErrorf("decodeSetNodeIDCommand: Message to %s. Error %s'. Message discarded.", setAddress, err)
	}

	logrus.Infof("decodeSetNodeIDCommand on address %s. isEncrypted=%t, isSigned=%t", setAddress, isEncrypted, isSigned)

	if setNodeID.handler != nil {
		setNodeID.handler(nodeAddr, &setNodeIDMessage)
	} else {
		logrus.Errorf("decodeSetNodeIDCommand: command received on address %s, but no handler is configured.", setAddress)
	}
	return nil
}

// MakeSetNodeIDAddress creates the address used to update a node's ID
// domain, publisherID, nodeID of the existing node
func MakeSetNodeIDAddress(domain string, publisherID string, nodeID string) string {
	address := fmt.Sprintf("%s/%s/%s/"+types.MessageTypeSetNodeID, domain, publisherID, nodeID)
	return address
}

// NewReceiveSetNodeID returns a new instance of handling of the setNodeId command.
func NewReceiveSetNodeID(
	domain string,
	publisherID string,
	setNodeIDHandler func(address string, message *types.SetNodeIDMessage),
	messageSigner *messaging.MessageSigner,
	privateKey *ecdsa.PrivateKey) *ReceiveSetNodeID {
	receiver := &ReceiveSetNodeID{
		domain:        domain,
		messageSigner: messageSigner,
		handler:       setNodeIDHandler,
		publisherID:   publisherID,
		privateKey:    privateKey,
		updateMutex:   &sync.Mutex{},
	}
	return receiver
}
