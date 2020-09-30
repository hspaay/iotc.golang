// Package outputs with publication of output values
package outputs

import (
	"strings"
	"time"

	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/sirupsen/logrus"
)

// PublishOutputHistory publishes the $history output values retained=true
func PublishOutputHistory(
	output *types.OutputDiscoveryMessage,
	history OutputHistory,
	messageSigner *messaging.MessageSigner,
) {
	// output values are published using their alias address, if any
	addr := ReplaceMessageType(output.Address, types.MessageTypeHistory)
	timeStampStr := time.Now().Format("2006-01-02T15:04:05.000-0700")
	logrus.Infof("PublishOutputHistory to: %s", addr)

	// todo: use output configuration to determine if history is published for this output
	historyMessage := &types.OutputHistoryMessage{
		Address:   addr,
		Duration:  0, // tbd
		Timestamp: timeStampStr,
		Unit:      output.Unit,
		History:   history,
	}
	logrus.Debugf("PublishOutputHistory: %d entries to: %s", len(historyMessage.History), addr)
	messageSigner.PublishObject(addr, true, historyMessage, nil)
}

// PublishOutputLatest publishes the $latest output value
// not thread-safe, using within a locked section
func PublishOutputLatest(
	output *types.OutputDiscoveryMessage,
	latest *types.OutputValue,
	messageSigner *messaging.MessageSigner,
) {
	// output values are published using their alias address, if any
	addr := ReplaceMessageType(output.Address, types.MessageTypeLatest)
	logrus.Infof("PublishOutputLatest to: %s", addr)

	// todo: use output configuration to determine if latest message is published for this output
	// zone/publisher/node/iotype/instance/$latest
	latestMessage := &types.OutputLatestMessage{
		Address:   addr,
		Timestamp: latest.Timestamp,
		Unit:      output.Unit,
		Value:     latest.Value,
	}
	messageSigner.PublishObject(addr, true, latestMessage, nil)
}

// PublishOutputRaw publishes the raw output $raw (retained)
// not thread-safe, using within a locked section
func PublishOutputRaw(output *types.OutputDiscoveryMessage, value string, messageSigner *messaging.MessageSigner,
) error {

	// replace output discovery with raw message type: domain/pub/nodeId/type/instance/messagetype
	addr := ReplaceMessageType(output.Address, types.MessageTypeRaw)

	// publish raw value with the $raw command
	s := value
	// don't log full images
	if len(s) > 30 {
		s = s[:30]
	}
	logrus.Infof("PublishOutputRaw: output value '%s' to: %s", s, addr)

	err := messageSigner.PublishSigned(addr, true, value)
	return err
}

// ReplaceMessageType replace the last segment  with a new message type
func ReplaceMessageType(addr string, newMessageType types.MessageType) string {
	segments := strings.Split(addr, "/")
	segments[len(segments)-1] = string(newMessageType)
	newAddr := strings.Join(segments, "/")
	return newAddr
}
