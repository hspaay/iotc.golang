// Package iotc with IoTConnect output message type definitions and constants
package iotc

// DefaultOutputInstance is the output instance identifier when only a single instance exists
const DefaultOutputInstance = "0"

// NodeOutput and actuator types
// These determine the available units and the datatype.
const (
	OutputTypeUnknown string = "" // Not a known property type

	OutputTypeAcceleration           string = "acceleration"
	OutputTypeAirQuality             string = "airquality"
	OutputTypeAlarm                  string = "alarm"
	OutputTypeAtmosphericPressure    string = "atmosphericpressure"
	OutputTypeBattery                string = "battery"
	OutputTypeCarbonDioxideLevel     string = "co2level"
	OutputTypeCarbonMonoxideDetector string = "codetector"
	OutputTypeCarbonMonoxideLevel    string = "colevel"
	OutputTypeChannel                string = "avchannel"
	OutputTypeColor                  string = "color"
	OutputTypeColorTemperature       string = "colortemperature"
	OutputTypeConnections            string = "connections"
	OutputTypeContact                string = "contact"
	OutputTypeCPULevel               string = "cpulevel"
	OutputTypeDewpoint               string = "dewpoint"
	OutputTypeDimmer                 string = "dimmer"
	OutputTypeDoorWindowSensor       string = "doorwindowsensor"
	OutputTypeElectricCurrent        string = "current"
	OutputTypeElectricEnergy         string = "energy"
	OutputTypeElectricPower          string = "power"
	OutputTypeErrors                 string = "errors"
	OutputTypeHeatIndex              string = "heatindex"
	OutputTypeHue                    string = "hue"
	OutputTypeHumidex                string = "humidex"
	OutputTypeHumidity               string = "humidity"
	OutputTypeImage                  string = "image"
	OutputTypeLatency                string = "Latency"
	OutputTypeLevel                  string = "level" // multilevel sensor
	OutputTypeLocation               string = "location"
	OutputTypeLock                   string = "lock"
	OutputTypeLuminance              string = "luminance"
	OutputTypeMotion                 string = "motion"
	OutputTypeMute                   string = "avmute"
	OutputTypeOnOffSwitch            string = "switch"
	OutputTypePlay                   string = "avplay"
	OutputTypePushButton             string = "pushbutton" // with nr of pushes
	OutputTypeRain                   string = "rain"
	OutputTypeSaturation             string = "saturation"
	OutputTypeScale                  string = "scale"
	OutputTypeSignalStrength         string = "signalstrength"
	OutputTypeSmokeDetector          string = "smokedetector"
	OutputTypeSnow                   string = "snow"
	OutputTypeSoundDetector          string = "sounddetector"
	OutputTypeTemperature            string = "temperature"
	OutputTypeUltraviolet            string = "ultraviolet"
	OutputTypeVibrationDetector      string = "vibrationdetector"
	OutputTypeValue                  string = "value" // generic value
	OutputTypeVoltage                string = "voltage"
	OutputTypeVolume                 string = "volume"
	OutputTypeWaterLevel             string = "waterlevel"
	OutputTypeWeather                string = "weather" // description of weather, eg sunny
	OutputTypeWindHeading            string = "windheading"
	OutputTypeWindSpeed              string = "windspeed"
)

