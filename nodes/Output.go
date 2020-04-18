// Package nodes with handling of node outputs objects
package nodes

import (
	"fmt"

	"github.com/hspaay/iotconnect.golang/standard"
)

// Output description
type Output struct {
	Address     string        `json:"address"`               // I/O address
	Config      ConfigAttrMap `json:"config,omitempty"`      // Configuration of input or output
	DataType    DataType      `json:"datatype,omitempty"`    //
	Description string        `json:"description,omitempty"` // optional description
	EnumValues  []string      `json:"enum,omitempty"`        // enum valid values
	Instance    string        `json:"instance,omitempty"`    // instance identifier for multi-I/O nodes
	OutputType  string        `json:"type,omitempty"`        // type of input or output as per IOTypeXyz
	NodeID      string        `json:"nodeID"`                // The node ID
	Unit        Unit          `json:"unit,omitempty"`        // unit of value
}

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
var OutputTypeMap = map[string]struct {
	DataType    DataType
	DefaultUnit Unit
	UnitValues  string
}{
	OutputTypeUnknown:                {},
	OutputTypeAcceleration:           {DataType: DataTypeNumber, DefaultUnit: "m/s2"},
	OutputTypeAirQuality:             {DataType: DataTypeNumber, DefaultUnit: ""},
	OutputTypeAlarm:                  {DataType: DataTypeString},
	OutputTypeAtmosphericPressure:    {DataType: DataTypeNumber, DefaultUnit: UnitMillibar, UnitValues: UnitValuesAtmosphericPressure},
	OutputTypeBattery:                {DataType: DataTypeNumber, DefaultUnit: UnitPercent},
	OutputTypeCarbonDioxideLevel:     {DataType: DataTypeNumber, DefaultUnit: UnitPartsPerMillion},
	OutputTypeCarbonMonoxideDetector: {DataType: DataTypeBool},
	OutputTypeCarbonMonoxideLevel:    {DataType: DataTypeNumber, DefaultUnit: UnitPartsPerMillion},
	OutputTypeChannel:                {DataType: DataTypeNumber},
	OutputTypeColor:                  {DataType: DataTypeString},
	OutputTypeColorTemperature:       {DataType: DataTypeNumber, DefaultUnit: UnitKelvin},
	OutputTypeConnections:            {DataType: DataTypeNumber},
	OutputTypeContact:                {DataType: DataTypeBool},
	OutputTypeCPULevel:               {DataType: DataTypeNumber, DefaultUnit: UnitPercent},
	OutputTypeDewpoint:               {DataType: DataTypeNumber, DefaultUnit: UnitCelcius, UnitValues: UnitValuesTemperature},
	OutputTypeDimmer:                 {DataType: DataTypeNumber, DefaultUnit: UnitPercent},
	OutputTypeDoorWindowSensor:       {DataType: DataTypeBool},
	OutputTypeElectricCurrent:        {DataType: DataTypeBool, DefaultUnit: UnitAmp},
	OutputTypeElectricEnergy:         {DataType: DataTypeNumber, DefaultUnit: UnitKWH},
	OutputTypeElectricPower:          {DataType: DataTypeNumber, DefaultUnit: UnitWatt},
	OutputTypeErrors:                 {DataType: DataTypeNumber},
	OutputTypeHeatIndex:              {DataType: DataTypeNumber, DefaultUnit: UnitCelcius, UnitValues: "C, F"},
	OutputTypeHue:                    {DataType: DataTypeString},
	OutputTypeHumidex:                {DataType: DataTypeNumber, DefaultUnit: UnitCelcius, UnitValues: "C, F"},
	OutputTypeHumidity:               {DataType: DataTypeNumber, DefaultUnit: UnitPercent},
	OutputTypeImage:                  {DataType: DataTypeBytes, DefaultUnit: "", UnitValues: UnitValuesImage},
	OutputTypeLatency:                {DataType: DataTypeNumber, DefaultUnit: UnitSecond},
	OutputTypeLevel:                  {DataType: DataTypeNumber, DefaultUnit: UnitPercent},
	OutputTypeLocation:               {DataType: DataTypeString},
	OutputTypeLock:                   {DataType: DataTypeString},
	OutputTypeLuminance:              {DataType: DataTypeNumber, DefaultUnit: UnitLux},
	OutputTypeMotion:                 {DataType: DataTypeBool},
	OutputTypeMute:                   {DataType: DataTypeBool},
	OutputTypeOnOffSwitch:            {DataType: DataTypeBool},
	OutputTypePlay:                   {DataType: DataTypeBool},
	OutputTypePushButton:             {DataType: DataTypeNumber},
	OutputTypeRain:                   {DataType: DataTypeNumber, DefaultUnit: UnitMeter, UnitValues: "m"},
	OutputTypeSaturation:             {DataType: DataTypeString},
	OutputTypeScale:                  {DataType: DataTypeNumber, DefaultUnit: UnitKG, UnitValues: UnitValuesWeight},
	OutputTypeSnow:                   {DataType: DataTypeNumber, DefaultUnit: UnitMeter, UnitValues: "m"},
	OutputTypeSignalStrength:         {DataType: DataTypeNumber, DefaultUnit: "dBm"},
	OutputTypeSmokeDetector:          {DataType: DataTypeBool},
	OutputTypeSoundDetector:          {DataType: DataTypeBool},
	OutputTypeTemperature:            {DataType: DataTypeNumber},
	OutputTypeUltraviolet:            {DataType: DataTypeNumber},
	OutputTypeVibrationDetector:      {DataType: DataTypeNumber},
	OutputTypeValue:                  {DataType: DataTypeNumber},
	OutputTypeVoltage:                {DataType: DataTypeNumber, DefaultUnit: UnitVolt},
	OutputTypeVolume:                 {DataType: DataTypeNumber, DefaultUnit: UnitPercent},
	OutputTypeWaterLevel:             {DataType: DataTypeNumber, DefaultUnit: UnitMeter, UnitValues: UnitValuesLength},
	OutputTypeWindHeading:            {DataType: DataTypeNumber, DefaultUnit: UnitDegree},
	OutputTypeWindSpeed:              {DataType: DataTypeNumber, DefaultUnit: UnitSpeed, UnitValues: UnitValuesSpeed},
}

// MakeOutputDiscoveryAddress for publishing or subscribing
func MakeOutputDiscoveryAddress(zone string, publisherID string, nodeID string, ioType string, instance string) string {
	address := fmt.Sprintf("%s/%s/%s/"+standard.CommandOutputDiscovery+"/%s/%s",
		zone, publisherID, nodeID, ioType, instance)
	return address
}

// NewOutput instance
func NewOutput(node *Node, outputType string, instance string) *Output {
	address := MakeOutputDiscoveryAddress(node.Zone, node.PublisherID, node.ID, outputType, instance)
	io := &Output{
		Address:    address,
		Config:     ConfigAttrMap{},
		Instance:   instance,
		OutputType: outputType,
		// History:  make([]*HistoryValue, 1),
	}
	return io
}
