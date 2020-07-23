// Package messaging - Dummy in-memory messenger for testing
package messaging

import (
	"fmt"
	"strings"
	"sync"

	"github.com/iotdomain/iotdomain-go/types"
	"github.com/sirupsen/logrus"
)

// DummyMessenger that implements IMessenger
type DummyMessenger struct {
	publications  map[string]string
	config        *MessengerConfig // for domain configuration
	subscriptions []Subscription
	publishMutex  *sync.Mutex // mutex for concurrent publishing of messages
}

// Subscription to messages
type Subscription struct {
	address string
	handler func(address string, message string) error
}

// Connect the messenger
func (messenger *DummyMessenger) Connect(lastWillAddress string, lastWillValue string) error {
	return nil
}

// Disconnect gracefully disconnects the messenger
func (messenger *DummyMessenger) Disconnect() {
}

// FindLastPublication with the given address
func (messenger *DummyMessenger) FindLastPublication(addr string) (message string) {
	messenger.publishMutex.Lock()
	pub := messenger.publications[addr]
	messenger.publishMutex.Unlock()
	return pub
}

// GetDomain returns the domain in which this messenger operates
// This is provided via the messenger config file or defaults to types.LocalDomainID
func (messenger *DummyMessenger) GetDomain() string {
	domain := messenger.config.Domain
	if domain == "" {
		domain = types.LocalDomainID
	}
	return domain
}

// NrPublications returns the number of received publications
func (messenger *DummyMessenger) NrPublications() int {
	return len(messenger.publications)
}

// OnReceive function to simulate a received message
func (messenger *DummyMessenger) OnReceive(address string, message string) {
	messenger.publishMutex.Lock()
	subs := messenger.subscriptions
	messenger.publishMutex.Unlock()

	for _, subscription := range subs {
		match := messenger.matchAddress(address, subscription.address)

		if match && subscription.handler != nil {
			subscription.handler(address, message)
		}
	}
}

// Publish a message
// address is the MQTT address to send to
// retained (ignored)
// message JSON text or raw message base64 encoded text
func (messenger *DummyMessenger) Publish(address string, retained bool, message string) error {
	messenger.publishMutex.Lock()
	messenger.publications[address] = message
	messenger.publishMutex.Unlock()
	// go messenger.OnReceive(address, payload)
	messenger.OnReceive(address, message)
	return nil
}

// Subscribe to a message by address
func (messenger *DummyMessenger) Subscribe(
	address string, onMessage func(address string, message string) error) {

	logrus.Infof("DummyMessenger.Subscribe: address %s", address)
	subscription := Subscription{address: address, handler: onMessage}
	messenger.publishMutex.Lock()
	messenger.subscriptions = append(messenger.subscriptions, subscription)
	messenger.publishMutex.Unlock()
}

// Unsubscribe an address and handler
func (messenger *DummyMessenger) Unsubscribe(
	address string, onMessage func(address string, message string) error) {
	messenger.publishMutex.Lock()
	// https://stackoverflow.com/questions/9643205/how-do-i-compare-two-functions-for-pointer-equality-in-the-latest-go-weekly
	var onMessageID = onMessage
	onMessageStr := fmt.Sprintf("%v", &onMessageID)
	for i, sub := range messenger.subscriptions {
		// can't compare addresses directly so convert to string
		var handlerID = sub.handler
		handlerStr := fmt.Sprintf("%v", &handlerID)
		if sub.address == address && (onMessage == nil || handlerStr == onMessageStr) {
			// shift remainder left one index
			copy(messenger.subscriptions[i:], messenger.subscriptions[i+1:])
			messenger.subscriptions = messenger.subscriptions[:len(messenger.subscriptions)-1]
			if onMessage != nil {
				break
			}
		}
	}
	messenger.publishMutex.Unlock()
}

// test if a given address matches a subscription address with wildcards
func (messenger *DummyMessenger) matchAddress(address string, subscription string) (match bool) {
	subscriptionSegments := strings.Split(subscription, "/")
	addressSegments := strings.Split(address, "/")

	// no match subscription is longer than address
	if len(subscriptionSegments) > len(addressSegments) {
		return false
	}

	// Match the segments accepting wildcards. Rather crude but only intended for testing.
	match = true
	for index, addrSegment := range addressSegments {
		// if index >= len(subscriptionSegments) {
		// 	return false
		// }
		subscriptionSegment := subscriptionSegments[index]

		if subscriptionSegment == "#" {
			return true
		} else if subscriptionSegment == "+" {
			// match, continue
		} else if addrSegment == subscriptionSegment {
			// match continue
		} else {
			return false
		}
	}
	return match
}

// NewDummyMessenger provides a messenger for messages that go no.where...
func NewDummyMessenger(config *MessengerConfig) *DummyMessenger {
	var messenger = &DummyMessenger{
		config:        config,
		publications:  make(map[string]string, 0),
		subscriptions: make([]Subscription, 0),
		publishMutex:  &sync.Mutex{},
	}
	return messenger
}
