// Package messenger - Interface of messengers for publishers and subscribers
package messenger

import (
	"encoding/json"

	"github.com/hspaay/iotconnect.golang/messaging"
)

// MessengerConfig with configuration of a messenger
type MessengerConfig struct {
	ClientID string `yaml:"clientid,omitempty"` // optional connect ID, must be unique. Default is generated.
	Login    string `yaml:"login"`              // messenger login name
	Port     uint16 `yaml:"port,omitempty"`     // optional port, default is 8883 for TLS
	Password string `yaml:"credentials"`        // messenger login credentials
	PubQos   byte   `yaml:"pubqos,omitempty"`   // publishing QOS 0-2. Default=0
	Server   string `yaml:"server"`             // Message bus server/broker hostname or ip address, required
	SubQos   byte   `yaml:"subqos,omitempty"`   // Subscription QOS 0-2. Default=0
	Type     string `yaml:"type,omitempty"`     // Messenger client type: "DummyMessenger" (default) or "MQTTMessenger"
	Zone     string `yaml:"zone"`               // Zone in which this messenger publishes. Default is "local"
}

// IMessenger interface for messenger implementations
type IMessenger interface {

	// Connect the messenger.
	// This contains the last-will & testament information which is useful to inform subscribers
	//  when a publisher is unintentionally disconnected. Non MQTT busses can replace this with
	// their equivalent if available. Subscribers-only leave this empty.
	//
	// lastWillAddress optional last will & testament address for publishing device state
	//                 on accidental disconnect. Subscribers use "" to ignore.
	// lastWillValue payload to use with the last will publication
	Connect(lastWillAddress string, lastWillValue string) error

	// Gracefully disconnect the messenger and unsubscribe to all subscribed messages.
	// This will prevent the LWT publication so publishers must publish a graceful disconnect
	// message.
	Disconnect()

	// Return the zone the messenger publishes in
	GetZone() string

	// Sign and Publish a message
	// address to subscribe to as per IotConnect standard
	// publication object to transmit, this is an object that will be converted into a JSON
	Publish(address string, retained bool, publication *messaging.Publication) error

	// Publis raw data
	// address to subscribe top as per IotConnect standard
	// raw data, published as-is
	PublishRaw(address string, retained bool, raw json.RawMessage) error

	// Subscribe to a message
	// address to subscribe to with support for wildcards '+' and '#'. Non MQTT busses must conver to equivalent
	// onMessage callback is invoked when a message on this address is received
	Subscribe(address string, onMessage func(address string, publication *messaging.Publication))
}