// OutputTypeMap defines data type and unit for an IOType
// Todo: option to download from file
// var OutputTypeMap = map[string]struct {
// 	DataType    DataType
// 	DefaultUnit Unit
// 	UnitValues  string
// }{
// 	OutputTypeUnknown:                {},
// 	OutputTypeAcceleration:           {DataType: DataTypeNumber, DefaultUnit: "m/s2"},
// 	OutputTypeAirQuality:             {DataType: DataTypeNumber, DefaultUnit: ""},
// 	OutputTypeAlarm:                  {DataType: DataTypeString},
// 	OutputTypeAtmosphericPressure:    {DataType: DataTypeNumber, DefaultUnit: UnitMillibar, UnitValues: UnitValuesAtmosphericPressure},
// 	OutputTypeBattery:                {DataType: DataTypeNumber, DefaultUnit: UnitPercent},
// 	OutputTypeCarbonDioxideLevel:     {DataType: DataTypeNumber, DefaultUnit: UnitPartsPerMillion},
// 	OutputTypeCarbonMonoxideDetector: {DataType: DataTypeBool},
// 	OutputTypeCarbonMonoxideLevel:    {DataType: DataTypeNumber, DefaultUnit: UnitPartsPerMillion},
// 	OutputTypeChannel:                {DataType: DataTypeNumber},
// 	OutputTypeColor:                  {DataType: DataTypeString},
// 	OutputTypeColorTemperature:       {DataType: DataTypeNumber, DefaultUnit: UnitKelvin},
// 	OutputTypeConnections:            {DataType: DataTypeNumber},
// 	OutputTypeContact:                {DataType: DataTypeBool},
// 	OutputTypeCPULevel:               {DataType: DataTypeNumber, DefaultUnit: UnitPercent},
// 	OutputTypeDewpoint:               {DataType: DataTypeNumber, DefaultUnit: UnitCelcius, UnitValues: UnitValuesTemperature},
// 	OutputTypeDimmer:                 {DataType: DataTypeNumber, DefaultUnit: UnitPercent},
// 	OutputTypeDoorWindowSensor:       {DataType: DataTypeBool},
// 	OutputTypeElectricCurrent:        {DataType: DataTypeBool, DefaultUnit: UnitAmp},
// 	OutputTypeElectricEnergy:         {DataType: DataTypeNumber, DefaultUnit: UnitKWH},
// 	OutputTypeElectricPower:          {DataType: DataTypeNumber, DefaultUnit: UnitWatt},
// 	OutputTypeErrors:                 {DataType: DataTypeNumber},
// 	OutputTypeHeatIndex:              {DataType: DataTypeNumber, DefaultUnit: UnitCelcius, UnitValues: "C, F"},
// 	OutputTypeHue:                    {DataType: DataTypeString},
// 	OutputTypeHumidex:                {DataType: DataTypeNumber, DefaultUnit: UnitCelcius, UnitValues: "C, F"},
// 	OutputTypeHumidity:               {DataType: DataTypeNumber, DefaultUnit: UnitPercent},
// 	OutputTypeImage:                  {DataType: DataTypeBytes, DefaultUnit: "", UnitValues: UnitValuesImage},
// 	OutputTypeLatency:                {DataType: DataTypeNumber, DefaultUnit: UnitSecond},
// 	OutputTypeLevel:                  {DataType: DataTypeNumber, DefaultUnit: UnitPercent},
// 	OutputTypeLocation:               {DataType: DataTypeString},
// 	OutputTypeLock:                   {DataType: DataTypeString},
// 	OutputTypeLuminance:              {DataType: DataTypeNumber, DefaultUnit: UnitLux},
// 	OutputTypeMotion:                 {DataType: DataTypeBool},
// 	OutputTypeMute:                   {DataType: DataTypeBool},
// 	OutputTypeOnOffSwitch:            {DataType: DataTypeBool},
// 	OutputTypePlay:                   {DataType: DataTypeBool},
// 	OutputTypePushButton:             {DataType: DataTypeNumber},
// 	OutputTypeRain:                   {DataType: DataTypeNumber, DefaultUnit: UnitMeter, UnitValues: "m"},
// 	OutputTypeSaturation:             {DataType: DataTypeString},
// 	OutputTypeScale:                  {DataType: DataTypeNumber, DefaultUnit: UnitKG, UnitValues: UnitValuesWeight},
// 	OutputTypeSnow:                   {DataType: DataTypeNumber, DefaultUnit: UnitMeter, UnitValues: "m"},
// 	OutputTypeSignalStrength:         {DataType: DataTypeNumber, DefaultUnit: "dBm"},
// 	OutputTypeSmokeDetector:          {DataType: DataTypeBool},
// 	OutputTypeSoundDetector:          {DataType: DataTypeBool},
// 	OutputTypeTemperature:            {DataType: DataTypeNumber},
// 	OutputTypeUltraviolet:            {DataType: DataTypeNumber},
// 	OutputTypeVibrationDetector:      {DataType: DataTypeNumber},
// 	OutputTypeValue:                  {DataType: DataTypeNumber},
// 	OutputTypeVoltage:                {DataType: DataTypeNumber, DefaultUnit: UnitVolt},
// 	OutputTypeVolume:                 {DataType: DataTypeNumber, DefaultUnit: UnitPercent},
// 	OutputTypeWaterLevel:             {DataType: DataTypeNumber, DefaultUnit: UnitMeter, UnitValues: UnitValuesLength},
// 	OutputTypeWindHeading:            {DataType: DataTypeNumber, DefaultUnit: UnitDegree},
// 	OutputTypeWindSpeed:              {DataType: DataTypeNumber, DefaultUnit: UnitSpeed, UnitValues: UnitValuesSpeed},
// }

// OutputValue of node output
type OutputValue struct {
	// Timestamp of the value is ISO 8601
	Timestamp string `json:"timestamp"`
	Value     string `json:"value"` // this can also be a string containing a list, eg "[ a, b, c ]""
	EpochTime int64  `json:"epoch"` // seconds since jan 1st, 1970,
}

// OutputDiscoveryMessage with node output description
type OutputDiscoveryMessage struct {
	Address     string   `json:"address"`               // I/O address
	DataType    string   `json:"datatype,omitempty"`    //
	Description string   `json:"description,omitempty"` // optional description
	EnumValues  []string `json:"enum,omitempty"`        // enum valid values
	Instance    string   `json:"instance,omitempty"`    // instance identifier for multi-I/O nodes
	NodeID      string   `json:"nodeID"`                // The node ID this output is part of
	OutputType  string   `json:"type,omitempty"`        // type of input or output as per IOTypeXyz
	Unit        Unit     `json:"unit,omitempty"`        // unit of output value
}

// OutputEventMessage message with multiple output values
type OutputEventMessage struct {
	Address   string            `json:"address"`
	Event     map[string]string `json:"event"`
	Sender    string            `json:"sender"`
	Timestamp string            `json:"timestamp"`
}

// OutputHistoryList List of history values
type OutputHistoryList []OutputValue

// OutputForecastMessage with prediction output values
type OutputForecastMessage struct {
	Address   string            `json:"address"`
	Duration  int               `json:"duration,omitempty"`
	Forecast  OutputHistoryList `json:"forecast"`
	Sender    string            `json:"sender"`
	Timestamp string            `json:"timestamp"`
	Unit      Unit              `json:"unit,omitempty"`
}

// OutputHistoryMessage with historical output value
type OutputHistoryMessage struct {
	Address   string            `json:"address"`
	Duration  int               `json:"duration,omitempty"`
	History   OutputHistoryList `json:"history"`
	Sender    string            `json:"sender"`
	Timestamp string            `json:"timestamp"`
	Unit      Unit              `json:"unit,omitempty"`
}

// OutputLatestMessage struct to send/receive the '$latest' command
type OutputLatestMessage struct {
	Address   string `json:"address"`
	Sender    string `json:"sender"`
	Timestamp string `json:"timestamp"` // timestamp of value
	Unit      Unit   `json:"unit,omitempty"`
	Value     string `json:"value"` // this can also be a string containing a list, eg "[ a, b, c ]""
}
