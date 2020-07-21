package messaging_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var messengerConfig = messaging.MessengerConfig{
	Server: "localhost", // set this to your broker
	Port:   8883,
	// ClientID: "test1",
}

const pub1Addr = "domain1/pub1/test"

type PubMessage struct {
	Name string
}

var pub1Message = &PubMessage{Name: "bob"}
var pub1JSON, _ = json.Marshal(pub1Message)

func TestNewMessenger(t *testing.T) {
	config := messaging.MessengerConfig{Messenger: "MQTTMessenger"}
	m := messaging.NewMessenger(&config)
	assert.NotNil(t, m, "Failed creating mqtt messenger")

	config.Messenger = "dummy"
	m = messaging.NewMessenger(&config)
	assert.NotNil(t, m, "Failed creating dummy messenger")
}

// TestConnect to mqtt broker
func TestConnect(t *testing.T) {
	messenger := messaging.NewMqttMessenger(&messengerConfig)
	err := messenger.Connect("", "")
	assert.NoError(t, err, "Connection failed")
	// messenger.Disconnect()

	// using LWT
	err = messenger.Connect("test/pub1/$lwt", "last will and testament")
	assert.NoError(t, err, "Connection failed")
	messenger.Disconnect()

}

// TestPublish onto mqtt broker
func TestPublish(t *testing.T) {
	logrus.SetReportCaller(true) // publisher logging includes caller and file:line#

	messenger := messaging.NewMqttMessenger(&messengerConfig)

	err := messenger.Connect("", "")
	assert.NoError(t, err, "Connection failed")

	err = messenger.Publish(pub1Addr, false, string(pub1JSON))
	assert.NoError(t, err, "Publish failed")

	// raw
	err = messenger.PublishRaw(pub1Addr, false, string(pub1JSON))
	assert.NoError(t, err, "Publish failed")

	messenger.Disconnect()

	// test publish without connection
	err = messenger.Publish(pub1Addr, false, string(pub1JSON))
	assert.Error(t, err, "Publish without connection succeeded??")
	err = messenger.PublishRaw(pub1Addr, false, string(pub1JSON))
	assert.Error(t, err, "PublishRaw without connection succeeded??")
}

// TestPublishSubscribe onto mqtt broker
func TestPublishSubscribe(t *testing.T) {
	var receivedMessage PubMessage
	txLength := len(pub1JSON)
	rxLength := 0
	logrus.SetReportCaller(true) // publisher logging includes caller and file:line#

	messenger := messaging.NewMqttMessenger(&messengerConfig)
	messenger.Subscribe(pub1Addr, func(addr string, message string) error {
		return nil
	})

	err := messenger.Connect("", "")
	assert.NoError(t, err, "Connection failed")

	messenger.Subscribe(pub1Addr, func(addr string, message string) error {
		err := json.Unmarshal([]byte(message), &receivedMessage)
		assert.NoError(t, err, "Received message can't be parsed")
		rxLength = len(message)
		logrus.Infof("TestPublishSubscribe: Received message. Length=%d: %s", len(message), message)
		return nil
	})

	logrus.Infof("TestPublishSubscribe: sending message. Length=%d", len(pub1JSON))
	err = messenger.Publish(pub1Addr, false, string(pub1JSON))
	assert.NoError(t, err, "Publish failed")
	time.Sleep(time.Second * 2)
	messenger.Unsubscribe(pub1Addr, nil)
	messenger.Disconnect()

	assert.Equal(t, txLength, rxLength, "Sent and received message sizes don't match")

	assert.Equal(t, "bob", receivedMessage.Name, "Did not receive published message")
}
