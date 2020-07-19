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

// InputFromSetCommands handles set commands aimed at inputs managed by this publisher.
// This decrypts incoming messages determines the sender and verifies the signature with
// the sender public key.
type InputFromSetCommands struct {
	domain           string // the domain of this publisher
	publisherID      string // the registered publisher for the inputs
	isRunning        bool
	messageSigner    *messaging.MessageSigner // subscription and publication messenger
	privateKey       *ecdsa.PrivateKey        // private key for decrypting set command messages
	senderTimestamp  map[string]string        // most recent timestamp of received commands by sender
	registeredInputs *RegisteredInputs        // registered inputs of this publisher
	// subscriptions of registered inputs
	subscriptions map[string]string // SetInput subscriptions of inputs [setAddr]setAddr
	updateMutex   *sync.Mutex       // mutex for async handling of inputs
}

// CreateInput creates a new input that responds to a set command from the message bus.
// If an input of the given nodeID, type and instance already exist it will be replaced.
// This returns the new input
func (ifset *InputFromSetCommands) CreateInput(
	nodeID string, inputType types.InputType, instance string,
	handler func(inputAddress string, sender string, value string)) *types.InputDiscoveryMessage {

	ifset.updateMutex.Lock()
	defer ifset.updateMutex.Unlock()

	input := ifset.registeredInputs.CreateInput(nodeID, inputType, instance, handler)
	ifset.subscribeToSetCommand(nodeID, inputType, instance)
	return input
}

// DeleteInput deletes the input and unsubscribes to the input's set command
func (ifset *InputFromSetCommands) DeleteInput(nodeID string, inputType types.InputType, instance string) {
	ifset.updateMutex.Lock()
	defer ifset.updateMutex.Unlock()

	ifset.unsubscribeFromSetCommand(nodeID, inputType, instance)
	ifset.registeredInputs.DeleteInput(nodeID, inputType, instance)
}

// decodeSetCommand decrypts and verifies the signature of an incoming set command.
// If successful this passes the set command to the setInputHandler callback
func (ifset *InputFromSetCommands) decodeSetCommand(address string, message string) {
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

	// Decrypt the message if encryption is enabled
	var dmessage string
	var isEncrypted bool
	var err error
	if ifset.privateKey != nil {
		isEncrypted, dmessage, err = messaging.DecryptMessage(message, ifset.privateKey)
		if !isEncrypted {
			logrus.Infof("decodeSetCommand: message to input '%s' is not encrypted. Message discarded.", address)
			return
		} else if err != nil {
			logrus.Warnf("decodeSetCommand: decryption failed of message to input '%s'. Message discarded.", address)
			return
		}
	}
	// Verify the message using the public key of the sender and decode the payload
	// This fails if decryption is disabled and the message is encrypted
	isSigned, err := ifset.messageSigner.VerifySignedMessage(dmessage, &setMessage)
	if !isSigned {
		// all inputs must use signed messages
		logrus.Warnf("decodeSetCommand: message to input '%s' is not signed. Message discarded.", address)
		return
	} else if err != nil {
		// signing failed, discard the message
		logrus.Warnf("decodeSetCommand: signature verification failed for message to input %s: %s. Message discarded.", address, err)
		return
	}
	// Verify this is the most recent message to protect against replay attacks
	prevTimestamp := ifset.senderTimestamp[setMessage.Sender]
	if prevTimestamp > setMessage.Timestamp {
		logrus.Warnf("decodeSetCommand: earlier timestamp of message to input %s from sender %s. Message discarded.", address, setMessage.Sender)
		return
	}
	ifset.senderTimestamp[setMessage.Sender] = setMessage.Timestamp
	logrus.Infof("decodeSetCommand successful for input %s. isEncrypted=%t, isSigned=%t", address, isEncrypted, isSigned)

	// the handler is responsible for authorization
	ifset.registeredInputs.NotifyInputHandler(inputAddr, setMessage.Sender, setMessage.Value)
}

// subscribeToSetCommand to receive set input commands for the given node, type and instance
func (ifset *InputFromSetCommands) subscribeToSetCommand(nodeID string, inputType types.InputType, instance string) {
	setAddr := MakeSetInputAddress(ifset.domain, ifset.publisherID, nodeID, inputType, instance)
	// prevent double subscription
	_, hasSubscription := ifset.subscriptions[setAddr]
	if !hasSubscription {
		ifset.subscriptions[setAddr] = setAddr
		ifset.messageSigner.Subscribe(setAddr, ifset.decodeSetCommand)
	}
}

// unsubscribeFromSetCommand removes previous subscription
func (ifset *InputFromSetCommands) unsubscribeFromSetCommand(nodeID string, inputType types.InputType, instance string) {
	setAddr := MakeSetInputAddress(ifset.domain, ifset.publisherID, nodeID, inputType, instance)
	_, hasSubscription := ifset.subscriptions[setAddr]
	if hasSubscription {
		delete(ifset.subscriptions, setAddr)
		ifset.messageSigner.Unsubscribe(setAddr, ifset.decodeSetCommand)
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

// NewInputFromSetCommands returns a new instance of handling of set input commands.
// The private key is used to decrypt set commands. Without it, decryption is disabled.
func NewInputFromSetCommands(
	domain string,
	publisherID string,
	messageSigner *messaging.MessageSigner,
	registeredInputs *RegisteredInputs,
	privateKey *ecdsa.PrivateKey) *InputFromSetCommands {

	recvsetin := &InputFromSetCommands{
		domain:           domain,
		messageSigner:    messageSigner,
		publisherID:      publisherID,
		registeredInputs: registeredInputs,
		senderTimestamp:  make(map[string]string),
		privateKey:       privateKey,
		subscriptions:    make(map[string]string),
		updateMutex:      &sync.Mutex{},
	}
	return recvsetin
}
