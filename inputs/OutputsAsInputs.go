// Package inputs with using subscribed domain outputs as input
package inputs

import (
	"sync"

	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/sirupsen/logrus"
)

// OutputsAsInputs subscribe to domain outputs to use as input
type OutputsAsInputs struct {
	errorHandler  func(url string, err error)     // invoke when error handling input message
	isRunning     bool                            // flag, subscriptions are active
	logger        logrus.Logger                   //
	messenger     messaging.IMessenger            // subscription messenger
	subscriptions map[string]func(string, string) // list of subscriptions and handlers
	updateMutex   *sync.Mutex                     // mutex for async updating of inputs
}

// Start subscribing to inputs
func (sin *OutputsAsInputs) Start() {
	sin.updateMutex.Lock()
	defer sin.updateMutex.Unlock()
	sin.isRunning = true
	for address, handler := range sin.subscriptions {
		sin.messenger.Subscribe(address, handler)
	}
}

// Stop polling for inputs
func (sin *OutputsAsInputs) Stop() {
	sin.updateMutex.Lock()
	defer sin.updateMutex.Unlock()
	sin.isRunning = false
	// for address, handler := range sin.subscriptions {
	// 	sin.messenger.Unsubscribe(address, handler)
	// }
}

// AddInput adds a subscription to an output to use as input
// If the given output address is already subscribed to, its handler will be replaced
//  The given address is one of $raw or $latest output address
//  The handler is provided with the address and the payload. For $raw outputs the payload is base64encoded.
// for $latest outputs the payload is a json text with the message.
func (sin *OutputsAsInputs) AddInput(outputAddress string, handler func(address string, payload string)) {
	sin.updateMutex.Lock()
	defer sin.updateMutex.Unlock()

	_, exists := sin.subscriptions[outputAddress]
	if exists {
		// sin.messenger.Unsubscribe(address)
	}
	sin.subscriptions[outputAddress] = handler
	sin.messenger.Subscribe(outputAddress, handler)
}

// RemoveInput by address
func (sin *OutputsAsInputs) RemoveInput(address string, handler func(string, string)) {
	sin.updateMutex.Lock()
	defer sin.updateMutex.Unlock()
	sin.messenger.Unsubscribe(address, handler)
	delete(sin.subscriptions, address)
}
