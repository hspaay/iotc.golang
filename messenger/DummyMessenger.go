// Package messenger - Dummy in-memory messenger for testing
package messenger

import (
	"encoding/json"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
)

// DummyMessenger that implements IMessenger
type DummyMessenger struct {
	Logger        *log.Logger
	Publications  map[string]*Publication
	subscriptions []Subscription
	publishMutex  *sync.Mutex // mutex for concurrent publishing of messages
}

// Subscription to messages
type Subscription struct {
	address string
	handler func(address string, publication *Publication)
}

// Connect the messenger
func (messenger *DummyMessenger) Connect(lastWillAddress string, lastWillValue string) error {
	return nil
}

// Disconnect gracefully disconnects the messenger
func (messenger *DummyMessenger) Disconnect() {
}

// FindLastPublication with the given address
func (messenger *DummyMessenger) FindLastPublication(addr string) *Publication {
	messenger.publishMutex.Lock()
	pub := messenger.Publications[addr]
	messenger.publishMutex.Unlock()
	return pub
}

// OnReceive function to simulate a received message
func (messenger *DummyMessenger) OnReceive(address string, rawPayload []byte) {
	messageParts := strings.Split(address, "/")
	var payload Publication
	var publication Publication
	var rawStr = string(rawPayload)
	_ = rawStr
	err := json.Unmarshal(rawPayload, &payload)
	// messageStr := string(publication.Message)
	if err != nil {
		messenger.Logger.Infof("Unable to unmarshal payload on address %s. Error: %s", address, err)
		return
	}
	publication.Signature = payload.Signature
	publication.Message = payload.Message

	for _, subscription := range messenger.subscriptions {
		subscriptionSegments := strings.Split(subscription.address, "/")
		// Match the address accepting wildcards. Rather crude but only intended for testing.
		match := true
		for index, addrSegment := range messageParts {
			if index >= len(subscriptionSegments) {
				match = false
				break // no match, message address is longer
			}
			subscriptionSegment := subscriptionSegments[index]
			if subscriptionSegment == "#" {
				match = true
				break
			} else if subscriptionSegment == "+" {
				// match, continue
			} else if addrSegment == subscriptionSegment {
				// match continue
			} else {
				match = false
				break // no match
			}
		}
		if match {
			subscription.handler(address, &publication)
		}
	}
}

// Publish a JSON encoded message
func (messenger *DummyMessenger) Publish(address string, retained bool, publication *Publication) error {
	messenger.publishMutex.Lock()
	messenger.Publications[address] = publication
	messenger.publishMutex.Unlock()
	//
	payload, err := json.Marshal(publication)
	if err != nil {
		messenger.Logger.Errorf("Failed marshalling publication for address %s", address)
		return err
	}
	go messenger.OnReceive(address, payload)
	return nil
}

// PublishRaw message
func (messenger *DummyMessenger) PublishRaw(address string, retained bool, message json.RawMessage) error {
	payload := Publication{
		Message: message,
	}
	messenger.publishMutex.Lock()
	messenger.Publications[address] = &payload
	messenger.publishMutex.Unlock()
	return nil
}

// Subscribe to a message by address
func (messenger *DummyMessenger) Subscribe(
	address string, onMessage func(address string, publication *Publication)) {

	subscription := Subscription{address: address, handler: onMessage}
	messenger.subscriptions = append(messenger.subscriptions, subscription)
}

// NewDummyMessenger provides a messenger for messages that go no.where...
// logger to use for debug messages
func NewDummyMessenger() *DummyMessenger {
	var logger = log.New()
	logger.SetReportCaller(true) // publisher logging includes caller and file:line#

	var messenger = &DummyMessenger{
		Logger:        logger,
		Publications:  make(map[string]*Publication, 0),
		subscriptions: make([]Subscription, 0),
		publishMutex:  &sync.Mutex{},
	}
	return messenger
}
