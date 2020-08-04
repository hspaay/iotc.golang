// Package inputs with using subscribed domain outputs as input
package inputs

import (
	"strings"
	"sync"

	"github.com/iotdomain/iotdomain-go/lib"
	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/iotdomain/iotdomain-go/types"
)

// ReceiveFromOutputs subscribe to domain outputs to use as input
type ReceiveFromOutputs struct {
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
func (ifout *ReceiveFromOutputs) CreateInput(
	nodeID string, inputType types.InputType, instance string,
	outputAddress string,
	handler func(input *types.InputDiscoveryMessage, sender string, payload string)) *types.InputDiscoveryMessage {

	input := ifout.registeredInputs.CreateInputWithSource(nodeID, inputType, instance, outputAddress, handler)

	ifout.updateMutex.Lock()
	defer ifout.updateMutex.Unlock()

	ifout.messageSigner.Subscribe(outputAddress, ifout.onReceiveOutput)
	return input
}

// DeleteInput by address
func (ifout *ReceiveFromOutputs) DeleteInput(inputID string) {
	ifout.updateMutex.Lock()
	defer ifout.updateMutex.Unlock()

	input := ifout.registeredInputs.GetInputByID(inputID)
	if input != nil {
		ifout.messageSigner.Unsubscribe(input.Source, ifout.onReceiveOutput)
		ifout.registeredInputs.DeleteInput(inputID)
	}
}

// onReceiveOutput verifies the message sender (for 'latest' outputs)
func (ifout *ReceiveFromOutputs) onReceiveOutput(address string, message string) error {
	var value string
	if strings.HasSuffix(address, types.MessageTypeRaw) {
		value = message
	} else if strings.HasSuffix(address, types.MessageTypeLatest) {
		latestMessage := types.OutputLatestMessage{}
		isSigned, err := ifout.messageSigner.VerifySignedMessage(message, &latestMessage)
		if err != nil {
			return lib.MakeErrorf("onReceiveOutput: Sender of output on address %s failed to verify: %s", address, err)
		}
		// Verify this is the most recent message to protect against replay attacks
		prevTimestamp := ifout.senderTimestamp[address]
		if prevTimestamp > latestMessage.Timestamp {
			return lib.MakeErrorf("onReceiveOutput: earlier timestamp of output %s. Message discarded.", address)
		}
		ifout.senderTimestamp[address] = latestMessage.Timestamp
		_ = isSigned
		value = latestMessage.Value
	}

	// Find inputs that subscribe to this output
	inputs := ifout.registeredInputs.GetInputsWithSource(address)
	for _, input := range inputs {
		ifout.registeredInputs.NotifyInputHandler(input.InputID, address, value)
	}
	return nil
}

// NewReceiveFromOutputs creates a input list with subscriptions to outputs to use as input
func NewReceiveFromOutputs(
	messageSigner *messaging.MessageSigner,
	registeredInputs *RegisteredInputs,
) *ReceiveFromOutputs {

	ifo := ReceiveFromOutputs{
		messageSigner:    messageSigner,
		registeredInputs: registeredInputs,
		senderTimestamp:  make(map[string]string),
		updateMutex:      &sync.Mutex{}, // mutex for async updating of inputs
	}
	return &ifo
}
