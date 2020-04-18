// Package standard with type definitions from the iotconnect standard
package standard

import (
	"fmt"
	"time"
)

// // Node definition.
// type Node struct {
// 	ID                string        `json:"id"`
// 	Address           string        `json:"address"`                     // Node discovery address
// 	Attr              AttrMap       `json:"attr,omitempty"`              // Node/service specific info attributes
// 	Config            ConfigAttrMap `json:"config,omitempty"`            // Node/service configuration.
// 	Identity          *Identity     `json:"identity,omitempty"`          // Identity if node is a publisher
// 	IdentitySignature string        `json:"identitySignature,omitempty"` // optional signature of the identity by the ZSAS
// 	PublisherID       string        `json:"publisher"`                   // publisher ID
// 	Status            AttrMap       `json:"status,omitempty"`            // include status at time of discovery
// 	Zone              string        `json:"zone"`                        // Zone in which node lives
// 	HistorySize       int           `json:"historySize"`                 // size of history for inputs and outputs, default automatically for 24 hours
// 	RepeatDelay       int           `json:"repeatDelay"`                 // delay in seconds before repeating the same value, default 1 hour
// }

// // Identity contains the identity information of a publisher node
// type Identity struct {
// 	Address          string `json:"address"`          // discovery address of the publisher (zone/pub/\$publisher/\$node)
// 	Expires          string `json:"expires"`          // timestamp this identity expires
// 	Location         string `json:"location"`         // city, province, country
// 	Organization     string `json:"organization"`     // publishing organization
// 	PublicKeyCrypto  string `json:"publicKeyCrypto"`  // public key for encrypting messages to this publisher
// 	PublicKeySigning string `json:"publicKeySigning"` // public key for verifying signature of messages published by this publisher
// 	Publisher        string `json:"publisher"`        // publisher ID
// 	Timestamp        string `json:"timestamp"`        // timestamp this identity was last renewed/verified
// 	URL              string `json:"url"`              // Web URL related to the publisher identity, if applicable
// 	Zone             string `json:"zone"`             // Zone in which publisher lives
// }

// // InOutput describing an input or output
// type InOutput struct {
// 	Address     string        `json:"address"`               // I/O address
// 	Config      ConfigAttrMap `json:"config,omitempty"`      // Configuration of input or output
// 	DataType    DataType      `json:"datatype,omitempty"`    //
// 	Description string        `json:"description,omitempty"` // optional description
// 	EnumValues  []string      `json:"enum,omitempty"`        // enum valid values
// 	Instance    string        `json:"instance,omitempty"`    // instance identifier for multi-I/O nodes
// 	IOType      string        `json:"type,omitempty"`        // type of input or output as per IOTypeXyz
// 	NodeID      string        `json:"nodeID"`                // The node ID
// 	Unit        Unit          `json:"unit,omitempty"`        // unit of value
// }

// // AttrMap for use in node attributes and node status attributes
// type AttrMap map[string]string

// // ConfigAttrMap for use in node and in/output configuration
// type ConfigAttrMap map[string]ConfigAttr

// // ConfigAttr describing the configuration of the device/service or sensor
// type ConfigAttr struct {
// 	DataType    DataType `json:"datatype,omitempty"`    // Data type of the attribute. [integer, float, boolean, string, bytes, enum, ...]
// 	Default     string   `json:"default,omitempty"`     // Default value
// 	Description string   `json:"description,omitempty"` // Description of the attribute
// 	Enum        []string `json:"enum,omitempty"`        // Possible valid enum values
// 	ID          string   `json:"id,omitempty"`          // Unique ID of this config
// 	Max         float64  `json:"max,omitempty"`         // Max value for numbers
// 	Min         float64  `json:"min,omitempty"`         // Min value for numbers
// 	Secret      bool     `json:"secret,omitempty"`      // the attribute value was set encrypted. Don't publish the value.
// 	Value       string   `json:"value,omitempty"`       // Current value of the attribute. Could be a string or map
// }

// ConfigureMessage with configuration parameters
type ConfigureMessage struct {
	Address   string            `json:"address"` // zone/publisher/node/$configure
	Config    map[string]string `json:"config"`
	Sender    string            `json:"sender"`
	Timestamp string            `json:"timestamp"`
}

// EventMessage message with multiple output values
type EventMessage struct {
	Address   string            `json:"address"`
	Event     map[string]string `json:"event"`
	Sender    string            `json:"sender"`
	Timestamp string            `json:"timestamp"`
}

