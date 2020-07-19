// Package nodes with receiving of the SetNodeAlias command
package nodes

import (
	"crypto/ecdsa"
	"fmt"
	"strings"
	"sync"

	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/sirupsen/logrus"
)

// NodeAliasHandler callback when command to change a node alias
type NodeAliasHandler func(address string, message *types.NodeAliasMessage)

// ReceiveNodeAlias listener
// This decrypts incoming messages, determines the sender and verifies the signature with
// the sender public key.
type ReceiveNodeAlias struct {
	domain          string                                // the domain of this publisher
	publisherID     string                                // the registered publisher for the inputs
	getPublisherKey func(address string) *ecdsa.PublicKey // obtain the verification key of signatures
	messageSigner   *messaging.MessageSigner              // subscription and publication messenger
	privateKey      *ecdsa.PrivateKey                     // private key for decrypting set command messages
	setAliasHandler NodeAliasHandler                      // handler to pass the set command to
	updateMutex     *sync.Mutex                           // mutex for async handling of inputs
}

// SetAliasHandler set the handler for updating node inputs
func (setAlias *ReceiveNodeAlias) SetAliasHandler(handler func(address string, message *types.NodeAliasMessage)) {
	setAlias.setAliasHandler = handler
}

// Start listening for node alias commands
func (setAlias *ReceiveNodeAlias) Start() {
	setAlias.updateMutex.Lock()
	defer setAlias.updateMutex.Unlock()
	addr := MakeAliasAddress(setAlias.domain, setAlias.publisherID, "+")
	setAlias.messageSigner.Subscribe(addr, setAlias.decodeAliasCommand)
}

// Stop listening for alias input command
func (setAlias *ReceiveNodeAlias) Stop() {
	setAlias.updateMutex.Lock()
	defer setAlias.updateMutex.Unlock()
	addr := MakeAliasAddress(setAlias.domain, setAlias.publisherID, "+")
	setAlias.messageSigner.Unsubscribe(addr, setAlias.decodeAliasCommand)
}

// decodeSetCommand decrypts and verifies the signature of an incoming set command.
// If successful this passes the set command to the setInputHandler callback
func (setAlias *ReceiveNodeAlias) decodeAliasCommand(address string, message string) {
	var aliasMessage types.NodeAliasMessage

	// Check that address is one of our inputs
	segments := strings.Split(address, "/")
	// a full address is required
	if len(segments) < 6 {
		return
	}
	// determine which node this message is for
	segments[3] = types.MessageTypeNodeDiscovery
	nodeAddr := strings.Join(segments, "/")

	// Decrypt the message if encrypted
	isEncrypted, dmessage, err := messaging.DecryptMessage(message, setAlias.privateKey)
	if !isEncrypted {
		logrus.Infof("decodeAliasCommand: message to input '%s' is not encrypted.", address)
		// this could be fine, just warning
	} else if err != nil {
		logrus.Warnf("decodeAliasCommand: decryption failed of message to input '%s'. Message discarded.", address)
		return
	}

	// Verify the message using the public key of the sender and decode the payload
	isSigned, err := messaging.VerifySenderSignature(dmessage, &aliasMessage, setAlias.getPublisherKey)
	if !isSigned {
		// all inputs must use signed messages
		logrus.Warnf("decodeAliasCommand: message to input '%s' is not signed. Message discarded.", address)
		return
	} else if err != nil {
		// signing failed, discard the message
		logrus.Warnf("decodeAliasCommand: signature verification failed for message to input %s: %s. Message discarded.", address, err)
		return
	}

	logrus.Infof("decodeAliasCommand on address %s. isEncrypted=%t, isSigned=%t", address, isEncrypted, isSigned)

	if setAlias.setAliasHandler != nil {
		setAlias.setAliasHandler(nodeAddr, &aliasMessage)
	} else {
		logrus.Errorf("decodeAliasCommand input command on address %s, but SetNodeInputHandler was used. Ignored.", address)
	}
}

// MakeAliasAddress creates the address used to update a node's alias
// nodeAddress is an address containing the node.
func MakeAliasAddress(domain string, publisherID string, nodeID string) string {

	address := fmt.Sprintf("%s/%s/%s/"+types.MessageTypeNodeAlias, domain, publisherID, nodeID)
	return address
}

// NewReceiveNodeAlias returns a new instance of handling of the alias command.
func NewReceiveNodeAlias(
	domain string,
	publisherID string,
	setAliasHandler func(address string, message *types.NodeAliasMessage),
	messageSigner *messaging.MessageSigner,
	privateKey *ecdsa.PrivateKey,
	getPublisherKey func(addr string) *ecdsa.PublicKey) *ReceiveNodeAlias {
	recvAlias := &ReceiveNodeAlias{
		domain:          domain,
		getPublisherKey: getPublisherKey,
		messageSigner:   messageSigner,
		setAliasHandler: setAliasHandler,
		publisherID:     publisherID,
		privateKey:      privateKey,
		updateMutex:     &sync.Mutex{},
	}
	return recvAlias
}
