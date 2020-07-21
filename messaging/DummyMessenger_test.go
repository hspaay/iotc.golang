package messaging_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var dummyConfig = messaging.MessengerConfig{
	Server: "localhost", // set this to your broker
	Port:   8883,
	// ClientID: "test1",
}

const dummy1Addr = "domain1/pub1/test"

type DummyMessage struct {
	Name string
}

// TestConnect to dummy
func TestDummyConnect(t *testing.T) {
	messenger := messaging.NewDummyMessenger(&dummyConfig)
	err := messenger.Connect("", "")

	domain := messenger.GetDomain()
	assert.Equal(t, types.LocalDomainID, domain)

	assert.NoError(t, err, "Connection failed")
	messenger.Disconnect()
}

// TestPublish a message
func TestDummyPublish(t *testing.T) {
	var pub1Message = &DummyMessage{Name: "bob"}
	var pub1JSON, _ = json.Marshal(pub1Message)

	messenger := messaging.NewDummyMessenger(&dummyConfig)
	err := messenger.Connect("", "")
	assert.NoError(t, err, "Connection failed")

	err = messenger.Publish(dummy1Addr, false, string(pub1JSON))
	assert.NoError(t, err, "Publish failed")
	messenger.Disconnect()
}

// TestPublishSubscribe onto mqtt broker
func TestDummyPublishSubscribe(t *testing.T) {
	var pub1Message = &DummyMessage{Name: "bob"}
	var pub1JSON, _ = json.Marshal(pub1Message)
	const dummy2Addr = "+/pub1/#"

	var receivedMessage DummyMessage
	txLength := len(pub1JSON)
	rxLength := 0

	messenger := messaging.NewDummyMessenger(&dummyConfig)
	err := messenger.Connect("", "")
	assert.NoError(t, err, "Connection failed")

	messenger.Subscribe(dummy2Addr, func(addr string, message string) error {
		err := json.Unmarshal([]byte(message), &receivedMessage)
		assert.NoError(t, err, "Received message can't be parsed")
		rxLength = len(message)
		logrus.Infof("TestDummyPublishSubscribe: Received message. Length=%d: %s", len(message), message)
		return nil
	})

	logrus.Infof("TestDummyPublishSubscribe: sending message. Length=%d", len(pub1JSON))
	err = messenger.Publish(dummy1Addr, false, string(pub1JSON))
	assert.NoError(t, err, "Publish failed")
	time.Sleep(time.Second * 2)

	lp := messenger.FindLastPublication(dummy1Addr)
	assert.NotEmpty(t, lp, "Cant find last pub")
	nrPublications := messenger.NrPublications()
	assert.Equal(t, 1, nrPublications, "Expected 1 publication")

	messenger.Unsubscribe(dummy2Addr, nil)
	messenger.Disconnect()

	assert.Equal(t, txLength, rxLength, "Sent and received message sizes don't match")

	assert.Equal(t, "bob", receivedMessage.Name, "Did not receive published message")
}
