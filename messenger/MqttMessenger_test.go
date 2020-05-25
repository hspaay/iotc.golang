// Package messenger - Publish and Subscribe to message using the MQTT message bus
// This requires a running MQTT server on localhost
package messenger

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/hspaay/iotc.golang/iotc"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var messengerConfig = MessengerConfig{
	Server: "localhost", // set this to your broker
	Port:   8883,
	// ClientID: "test1",
}

const pub1Addr = "zone1/pub1/test"

type PubMessage struct {
	Name string
}

var pub1Message = &PubMessage{Name: "bob"}
var pub1Buffer, _ = json.Marshal(pub1Message)
var pub1 = iotc.Publication{Message: string(pub1Buffer)}

// TestConnect to mqtt broker
func TestConnect(t *testing.T) {
	logger := logrus.New()
	logger.SetReportCaller(true) // publisher logging includes caller and file:line#

	messenger := NewMqttMessenger(&messengerConfig, logger)
	err := messenger.Connect("", "")
	assert.NoError(t, err, "Connection failed")
	messenger.Disconnect()
}

// TestPublish onto mqtt broker
func TestPublish(t *testing.T) {
	logger := logrus.New()
	logger.SetReportCaller(true) // publisher logging includes caller and file:line#

	messenger := NewMqttMessenger(&messengerConfig, logger)
	err := messenger.Connect("", "")
	assert.NoError(t, err, "Connection failed")

	err = messenger.Publish(pub1Addr, false, &pub1)
	assert.NoError(t, err, "Publish failed")
	messenger.Disconnect()
}

// TestPublishSubscribe onto mqtt broker
func TestPublishSubscribe(t *testing.T) {
	var receivedMessage PubMessage
	logger := logrus.New()
	txLength := len(pub1.Message)
	rxLength := 0
	logger.SetReportCaller(true) // publisher logging includes caller and file:line#

	messenger := NewMqttMessenger(&messengerConfig, logger)
	err := messenger.Connect("", "")
	assert.NoError(t, err, "Connection failed")

	messenger.Subscribe(pub1Addr, func(addr string, pub *iotc.Publication) {
		err := json.Unmarshal([]byte(pub.Message), &receivedMessage)
		assert.NoError(t, err, "Received message can't be parsed")
		rxLength = len(pub.Message)
		logger.Infof("TestPublishSubscribe: Received message. Length=%d: %s", len(pub.Message), pub.Message)
	})

	logger.Infof("TestPublishSubscribe: sending message. Length=%d", len(pub1.Message))
	err = messenger.Publish(pub1Addr, false, &pub1)
	assert.NoError(t, err, "Publish failed")
	time.Sleep(time.Second * 2)
	messenger.Disconnect()

	assert.Equal(t, txLength, rxLength, "Sent and received message sizes don't match")

	assert.Equal(t, "bob", receivedMessage.Name, "Did not receive published message")
}
