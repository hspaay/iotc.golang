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
	// Config      ConfigAttrMap `json:"config,omitempty"`      // Configuration of input or output
	DataType    DataType `json:"datatype,omitempty"`    //
	Description string   `json:"description,omitempty"` // optional description
	EnumValues  []string `json:"enum,omitempty"`        // enum valid values
	Instance    string   `json:"instance,omitempty"`    // instance identifier for multi-I/O nodes
	InputType   string   `json:"type,omitempty"`        // type of input or output as per IOTypeXyz
	NodeID      string   `json:"nodeID"`                // The node ID
	Unit        Unit     `json:"unit,omitempty"`        // unit of value
}

// SetInputMessage to control an input
type SetInputMessage struct {
	Address   string `json:"address"` // zone/publisher/node/$set/type/instance
	Sender    string `json:"sender"`
	Timestamp string `json:"timestamp"`
	Value     string `json:"value"` // this can also be a string containing a list, eg "[ a, b, c ]""
}

// UpgradeFirmwareMessage with node firmware
type UpgradeFirmwareMessage struct {
	Address   string `json:"address"`   // message address
	MD5       string `json:"md5"`       // firmware MD5
	Firmware  []byte `json:"firmware"`  // firmware code
	FWVersion string `json:"fwversion"` // firmware version
	Sender    string `json:"sender"`
	Timestamp string `json:"timestamp"`
}
