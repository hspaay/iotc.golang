// Package nodes with handling of inputs
package nodes

import (
	"fmt"

	"github.com/hspaay/iotconnect.golang/standard"
)

// Input describing an input or output
type Input struct {
	Address     string        `json:"address"`               // I/O address
	Config      ConfigAttrMap `json:"config,omitempty"`      // Configuration of input or output
	DataType    DataType      `json:"datatype,omitempty"`    //
	Description string        `json:"description,omitempty"` // optional description
	EnumValues  []string      `json:"enum,omitempty"`        // enum valid values
	Instance    string        `json:"instance,omitempty"`    // instance identifier for multi-I/O nodes
	InputType   string        `json:"type,omitempty"`        // type of input or output as per IOTypeXyz
	NodeID      string        `json:"nodeID"`                // The node ID
	Unit        Unit          `json:"unit,omitempty"`        // unit of value
}

// DefaultInputInstance is the input instance identifier when only a single instance exists
const DefaultInputInstance = "0"

// NodeInput types
// These determine the available units and the datatype.
const (
	InputTypeUnknown string = "" // Not a known property type

	InputTypeChannel          string = "avchannel"
	InputTypeColor            string = "color"
	InputTypeColorTemperature string = "colortemperature"
	InputTypeCommand          string = "command"
	InputTypeDimmer           string = "dimmer"
	InputTypeElectricCurrent  string = "current"
	InputTypeElectricPower    string = "power"
	InputTypeHumidity         string = "humidity"
	InputTypeImage            string = "image"
	InputTypeLevel            string = "level" // multilevel sensor
	InputTypeLock             string = "lock"
	InputTypeMute             string = "avmute"
	InputTypeOnOffSwitch      string = "switch"
	InputTypePlay             string = "avplay"
	InputTypePushButton       string = "pushbutton" // with nr of pushes
	InputTypeTemperature      string = "temperature"
	InputTypeValue            string = "value" // generic value
	InputTypeVoltage          string = "voltage"
	InputTypeVolume           string = "volume"
	InputTypeWaterLevel       string = "waterlevel"
)

// MakeInputDiscoveryAddress for publishing or subscribing
func MakeInputDiscoveryAddress(zone string, publisherID string, nodeID string, ioType string, instance string) string {
	address := fmt.Sprintf("%s/%s/%s/"+standard.CommandInputDiscovery+"/%s/%s",
		zone, publisherID, nodeID, ioType, instance)
	return address
}

// NewInput instance
func NewInput(node *Node, inputType string, instance string) *Input {
	address := MakeInputDiscoveryAddress(node.Zone, node.PublisherID, node.ID, inputType, instance)
	io := &Input{
		Address:   address,
		Config:    ConfigAttrMap{},
		Instance:  instance,
		InputType: inputType,
		// History:  make([]*HistoryValue, 1),
	}
	return io
}
