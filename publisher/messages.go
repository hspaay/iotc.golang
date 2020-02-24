// Package publisher with message definitions for publishing and subscribing
package publisher

import (
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
