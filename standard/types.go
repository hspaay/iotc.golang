// Package standard with type definitions from the iotconnect standard
package standard

import (
	"fmt"
	"strings"
	"time"
)

// Node definition
type Node struct {
	ID                string        `yaml:"id"                 json:"id"`
	Address           string        `yaml:"address"            json:"address"`                              // Node discovery address
	Attr              AttrMap       `yaml:"attr,omitempty"     json:"attr,omitempty"`                       // Node/service specific info attributes
	Config            ConfigAttrMap `yaml:"config,omitempty"   json:"config,omitempty"`                     // Node/service configuration.
	Identity          *Identity     `yaml:"identity,omitempty" json:"identity,omitempty"`                   // Identity if node is a publisher
	IdentitySignature string        `yaml:"identitySignature,omitempty" json:"identitySignature,omitempty"` // optional signature of the identity by the ZSAS
	Publisher         string        `yaml:"publisher"          json:"publisher"`                            // publisher ID
	Status            AttrMap       `yaml:"status,omitempty"   json:"status,omitempty"`                     // include status at time of discovery
	// Inputs      map[string]*InOutput `yaml:"inputs,omitempty"  json:"inputs,omitempty"`  // inputs by their type.instance
	// Outputs     map[string]*InOutput `yaml:"outputs,omitempty" json:"outputs,omitempty"` // outputs by their type.instance
	// historySize int                  `yaml:"-" json:"-"`                                 // size of history for inputs and outputs
	// repeatDelay int                  `yaml:"-" json:"-"`
	// modified    bool                 `yaml:"-" json:"-"` // Node/service attribute or config has been updated
}

// Identity contains the identity information of a publisher node
type Identity struct {
	Address          string `yaml:"address"   json:"address"`                   // discovery address of the publisher (zone/pub/\$publisher/\$node)
	Expires          string `yaml:"expires"   json:"expires"`                   // timestamp this identity expires
	Location         string `yaml:"location"  json:"location"`                  // city, province, country
	Organization     string `yaml:"organization"    json:"organization"`        // publishing organization
	PublicKeyCrypto  string `yaml:"publicKeyCrypto"    json:"publicKeyCrypto"`  // public key for encrypting messages to this publisher
	PublicKeySigning string `yaml:"publicKeySigning"   json:"publicKeySigning"` // public key for verifying signature of messages published by this publisher
	Publisher        string `yaml:"publisher"    json:"publisher"`              // publisher ID
	Timestamp        string `yaml:"timestamp"    json:"timestamp"`              // timestamp this identity was last renewed/verified
	URL              string `yaml:"url"      json:"url"`                        // Web URL related to the publisher identity, if applicable
	Zone             string `yaml:"zone"     json:"zone"`                       // Zone in which publisher lives
}

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

// AttrMap for use in node attributes and node status attributes
type AttrMap map[string]string

// ConfigAttrMap for use in node and in/output configuration
type ConfigAttrMap map[string]*ConfigAttr

// ConfigAttr describing the configuration of the device/service or sensor
type ConfigAttr struct {
	DataType    DataType `yaml:"datatype,omitempty" json:"datatype,omitempty"`       // Data type of the attribute. [integer, float, boolean, string, bytes, enum, ...]
	Default     string   `yaml:"default,omitempty" json:"default,omitempty"`         // Default value
	Description string   `yaml:"description,omitempty" json:"description,omitempty"` // Description of the attribute
	Enum        []string `yaml:"enum,omitempty" json:"enum,omitempty"`               // Possible valid enum values
	Max         float64  `yaml:"max,omitempty" json:"max,omitempty"`                 // Max value for numbers
	Min         float64  `yaml:"min,omitempty" json:"min,omitempty"`                 // Min value for numbers
	//Name        string   `yaml:"name,omitempty" json:"name,omitempty"`               // Friendly name
	Secret bool   `yaml:"secret,omitempty" json:"secret,omitempty"` // the attribute value was set encrypted. Don't publish the value.
	Value  string `yaml:"value,omitempty" json:"value,omitempty"`   // Current value of the attribute. Could be a string or map
	//ValueList   []interface{} `yaml:"valuelist,omitempty" json:"valuelist,omitempty"`             // Current value of the attribute. Could be a string or map
	//ValueList   []interface{} `yaml:"-" json:"-"`             // Current value of the attribute. Could be a string or map
}

// ConfigureMessage with configuration parameters
type ConfigureMessage struct {
	Address   string  `json:"address"` // zone/publisher/node/$configure
	Config    AttrMap `json:"config"`
	Sender    string  `json:"sender"`
	Timestamp string  `json:"timestamp"`
}

// EventMessage message with multiple output values
type EventMessage struct {
	Address   string            `json:"address"`
	Event     map[string]string `json:"event"`
	Sender    string            `json:"sender"`
	Timestamp string            `json:"timestamp"`
}

// UpgradeMessage with node firmware
type UpgradeMessage struct {
	Address   string `json:"address"`
	MD5       string `json:"md5"`
	Firmware  []byte `json:"firmware"`
	FWVersion string `json:"fwversion"`
	Sender    string `json:"sender"`
	Timestamp string `json:"timestamp"`
}

// HistoryMessage with the '$latest' output value and metadata
type HistoryMessage struct {
	Address   string         `json:"address"`
	Duration  int            `json:"duration,omitempty"`
	History   []HistoryValue `json:"history"`
	Sender    string         `json:"sender"`
	Timestamp string         `json:"timestamp"`
	Unit      Unit           `json:"unit,omitempty"`
}

// HistoryValue of node output
type HistoryValue struct {
	time      time.Time
	TimeStamp string `yaml:"timestamp"`
	Value     string `yaml:"value"`
}

// LatestMessage struct to send/receive the '$latest' command
type LatestMessage struct {
	Address   string `json:"address"`
	Sender    string `json:"sender"`
	Timestamp string `json:"timestamp"` // timestamp of value
	Unit      Unit   `json:"unit,omitempty"`
	Value     string `json:"value"`
}

// SetMessage to set node input
type SetMessage struct {
	Address   string `json:"address"` // zone/publisher/node/$set/type/instance
	Sender    string `json:"sender"`
	Timestamp string `json:"timestamp"`
	Value     string `json:"value"`
}

// GetHistory returns the last updated time and value
func GetHistory(inout *InOutput) []HistoryValue {
	return inout.history
}

// GetOutputValue returns the current output value
func GetOutputValue(output *InOutput) string {
	return output.history[0].Value
}

// UpdateValue inserts a new value at the front of the output history
// If the value hasn't change it is ignored unless the the previous value is older than the repeatDelay
// This function is not thread-safe and should only be used from within a locked section
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
	address := fmt.Sprintf("%s/%s/%s/"+CommandInputDiscovery+"/%s/%s", segments[0], segments[1], segments[2], ioType, instance)
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
	address := fmt.Sprintf("%s/%s/%s/"+CommandOutputDiscovery+"/%s/%s", segments[0], segments[1], segments[2], ioType, instance)
	io := &InOutput{
		Address:  address,
		Config:   ConfigAttrMap{},
		Instance: instance,
		IOType:   ioType,
		history:  make([]HistoryValue, 1),
	}
	return io
}

// NewNode instance
func NewNode(zoneID string, publisherID string, nodeID string) *Node {
	address := fmt.Sprintf("%s/%s/%s/"+CommandNodeDiscovery, zoneID, publisherID, nodeID)
	return &Node{
		Address:   address,
		Attr:      AttrMap{},
		Config:    ConfigAttrMap{},
		ID:        nodeID,
		Publisher: publisherID,
		Status:    AttrMap{},
	}
}
