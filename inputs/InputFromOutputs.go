// Package inputs with using subscribed domain outputs as input
package inputs

import (
	"strings"
	"sync"

	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/sirupsen/logrus"
)

// InputFromOutputs subscribe to domain outputs to use as input
type InputFromOutputs struct {
	errorHandler     func(url string, err error) // invoke when error handling input message
	isRunning        bool                        // flag, subscriptions are active
	messageSigner    *messaging.MessageSigner    // subscription and publication messenger
	registeredInputs *RegisteredInputs           // registered inputs of this publisher
	senderTimestamp  map[string]string           // most recent timestamp of received commands by sender
	updateMutex      *sync.Mutex                 // mutex for async updating of inputs
}

// CreateInput adds a subscription to an output to use as input
// If the given output address is already subscribed to, its handler will be replaced
//  The given outputAddress is one of $raw or $latest output address
//  The handler is provided with the address, the sender and the received output value.
// This returns the new input address.
func (ifout *InputFromOutputs) CreateInput(
	nodeID string, inputType types.InputType, instance string,
	outputAddress string,
	handler func(address string, sender string, payload string)) string {

	input := ifout.registeredInputs.CreateInputWithSource(nodeID, inputType, instance, outputAddress, handler)

	ifout.updateMutex.Lock()
	defer ifout.updateMutex.Unlock()

	ifout.messageSigner.Subscribe(outputAddress, ifout.onReceiveOutput)
	return input.Address
}

// DeleteInput by address
func (ifout *InputFromOutputs) DeleteInput(nodeID string, inputType types.InputType, instance string) {
	ifout.updateMutex.Lock()
	defer ifout.updateMutex.Unlock()

	input := ifout.registeredInputs.GetInput(nodeID, inputType, instance)
	if input != nil {
		ifout.messageSigner.Unsubscribe(input.Source, ifout.onReceiveOutput)
		ifout.registeredInputs.DeleteInput(nodeID, inputType, instance)
	}
}

// onReceiveOutput verifies the message sender (for 'latest' outputs)
func (ifout *InputFromOutputs) onReceiveOutput(address string, message string) {
	var value string
	if strings.HasSuffix(address, types.MessageTypeRaw) {
		value = message
	} else if strings.HasSuffix(address, types.MessageTypeLatest) {
		latestMessage := types.OutputLatestMessage{}
		isSigned, err := ifout.messageSigner.VerifySignedMessage(message, &latestMessage)
		if err != nil {
			logrus.Warningf("onReceiveOutput: Sender of output on address %s failed to verify", address)
			return
		}
		// Verify this is the most recent message to protect against replay attacks
		prevTimestamp := ifout.senderTimestamp[address]
		if prevTimestamp > latestMessage.Timestamp {
			logrus.Warnf("onReceiveOutput: earlier timestamp of output %s. Message discarded.", address)
			return
		}
		ifout.senderTimestamp[address] = latestMessage.Timestamp
		_ = isSigned
		value = latestMessage.Value
	}

	// Find inputs that subscribe to this output
	inputs := ifout.registeredInputs.GetInputsWithSource(address)
	for _, input := range inputs {
		ifout.registeredInputs.NotifyInputHandler(input.Address, address, value)
	}
}

// NewInputFromOutputs creates a input list with subscriptions to outputs to use as input
func NewInputFromOutputs(
	messageSigner *messaging.MessageSigner,
	registeredInputs *RegisteredInputs,
) *InputFromOutputs {

	ifo := InputFromOutputs{
		messageSigner:    messageSigner,
		registeredInputs: registeredInputs,
		senderTimestamp:  make(map[string]string),
		updateMutex:      &sync.Mutex{}, // mutex for async updating of inputs
	}
	return &ifo
}
