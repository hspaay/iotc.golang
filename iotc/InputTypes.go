// Package iotc with IoTConnect input message type definitions and constants
package iotc

// DefaultInputInstance is the input instance identifier when only a single instance exists
const DefaultInputInstance = "0"

// InputType defines the convention names for output types
type InputType string

// NodeInput types
// These determine the available units and the datatype.
const (
	InputTypeUnknown InputType = "" // Not a known property type

	InputTypeChannel          InputType = "avchannel"        // select audio video input channel
	InputTypeColor            InputType = "color"            // set light color in hex: #RRGGBB
	InputTypeColorTemperature InputType = "colortemperature" // set light color temperature in kelvin
	InputTypeCommand          InputType = "command"          // issue input command
	InputTypeDimmer           InputType = "dimmer"           // control light dimmer 0-100%
	InputTypeHumidity         InputType = "humidity"         // humidity setting control 0-100%
	InputTypeImage            InputType = "image"            // image input
	InputTypeLevel            InputType = "level"            // multilevel input control
	InputTypeLock             InputType = "lock"             // lock "open" or "closed"
	InputTypeMute             InputType = "avmute"           // audi/video mute: "on" "off"
	InputTypeSwitch           InputType = "switch"           // set on/off switch: "on" "off"
	InputTypePlay             InputType = "avplay"           // audio/video play pushbutton
	InputTypePushButton       InputType = "pushbutton"       // push button with nr of pushes
	InputTypeRPM              InputType = "rpm"              // control rotations per minute
	InputTypeSpeed            InputType = "speed"            // control speed
	InputTypeTemperature      InputType = "temperature"      // set thermostat temperature
	InputTypeValue            InputType = "value"            // generic input value if not a level
	InputTypeVoltage          InputType = "voltage"          // set input control for voltage
	InputTypeWaterLevel       InputType = "waterlevel"       // set input control for water level
)

// InputDiscoveryMessage with node input description
type InputDiscoveryMessage struct {
	Address string `json:"address"` // Address of the input
	// Config      ConfigAttrMap `json:"config,omitempty"`      // Configuration of input
	DataType    DataType  `json:"dataType,omitempty"`    // input value data type
	Description string    `json:"description,omitempty"` // optional description
	EnumValues  []string  `json:"enumValues,omitempty"`  // enum valid input values for enum datatypes
	Instance    string    `json:"instance,omitempty"`    // instance identifier for multi-I/O nodes
	Max         float32   `json:"max,omitempty"`         // optional max value of input for numeric data types
	Min         float32   `json:"min,omitempty"`         // optional min value of input for numeric data types
	InputType   InputType `json:"inputType,omitempty"`   // type of input or output as per IOTypeXyz
	Unit        Unit      `json:"unit,omitempty"`        // unit of value
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
	FWVersion string `json:"fwVersion"` // firmware version
	Sender    string `json:"sender"`    // sending node: zone/publisher/nodeId
	Timestamp string `json:"timestamp"`
}
