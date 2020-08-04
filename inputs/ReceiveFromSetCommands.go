// Package inputs with receiving of the SetInputMessage
package inputs

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/iotdomain/iotdomain-go/lib"
	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/sirupsen/logrus"
)

// ReceiveFromSetCommands handles set commands aimed at inputs managed by this publisher.
// This decrypts incoming messages determines the sender and verifies the signature with
// the sender public key. Last it translates from the publishing address to the input ID
// before passing the request to the handler associated with the input.
type ReceiveFromSetCommands struct {
	domain           string // the domain of this publisher
	publisherID      string // the registered publisher for the inputs
	isRunning        bool
	messageSigner    *messaging.MessageSigner // subscription and publication messenger
	senderTimestamp  map[string]string        // most recent timestamp of received commands by sender
	registeredInputs *RegisteredInputs        // registered inputs of this publisher
	// subscriptions of registered inputs
	subscriptions map[string]string // SetInput subscriptions of inputs [setAddr]setAddr
	updateMutex   *sync.Mutex       // mutex for async handling of inputs
}

// CreateInput creates a new input that responds to a set command from the message bus.
// If an input of the given deviceID, type and instance already exist it will be replaced.
// This returns the new input
func (ifset *ReceiveFromSetCommands) CreateInput(
	deviceID string, inputType types.InputType, instance string,
	handler func(input *types.InputDiscoveryMessage, sender string, value string)) *types.InputDiscoveryMessage {

	ifset.updateMutex.Lock()
	defer ifset.updateMutex.Unlock()

	input := ifset.registeredInputs.CreateInput(deviceID, inputType, instance, handler)
	// only subscribe if this is a new input
	ifset.subscribeToSetCommand(input)
	return input
}

// DeleteInput deletes the input and unsubscribes to the input's set command
func (ifset *ReceiveFromSetCommands) DeleteInput(inputID string) {
	ifset.updateMutex.Lock()
	defer ifset.updateMutex.Unlock()

	ifset.unsubscribeFromSetCommand(inputID)
	ifset.registeredInputs.DeleteInput(inputID)
}

// decodeSetCommand decrypts and verifies the signature of an incoming set command.
// If successful this passes the set command to the setInputHandler callback
func (ifset *ReceiveFromSetCommands) decodeSetCommand(address string, message string) error {
	var setMessage types.SetInputMessage

	// Check that address is one of our inputs
	segments := strings.Split(address, "/")
	// a full address is required
	if len(segments) < 6 {
		errText := fmt.Sprintf("decodeSetCommand: Destination address '%s' is incomplete.", address)
		return errors.New(errText)
	}
	// domain/pub/node/inputtype/instance/$input
	segments[5] = types.MessageTypeInputDiscovery
	inputAddr := strings.Join(segments, "/")

	isEncrypted, isSigned, err := ifset.messageSigner.DecodeMessage(message, &setMessage)

	if !isEncrypted {
		return lib.MakeErrorf("decodeSetCommand: Alias update of '%s' is not encrypted. Message discarded.", address)
	} else if !isSigned {
		return lib.MakeErrorf("decodeSetCommand: Alias update of '%s' is not signed. Message discarded.", address)
	} else if err != nil {
		return lib.MakeErrorf("decodeSetCommand: Message to %s. Error %s'. Message discarded.", address, err)
	}

	// Verify this is the most recent message to protect against replay attacks
	prevTimestamp := ifset.senderTimestamp[setMessage.Sender]
	if prevTimestamp > setMessage.Timestamp {
		errText := fmt.Sprintf("decodeSetCommand: earlier timestamp of message to input %s from sender %s."+
			" Message discarded.", address, setMessage.Sender)
		logrus.Warning(errText)
		return errors.New(errText)
	}
	ifset.senderTimestamp[setMessage.Sender] = setMessage.Timestamp
	logrus.Infof("decodeSetCommand successful for input %s. isEncrypted=%t, isSigned=%t",
		address, isEncrypted, isSigned)

	// the handler is responsible for authorization
	inputID := ifset.registeredInputs.addressMap[inputAddr]
	ifset.registeredInputs.NotifyInputHandler(inputID, setMessage.Sender, setMessage.Value)
	return nil
}

// subscribeToSetCommand to receive set input commands for the given node, type and instance
func (ifset *ReceiveFromSetCommands) subscribeToSetCommand(input *types.InputDiscoveryMessage) {
	// change message type $input to $set to make the set address from the input address
	segments := strings.Split(input.Address, "/")
	segments[5] = types.MessageTypeSet
	setAddr := strings.Join(segments, "/")

	// prevent double subscription
	_, hasSubscription := ifset.subscriptions[input.Address]
	if !hasSubscription {
		ifset.subscriptions[setAddr] = setAddr
		ifset.messageSigner.Subscribe(setAddr, ifset.decodeSetCommand)
	}
}

// unsubscribeFromSetCommand removes previous subscription
func (ifset *ReceiveFromSetCommands) unsubscribeFromSetCommand(inputID string) {
	// change message type $input to $set to make the set address from the input address
	input := ifset.registeredInputs.GetInputByID(inputID)
	segments := strings.Split(input.Address, "/")
	segments[5] = types.MessageTypeSet
	setAddr := strings.Join(segments, "/")

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

// NewReceiveFromSetCommands returns a new instance of handling of set input commands.
// The private key is used to decrypt set commands. Without it, decryption is disabled.
func NewReceiveFromSetCommands(
	domain string,
	publisherID string,
	messageSigner *messaging.MessageSigner,
	registeredInputs *RegisteredInputs) *ReceiveFromSetCommands {

	recvsetin := &ReceiveFromSetCommands{
		domain:           domain,
		messageSigner:    messageSigner,
		publisherID:      publisherID,
		registeredInputs: registeredInputs,
		senderTimestamp:  make(map[string]string),
		subscriptions:    make(map[string]string),
		updateMutex:      &sync.Mutex{},
	}
	return recvsetin
}
