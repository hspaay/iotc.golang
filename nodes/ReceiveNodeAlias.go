// Package nodes with receiving of the SetNodeAlias command
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

// NodeAliasHandler callback when command to change a node alias
type NodeAliasHandler func(address string, message *types.NodeAliasMessage)

// ReceiveNodeAlias listener
// This decrypts incoming messages, determines the sender and verifies the signature with
// the sender public key.
type ReceiveNodeAlias struct {
	domain          string                   // the domain of this publisher
	publisherID     string                   // the registered publisher for the inputs
	messageSigner   *messaging.MessageSigner // subscription and publication messenger
	privateKey      *ecdsa.PrivateKey        // private key for decrypting set command messages
	setAliasHandler NodeAliasHandler         // handler to pass the set command to
	updateMutex     *sync.Mutex              // mutex for async handling of inputs
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
func (setAlias *ReceiveNodeAlias) decodeAliasCommand(address string, message string) error {
	var aliasMessage types.NodeAliasMessage

	// Check that address is one of our inputs
	segments := strings.Split(address, "/")
	// a full address is required: domain/pub/node/$alias
	if len(segments) < 4 {
		return lib.MakeErrorf("decodeAliasCommand: address '%s' is incomplete", address)
	}
	// determine which node this message is for
	segments[3] = types.MessageTypeNodeDiscovery
	nodeAddr := strings.Join(segments, "/")

	isSigned, isEncrypted, err := setAlias.messageSigner.DecodeMessage(message, &aliasMessage)

	if !isEncrypted {
		return lib.MakeErrorf("decodeAliasCommand: Alias update of '%s' is not encrypted. Message discarded.", address)
	} else if !isSigned {
		return lib.MakeErrorf("decodeAliasCommand: Alias update of '%s' is not signed. Message discarded.", address)
	} else if err != nil {
		return lib.MakeErrorf("decodeAliasCommand: Message to %s. Error %s'. Message discarded.", address, err)
	}

	logrus.Infof("decodeAliasCommand on address %s. isEncrypted=%t, isSigned=%t", address, isEncrypted, isSigned)

	if setAlias.setAliasHandler != nil {
		setAlias.setAliasHandler(nodeAddr, &aliasMessage)
	} else {
		logrus.Errorf("decodeAliasCommand input command on address %s, but SetNodeInputHandler was used. Ignored.", address)
	}
	return nil
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
	privateKey *ecdsa.PrivateKey) *ReceiveNodeAlias {
	recvAlias := &ReceiveNodeAlias{
		domain:          domain,
		messageSigner:   messageSigner,
		setAliasHandler: setAliasHandler,
		publisherID:     publisherID,
		privateKey:      privateKey,
		updateMutex:     &sync.Mutex{},
	}
	return recvAlias
}
