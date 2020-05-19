// Package iotc with IoTConnect input message type definitions and constants
package iotc

// DefaultInputInstance is the input instance identifier when only a single instance exists
const DefaultInputInstance = "0"

// NodeInput types
// These determine the available units and the datatype.
const (
	InputTypeUnknown string = "" // Not a known property type

	InputTypeChannel          string = "avchannel"        // select audio video input channel
	InputTypeColor            string = "color"            // set light color in hex: #RRGGBB
	InputTypeColorTemperature string = "colortemperature" // set light color temperature in kelvin
	InputTypeCommand          string = "command"          // issue input command
	InputTypeDimmer           string = "dimmer"           // control light dimmer 0-100%
	InputTypeHumidity         string = "humidity"         // humidity setting control 0-100%
	InputTypeImage            string = "image"            // image input
	InputTypeLevel            string = "level"            // multilevel input control
	InputTypeLock             string = "lock"             // lock "open" or "closed"
	InputTypeMute             string = "avmute"           // audi/video mute: "on" "off"
	InputTypeSwitch           string = "switch"           // set on/off switch: "on" "off"
	InputTypePlay             string = "avplay"           // audio/video play pushbutton
	InputTypePushButton       string = "pushbutton"       // push button with nr of pushes
	InputTypeRPM              string = "rpm"              // control rotations per minute
	InputTypeSpeed            string = "speed"            // control speed
	InputTypeTemperature      string = "temperature"      // set thermostat temperature
	InputTypeValue            string = "value"            // generic input value if not a level
	InputTypeVoltage          string = "voltage"          // set input control for voltage
	InputTypeWaterLevel       string = "waterlevel"       // set input control for water level
)

// InputDiscoveryMessage with node input description
type InputDiscoveryMessage struct {
	Address string `json:"address"` // Address of the input
	// Config      ConfigAttrMap `json:"config,omitempty"`      // Configuration of input
	DataType    DataType `json:"datatype,omitempty"`    // input value data type
	Description string   `json:"description,omitempty"` // optional description
	EnumValues  []string `json:"enumValues,omitempty"`  // enum valid input values for enum datatypes
	Instance    string   `json:"instance,omitempty"`    // instance identifier for multi-I/O nodes
	Max         float32  `json:"max,omitempty"`         // optional max value of input for numeric data types
	Min         float32  `json:"min,omitempty"`         // optional min value of input for numeric data types
	InputType   string   `json:"inputtype,omitempty"`   // type of input or output as per IOTypeXyz
	Unit        Unit     `json:"unit,omitempty"`        // unit of value
}

// SetInputMessage to control an input
type SetInputMessage struct {
	Address   string `json:"address"` // zone/publisher/node/$set/type/instance
	Timestamp string `json:"timestamp"`
	Sender    string `json:"sender"` // sending node: zone/publisher/nodeId
	Value     string `json:"value"`  // this can also be a string containing a list, eg "[ a, b, c ]""
}

// UpgradeFirmwareMessage with node firmware
type UpgradeFirmwareMessage struct {
	Address   string `json:"address"`   // message address
	MD5       string `json:"md5"`       // firmware MD5
	Firmware  []byte `json:"firmware"`  // firmware code
	FWVersion string `json:"fwversion"` // firmware version
	Sender    string `json:"sender"`    // sending node: zone/publisher/nodeId
	Timestamp string `json:"timestamp"`
}
