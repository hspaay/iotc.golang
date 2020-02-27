// Package messenger - Publish and Subscribe to message using the MQTT message bus
package messenger

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const hostName = "pi3" // set this to your broker
const port = 1883
const login = ""
const password = ""
const clientID = "test1"

const pub1Addr = "zone1/pub1/test"

type PubMessage struct {
	name string
}

var pub1Message = PubMessage{name: "bob"}
var pub1Buffer, _ = json.Marshal(&pub1Message)
var pub1 = Publication{Message: pub1Buffer}

// TestConnect to mqtt broker
func TestConnect(t *testing.T) {
	logger := logrus.New()
	logger.SetReportCaller(true) // publisher logging includes caller and file:line#

	messenger := NewMqttMessenger(hostName, port, login, password, clientID, logger)
	err := messenger.Connect("", "")
	assert.NoError(t, err, "Connection failed")
	messenger.Disconnect()
}

// TestPublish onto mqtt broker
func TestPublish(t *testing.T) {
	logger := logrus.New()
	logger.SetReportCaller(true) // publisher logging includes caller and file:line#

	messenger := NewMqttMessenger(hostName, port, login, password, clientID, logger)
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
	logger.SetReportCaller(true) // publisher logging includes caller and file:line#

	messenger := NewMqttMessenger(hostName, port, login, password, clientID, logger)
	err := messenger.Connect("", "")
	assert.NoError(t, err, "Connection failed")

	err = messenger.Subscribe(pub1Addr, func(addr string, pub *Publication) {
		err := json.Unmarshal(pub.Message, &receivedMessage)
		assert.NoError(t, err, "Received message can't be parsed")

		logger.Infof("Received message yeaah")
	})

	err = messenger.Publish(pub1Addr, false, &pub1)
	assert.NoError(t, err, "Publish failed")
	time.Sleep(time.Second * 3)
	messenger.Disconnect()

	assert.Equal(t, "bob", receivedMessage.name, "Did not receive published message")
}
