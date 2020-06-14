// Package publisher with handling of input commands
package publisher

import (
	"strings"
	"time"

	"github.com/hspaay/iotc.golang/iotc"
)

// PublishSetInput sets the input of a remote node by this publisher
func (publisher *Publisher) PublishSetInput(remoteNodeInputAddress string, value string) {
	// Check that address is one of our inputs
	segments := strings.Split(remoteNodeInputAddress, "/")
	// a full address is required
	if len(segments) < 6 {
		return
	}
	// zone/pub/node/inputtype/instance/$set
	segments[5] = iotc.MessageTypeSet
	inputAddr := strings.Join(segments, "/")

	// Encecode the SetMessage
	timeStampStr := time.Now().Format("2006-01-02T15:04:05.000-0700")
	var setMessage = iotc.SetInputMessage{
		Address:   inputAddr,
		Sender:    publisher.Address(),
		Timestamp: timeStampStr,
		Value:     value,
	}
	publisher.publishObject(inputAddr, false, &setMessage)
}
