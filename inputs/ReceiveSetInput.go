// Package inputs with receiving of the SetInputMessage
package inputs

import (
	"crypto/ecdsa"
	"fmt"
	"strings"
	"sync"

	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/sirupsen/logrus"
)

// SetInputHandler callback when command to update input is received
type SetInputHandler func(address string, message *types.SetInputMessage)

// ReceiveSetInput with handling of input set commands aimed at inputs managed by a publisher.
// This decrypts incoming messages determines the sender and verifies the signature with
// the sender public key.
type ReceiveSetInput struct {
	domain           string                                // the domain of this publisher
	publisherID      string                                // the registered publisher for the inputs
	getPublisherKey  func(address string) *ecdsa.PublicKey // obtain the verification key of signatures
	messageSigner    *messaging.MessageSigner              // subscription and publication messenger
	privateKey       *ecdsa.PrivateKey                     // private key for decrypting set command messages
	registeredInputs *RegisteredInputs                     // registered inputs of this publisher
	setInputHandler  SetInputHandler                       // handler to pass the set command to
	updateMutex      *sync.Mutex                           // mutex for async handling of inputs
}

// SetNodeInputHandler set the handler for updating node inputs
func (setInputs *ReceiveSetInput) SetNodeInputHandler(handler func(address string, message *types.SetInputMessage)) {
	setInputs.setInputHandler = handler
}

// Start listening for set input commands
func (setInputs *ReceiveSetInput) Start() {
	setInputs.updateMutex.Lock()
	defer setInputs.updateMutex.Unlock()
	// subscribe to all set commands for inputs of this publisher nodes
	addr := MakeSetInputAddress(setInputs.domain, setInputs.publisherID, "+", "+", "+")
	setInputs.messageSigner.Subscribe(addr, setInputs.decodeSetCommand)
}

// Stop listening for set input command
func (setInputs *ReceiveSetInput) Stop() {
	setInputs.updateMutex.Lock()
	defer setInputs.updateMutex.Unlock()
	addr := MakeSetInputAddress(setInputs.domain, setInputs.publisherID, "+", "+", "+")
	setInputs.messageSigner.Unsubscribe(addr, setInputs.decodeSetCommand)
}

// decodeSetCommand decrypts and verifies the signature of an incoming set command.
// If successful this passes the set command to the setInputHandler callback
func (setInputs *ReceiveSetInput) decodeSetCommand(address string, message string) {
	var setMessage types.SetInputMessage

	// Check that address is one of our inputs
	segments := strings.Split(address, "/")
	// a full address is required
	if len(segments) < 6 {
		return
	}
	// domain/pub/node/inputtype/instance/$input
	segments[5] = types.MessageTypeInputDiscovery
	inputAddr := strings.Join(segments, "/")
	// input := sin.publisherInputs.GetInputByAddress(inputAddr)

	// if input == nil || message == "" {
	// 	sin.logger.Infof("handleNodeInput unknown input for address %s or missing message", address)
	// 	return
	// }

	// Decrypt the message if encrypted
	isEncrypted, dmessage, err := messaging.DecryptMessage(message, setInputs.privateKey)
	if !isEncrypted {
		logrus.Infof("decodeSetCommand: message to input '%s' is not encrypted.", address)
		// this could be fine, just warning
	} else if err != nil {
		logrus.Warnf("decodeSetCommand: decryption failed of message to input '%s'. Message discarded.", address)
		return
	}

	// Verify the message using the public key of the sender and decode the payload
	isSigned, err := messaging.VerifySignature(dmessage, &setMessage, setInputs.getPublisherKey)
	if !isSigned {
		// all inputs must use signed messages
		logrus.Warnf("decodeSetCommand: message to input '%s' is not signed. Message discarded.", address)
		return
	} else if err != nil {
		// signing failed, discard the message
		logrus.Warnf("decodeSetCommand: signature verification failed for message to input %s: %s. Message discarded.", address, err)
		return
	}

	logrus.Infof("decodeSetCommand on address %s. isEncrypted=%t, isSigned=%t", address, isEncrypted, isSigned)

	if setInputs.setInputHandler != nil {
		setInputs.setInputHandler(inputAddr, &setMessage)
	} else {
		logrus.Errorf("handleNodeInput input command on address %s, but SetNodeInputHandler was used. Ignored.", address)
	}
}

// MakeSetInputAddress creates the address used to update a node input value
// nodeAddress is an address containing the node.
func MakeSetInputAddress(domain string, publisherID string, nodeID string,
	inputType types.InputType, instance string) string {

	address := fmt.Sprintf("%s/%s/%s"+"/%s/%s/"+types.MessageTypeSet,
		domain, publisherID, nodeID, inputType, instance)
	return address
}

// NewReceiveSetInput returns a new instance of handling of set input commands.
func NewReceiveSetInput(
	domain string,
	publisherID string,
	inputHandler func(address string, message *types.SetInputMessage),
	messageSigner *messaging.MessageSigner,
	registeredInputs *RegisteredInputs,
	privateKey *ecdsa.PrivateKey,
	getPublisherKey func(addr string) *ecdsa.PublicKey) *ReceiveSetInput {
	recvsetin := &ReceiveSetInput{
		domain:           domain,
		getPublisherKey:  getPublisherKey,
		messageSigner:    messageSigner,
		setInputHandler:  inputHandler,
		publisherID:      publisherID,
		registeredInputs: registeredInputs,
		privateKey:       privateKey,
		updateMutex:      &sync.Mutex{},
	}
	return recvsetin
}
