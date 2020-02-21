// Package nodes with inputs and output definitions
package nodes

import (
	"fmt"
	"strings"
	"time"
)

// InOutput describing an input or output
type InOutput struct {
	Address     string        `yaml:"address"                 json:"address"`               // I/O address
	Config      ConfigAttrMap `yaml:"config,omitempty"        json:"config,omitempty"`      // Configuration of input or output
	DataType    DataType      `yaml:"datatype,omitempty"      json:"datatype,omitempty"`    //
	Description string        `yaml:"description,omitempty"   json:"description,omitempty"` // optional description
	EnumValues  []string      `yaml:"enum,omitempty"          json:"enum,omitempty"`        // enum valid values
	Instance    string        `yaml:"instance,omitempty"      json:"instance,omitempty"`    // instance identifier for multi-I/O nodes
	IOType      string        `yaml:"type,omitempty"          json:"type,omitempty"`        // type of input or output as per IOTypeXyz
	NodeID      string        `yaml:"nodeID"                  json:"nodeID"`                // The node ID
	Unit        Unit          `yaml:"unit,omitempty"          json:"unit,omitempty"`        // unit of value

	history     []HistoryValue // change history
	repeatDelay int            // debounce interval in seconds before repeating updates with the same value
}

// HistoryValue of node output
type HistoryValue struct {
	time      time.Time
	TimeStamp string `yaml:"timestamp"`
	Value     string `yaml:"value"`
}

// GetLatest returns the last updated time and value
func GetLatest(inout *InOutput) HistoryValue {
	return inout.history[0]
}

// UpdateValue inserts a new value at the front of the output history
// If the value hasn't change it is ignored unless the the previous value is older than the repeatDelay
func UpdateValue(inout *InOutput, newValue string) {
	timeStamp := time.Now()
	previous := inout.history[0]
	age := time.Now().Sub(previous.time)
	ageSeconds := int(age.Seconds())
	updated := (newValue != previous.Value) || (ageSeconds > inout.repeatDelay)
	if updated {
		timeStampStr := timeStamp.Format("2006-01-02T15:04:05.000-0700")

		latest := HistoryValue{
			time:      timeStamp,
			TimeStamp: timeStampStr,
			Value:     newValue,
		}
		// TODO: max history depth = 24 hours or limit count
		copy(inout.history[1:], inout.history[0:])
		inout.history[0] = latest
	}
}

// NewInput instance
func NewInput(node *Node, ioType string, instance string) *InOutput {
	segments := strings.Split(node.Address, "/")
	address := fmt.Sprintf("%s/%s/%s/$input/%s/%s", segments[0], segments[1], segments[2], ioType, instance)
	io := &InOutput{
		Address:  address,
		Config:   ConfigAttrMap{},
		Instance: instance,
		IOType:   ioType,
		history:  make([]HistoryValue, 1),
	}
	return io
}

// NewOutput instance
func NewOutput(node *Node, ioType string, instance string) *InOutput {
	segments := strings.Split(node.Address, "/")
	address := fmt.Sprintf("%s/%s/%s/$output/%s/%s", segments[0], segments[1], segments[2], ioType, instance)
	io := &InOutput{
		Address:  address,
		Config:   ConfigAttrMap{},
		Instance: instance,
		IOType:   ioType,
		history:  make([]HistoryValue, 1),
	}
	return io
}