// ForecastMessage with the '$forecast' output value and metadata
type ForecastMessage struct {
	Address   string      `json:"address"`
	Duration  int         `json:"duration,omitempty"`
	Forecast  HistoryList `json:"forecast"`
	Sender    string      `json:"sender"`
	Timestamp string      `json:"timestamp"`
	Unit      string      `json:"unit,omitempty"`
}

// HistoryMessage with the '$history' output value and metadata
type HistoryMessage struct {
	Address   string      `json:"address"`
	Duration  int         `json:"duration,omitempty"`
	History   HistoryList `json:"history"`
	Sender    string      `json:"sender"`
	Timestamp string      `json:"timestamp"`
	Unit      string      `json:"unit,omitempty"`
}

// HistoryList is a list of history values in most recent first order
type HistoryList []*HistoryValue

// HistoryValue of node output
type HistoryValue struct {
	Timestamp time.Time `json:"timestamp"`
	// Timestamp string `json:"timestamp"`
	Value string `json:"value"` // this can also be a string containing a list, eg "[ a, b, c ]""
}

// LatestMessage struct to send/receive the '$latest' command
type LatestMessage struct {
	Address   string `json:"address"`
	Sender    string `json:"sender"`
	Timestamp string `json:"timestamp"` // timestamp of value
	Unit      string `json:"unit,omitempty"`
	Value     string `json:"value"` // this can also be a string containing a list, eg "[ a, b, c ]""
}

// SetMessage to set node input
type SetMessage struct {
	Address   string `json:"address"` // zone/publisher/node/$set/type/instance
	Sender    string `json:"sender"`
	Timestamp string `json:"timestamp"`
	Value     string `json:"value"` // this can also be a string containing a list, eg "[ a, b, c ]""
}

// UpgradeMessage with node firmware
type UpgradeMessage struct {
	Address   string `json:"address"`   // message address
	MD5       string `json:"md5"`       // firmware MD5
	Firmware  []byte `json:"firmware"`  // firmware code
	FWVersion string `json:"fwversion"` // firmware version
	Sender    string `json:"sender"`
	Timestamp string `json:"timestamp"`
}

// // NewOutput instance
// func NewOutput(node *Node, ioType string, instance string) *InOutput {
// 	address := MakeOutputDiscoveryAddress(node.Zone, node.PublisherID, node.ID, ioType, instance)
// 	io := &InOutput{
// 		Address:  address,
// 		Config:   ConfigAttrMap{},
// 		Instance: instance,
// 		IOType:   ioType,
// 		// History:  make([]*HistoryValue, 1),
// 	}
// 	return io
// }

// // NewNode instance
// func NewNode(zoneID string, publisherID string, nodeID string) *Node {
// 	address := MakeNodeDiscoveryAddress(zoneID, publisherID, nodeID)
// 	return &Node{
// 		Address:     address,
// 		Attr:        AttrMap{},
// 		Config:      ConfigAttrMap{},
// 		ID:          nodeID,
// 		PublisherID: publisherID,
// 		Status:      AttrMap{},
// 		Zone:        zoneID,
// 	}
// }

// NewConfigAttr instance for holding node configuration
// func NewConfigAttr(id string, dataType DataType, description string, value string) *ConfigAttr {
// 	config := ConfigAttr{
// 		ID:          id,
// 		DataType:    dataType,
// 		Description: description,
// 		Value:       value,
// 	}
// 	return &config
// }

// // MakeOutputDiscoveryAddress for publishing or subscribing
// func MakeOutputDiscoveryAddress(zone string, publisherID string, nodeID string, ioType string, instance string) string {
// 	address := fmt.Sprintf("%s/%s/%s/"+CommandOutputDiscovery+"/%s/%s",
// 		zone, publisherID, nodeID, ioType, instance)
// 	return address
// }

// // MakeInputDiscoveryAddress for publishing or subscribing
// func MakeInputDiscoveryAddress(zone string, publisherID string, nodeID string, ioType string, instance string) string {
// 	address := fmt.Sprintf("%s/%s/%s/"+CommandInputDiscovery+"/%s/%s",
// 		zone, publisherID, nodeID, ioType, instance)
// 	return address
// }

// MakeInputSetAddress for publishing or subscribing
func MakeInputSetAddress(zone string, publisherID string, nodeID string, ioType string, instance string) string {
	address := fmt.Sprintf("%s/%s/%s/"+CommandSet+"/%s/%s",
		zone, publisherID, nodeID, ioType, instance)
	return address
}

// // MakeNodeDiscoveryAddress for publishing
// func MakeNodeDiscoveryAddress(zoneID string, publisherID string, nodeID string) string {
// 	address := fmt.Sprintf("%s/%s/%s/"+CommandNodeDiscovery, zoneID, publisherID, nodeID)
// 	return address
// }
