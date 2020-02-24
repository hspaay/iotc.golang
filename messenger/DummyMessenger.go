// Package messenger - Dummy in-memory messenger for testing
package messenger

import (
	"encoding/json"
	"strings"

	log "github.com/sirupsen/logrus"
)

// DummyMessenger that implements IMessenger
type DummyMessenger struct {
	Logger        *log.Logger
	Publications  map[string]*Publication
	subscriptions map[string]func(address string, publication *Publication)
}

// Payload for parsing raw message without parsing Message.
// At this stage we don't know the type yet
type Payload struct {
	Signature string          `json:"signature"`
	Message   json.RawMessage `json:"message"`
}

// Connect the messenger
func (messenger *DummyMessenger) Connect(lastWillAddress string, lastWillValue string) {
}

// Disconnect gracefully disconnects the messenger
func (messenger *DummyMessenger) Disconnect() {
}

// FindLastPublication with the given address
func (messenger *DummyMessenger) FindLastPublication(addr string) *Publication {
	pub := messenger.Publications[addr]
	return pub
}

// OnReceive function to simulate a received message
func (messenger *DummyMessenger) OnReceive(address string, rawPayload []byte) {
	messageParts := strings.Split(address, "/")
	var payload Payload
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
	publication.Message = string(payload.Message)

	for subscription, handler := range messenger.subscriptions {
		subscriptionSegments := strings.Split(subscription, "/")
		// Match the subscription. Rather crude but only intended for testing.
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
			handler(address, &publication)
		}
	}
}

// Publish a JSON encoded message
func (messenger *DummyMessenger) Publish(address string, publication *Publication) {
	messenger.Publications[address] = publication
}

// PublishRaw message
func (messenger *DummyMessenger) PublishRaw(address string, message string) {
	payload := Publication{
		Message: message,
	}
	messenger.Publications[address] = &payload
}

// Subscribe to a message by address
func (messenger *DummyMessenger) Subscribe(
	address string, onMessage func(address string, publication *Publication)) {

	messenger.subscriptions[address] = onMessage
}

// NewDummyMessenger provides a messenger for messages that go no.where...
// logger to use for debug messages
func NewDummyMessenger() *DummyMessenger {
	var logger = log.New()
	logger.SetReportCaller(true) // publisher logging includes caller and file:line#

	var messenger = &DummyMessenger{
		Logger:        logger,
		Publications:  make(map[string]*Publication, 0),
		subscriptions: make(map[string]func(addr string, publication *Publication)),
	}
	return messenger
}
