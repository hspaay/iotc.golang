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
	OutputTypeRelay                  string = "relay"
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

// OutputBatchMessage message with multiple output events
type OutputBatchMessage struct {
	Address string `json:"address"` // Address of the publication: zone/publisher/node/$output/type/instance
	Batch   []struct {
		Timestasmp string            // Tunestamp if event
		Event      map[string]string // event values
	} `json:"batch"` // time ordered list of events
	Timestamp string `json:"timestamp"` // timestamp the batch is created
}

// OutputDiscoveryMessage with node output description
type OutputDiscoveryMessage struct {
	Address     string   `json:"address"`               // Address of the publication: zone/publisher/node/$output/type/instance
	DataType    DataType `json:"datatype,omitempty"`    // output value data type (DataType)
	Description string   `json:"description,omitempty"` // optional description for humans
	EnumValues  []string `json:"enumValues,omitempty"`  // possible enum output values for enum datatype
	Instance    string   `json:"instance"`              // instance identifier for multi-I/O nodes
	Max         float32  `json:"max,omitempty"`         // optional max value of output for numeric data types
	Min         float32  `json:"min,omitempty"`         // optional min value of output for numeric data types
	Timestamp   string   `json:"timestamp"`             // time the record is created
	OutputType  string   `json:"outputtype"`            // type of output as per OutputTypeXyz
	Unit        Unit     `json:"unit,omitempty"`        // unit of output value
}

// OutputEventMessage message with multiple output values
type OutputEventMessage struct {
	Address   string            `json:"address"` // Address of the publication: zone/publisher/node/$output/type/instance
	Event     map[string]string `json:"event"`
	Timestamp string            `json:"timestamp"`
}

// OutputForecast with forecasted values
// type OutputForecast []OutputValue

// OutputForecastMessage with prediction output values
type OutputForecastMessage struct {
	Address   string        `json:"address"` // Address of the publication: zone/publisher/node/$output/type/instance
	Duration  int           `json:"duration,omitempty"`
	Forecast  []OutputValue `json:"forecast"`  // list of timestamp and value pairs
	Timestamp string        `json:"timestamp"` // timestamp the forecast was created
	Unit      Unit          `json:"unit,omitempty"`
}

// OutputHistoryMessage with historical output value
type OutputHistoryMessage struct {
	Address   string        `json:"address"` // Address of the publication: zone/publisher/node/$output/type/instance
	Duration  int           `json:"duration,omitempty"`
	History   []OutputValue `json:"history"`
	Timestamp string        `json:"timestamp"`
	Unit      Unit          `json:"unit,omitempty"`
}

// OutputLatestMessage struct to send/receive the '$latest' command
type OutputLatestMessage struct {
	Address   string `json:"address"`   // Address of the publication: zone/publisher/node/$output/type/instance
	Timestamp string `json:"timestamp"` // timestamp of value
	Unit      Unit   `json:"unit,omitempty"`
	Value     string `json:"value"` // this can also be a string containing a list, eg "[ a, b, c ]""
}

// OutputValue struct for history and forecast
type OutputValue struct {
	Timestamp string `json:"timestamp"` // Timestamp of the value is ISO 8601
	Value     string `json:"value"`     // this can also be a string containing a list, eg "[ a, b, c ]""
	EpochTime int64  `json:"epoch"`     // seconds since jan 1st, 1970,
}
