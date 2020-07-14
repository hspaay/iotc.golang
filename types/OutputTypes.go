// Package types with output message type definitions and constants
package types

// DefaultOutputInstance is the output instance identifier when only a single instance exists
const DefaultOutputInstance = "0"

// // OutputAttr output configuration attributes
// type OutputAttr string

// // Output Attributes
// const (
// 	OutputAttrPublishBatch   OutputAttr = "publishBatch"   // int, publish batch of N events
// 	OutputAttrPublishEvent   OutputAttr = "publishEvent"   // bool, include output in $event publication
// 	OutputAttrPublishFile    OutputAttr = "publishFile"    // string, save output to local file, "" to disable
// 	OutputAttrPublishHistory OutputAttr = "publishHistory" // bool, publish output with $history message
// 	OutputAttrPublishLatest  OutputAttr = "publishLatest"  // bool, publish output with $latest message
// 	OutputAttrPublishRaw     OutputAttr = "publishRaw"     // bool, publish output with $raw message
// )

// OutputType defines the convention names for output types
type OutputType string

// NodeOutput and actuator types
// These determine the available units and the datatype.
const (
	OutputTypeUnknown string = "" // Not a known property type

	OutputTypeAcceleration           OutputType = "acceleration"
	OutputTypeAirQuality             OutputType = "airquality"
	OutputTypeAlarm                  OutputType = "alarm"
	OutputTypeAtmosphericPressure    OutputType = "atmosphericpressure"
	OutputTypeBattery                OutputType = "battery"
	OutputTypeCarbonDioxideLevel     OutputType = "co2level"
	OutputTypeCarbonMonoxideDetector OutputType = "codetector"
	OutputTypeCarbonMonoxideLevel    OutputType = "colevel"
	OutputTypeChannel                OutputType = "avchannel"
	OutputTypeColor                  OutputType = "color"
	OutputTypeColorTemperature       OutputType = "colortemperature"
	OutputTypeConnections            OutputType = "connections"
	OutputTypeCPULevel               OutputType = "cpulevel"
	OutputTypeDewpoint               OutputType = "dewpoint"
	OutputTypeDimmer                 OutputType = "dimmer"
	OutputTypeDoorWindowSensor       OutputType = "doorwindowsensor"
	OutputTypeElectricCurrent        OutputType = "current"
	OutputTypeElectricEnergy         OutputType = "energy"
	OutputTypeElectricPower          OutputType = "power"
	OutputTypeErrors                 OutputType = "errors"
	OutputTypeHeatIndex              OutputType = "heatindex"
	OutputTypeHue                    OutputType = "hue"
	OutputTypeHumidex                OutputType = "humidex"
	OutputTypeHumidity               OutputType = "humidity"
	OutputTypeImage                  OutputType = "image"
	OutputTypeLatency                OutputType = "latency"
	OutputTypeLevel                  OutputType = "level" // multilevel sensor
	OutputTypeLocation               OutputType = "location"
	OutputTypeLock                   OutputType = "lock"
	OutputTypeLuminance              OutputType = "luminance"
	OutputTypeMotion                 OutputType = "motion"
	OutputTypeMute                   OutputType = "avmute"
	OutputTypeOnOffSwitch            OutputType = "switch"
	OutputTypePlay                   OutputType = "avplay"
	OutputTypePushButton             OutputType = "pushbutton" // with nr of pushes
	OutputTypeRain                   OutputType = "rain"
	OutputTypeRelay                  OutputType = "relay"
	OutputTypeSaturation             OutputType = "saturation"
	OutputTypeScale                  OutputType = "scale"
	OutputTypeSignalStrength         OutputType = "signalstrength"
	OutputTypeSmokeDetector          OutputType = "smokedetector"
	OutputTypeSnow                   OutputType = "snow"
	OutputTypeSoundDetector          OutputType = "sounddetector"
	OutputTypeSwitch                 OutputType = "switch" // on/off switch: "on" "off"
	OutputTypeTemperature            OutputType = "temperature"
	OutputTypeUltraviolet            OutputType = "ultraviolet"
	OutputTypeVibrationDetector      OutputType = "vibrationdetector"
	OutputTypeValue                  OutputType = "value" // generic value
	OutputTypeVoltage                OutputType = "voltage"
	OutputTypeVolume                 OutputType = "volume"
	OutputTypeWaterLevel             OutputType = "waterlevel"
	OutputTypeWeather                OutputType = "weather" // description of weather, eg sunny
	OutputTypeWindHeading            OutputType = "windheading"
	OutputTypeWindSpeed              OutputType = "windspeed"
)

// OutputTypeMap defines data type and unit for an IOType
// Todo: option to download from file
// var OutputTypeMap = map[OutputType]struct {
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
	Address    string        `json:"address"`              // Address of the publication: zone/publisher/node/$output/type/instance
	Attr       NodeAttrMap   `json:"attr,omitempty"`       // Attributes describing this output
	Config     ConfigAttrMap `json:"config,omitempty"`     // Optional configuration of output
	DataType   DataType      `json:"dataType,omitempty"`   // output value data type, default is string
	EnumValues []string      `json:"enumValues,omitempty"` // possible enum output values for enum datatype
	Max        float32       `json:"max,omitempty"`        // optional max value of output for numeric data types
	Min        float32       `json:"min,omitempty"`        // optional min value of output for numeric data types
	Timestamp  string        `json:"timestamp"`            // time the record is last updated
	Unit       Unit          `json:"unit,omitempty"`       // unit of output value
	// For convenience, filled when registering or receiving
	NodeID      string     `json:"-"`
	PublisherID string     `json:"-"`
	OutputType  OutputType `json:"-"`
	Instance    string     `json:"-"`
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
