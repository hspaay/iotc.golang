// Package publisher with message definitions for publishing and subscribing
package publisher

import (
	"encoding/json"
	"iotzone/messenger"
	"iotzone/nodes"
)

// ConfigureMessage with configuration parameters
type ConfigureMessage struct {
	Address   string        `json:"address"` // zone/publisher/node/$configure
	Config    nodes.AttrMap `json:"config"`
	Sender    string        `json:"sender"`
	Timestamp string        `json:"timestamp"`
}

// EventMessage message with multiple output values
type EventMessage struct {
	Address   string            `json:"address"`
	Event     map[string]string `json:"event"`
	Sender    string            `json:"sender"`
	Timestamp string            `json:"timestamp"`
}

// HistoryMessage struct to send/receive the '$latest' command
type HistoryMessage struct {
	Address   string               `json:"address"`
	Duration  int                  `json:"duration,omitempty"`
	History   []nodes.HistoryValue `json:"history"`
	Sender    string               `json:"sender"`
	Timestamp string               `json:"timestamp"`
	Unit      nodes.Unit           `json:"unit,omitempty"`
}

// LatestMessage struct to send/receive the '$latest' command
type LatestMessage struct {
	Address   string     `json:"address"`
	Sender    string     `json:"sender"`
	Timestamp string     `json:"timestamp"` // timestamp of value
	Unit      nodes.Unit `json:"unit,omitempty"`
	Value     string     `json:"value"`
}

// SetMessage to set node input
type SetMessage struct {
	Address   string `json:"address"` // zone/publisher/node/$set/type/instance
	Sender    string `json:"sender"`
	Timestamp string `json:"timestamp"`
	Value     string `json:"value"`
}

// handle an incoming a set command for one of our nodes. This:
// - check if the signature is valid
// - check if the node is valid
// - pass the input value update to the adapter's callback method set in Start()
//
func (publisher *ThisPublisherState) handleNodeInput(address string, publication *messenger.Publication) {
	// TODO: authorization check
	input := publisher.GetInput(address)
	if input == nil || publication.Message == "" {
		publisher.Logger.Infof("handleNodeInput unknown input for address %s or missing message", address)
		return
	}
	var setMessage SetMessage
	err := json.Unmarshal([]byte(publication.Message), &setMessage)
	if err != nil {
		publisher.Logger.Infof("Unable to unmarshal SetMessage in %s", address)
		return
	}
	if publisher.onSetMessage != nil {
		publisher.onSetMessage(input, &setMessage)
	}
}
