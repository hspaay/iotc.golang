// Package standard with constants of the iotconnect standard
package standard

import (
	"fmt"
)

// TimeFormat for publishing messages
const TimeFormat = "2006-01-02T15:04:05.000-0700"

// publication address reserved keywords
const (
	CommandConfigure       = "$configure" // node configuration, payload is ConfigureMessage
	CommandEvent           = "$event"     // node outputs event, payload is EventMessage
	CommandHistory         = "$history"   // output history, payload is HistoryMessage
	CommandInputDiscovery  = "$input"     // input discovery, payload is InOutput object
	CommandLatest          = "$latest"    // latest output, payload is latest message
	CommandNodeDiscovery   = "$node"      // node discovery, payload is Node object
	CommandOutputDiscovery = "$output"    // output discovery, payload output definition
	CommandSet             = "$set"       // control input command, payload is input value
	CommandUpgrade         = "$upgrade"   // perform firmware upgrade, payload is UpgradeMessage
	CommandValue           = "$value"     // raw output value
	// LocalZone ID for local-only zones (eg, no sharing outside this zone)
	LocalZoneID = "$local"
	// PublisherNodeID to use when none is provided
	PublisherNodeID = "$publisher" // reserved node ID for publishers
)

// Standard attribute names for device and sensor attributes and config
const (
	// Status attributes that mutate easily
	AttrNameAlert       string = "alert"     // alert provided with error count
	AttrNameLatencyMSec string = "latency"   // duration connect to sensor in milliseconds
	AttrNameRunState    string = "runstate"  // Node is connected, ready, lost
	AttrNameErrorCount  string = "errors"    // nr errors reported on this device
	AttrNameHealth      string = "health"    // health status of the device 0-100%
	AttrNameLastSeen    string = "lastseen"  // ISO time the device was last seen
	AttrNameNeighbors   string = "neighbors" // mesh network device neighbors list
	AttrNameRxCount     string = "received"  // Nr of messages received from device
	AttrNameTxCount     string = "sent"      // Nr of messages send to device
	AttrNameValue       string = "value"     // configuration attribute with the configuration value

	// Information attributes that describe the device or sensor
	AttrNameAddress      string = "address"
	AttrNameAlias        string = "alias"        // node alias for publishing inputs and outputs
	AttrNameDescription  string = "description"  // device description
	AttrNameDisabled     string = "disabled"     // device or sensor is disabled
	AttrNameNodeType     string = "type"         // type of node. See NodeTypeXxx
	AttrNameHostname     string = "hostname"     // network device hostname
	AttrNameLocalIP      string = "localip"      // for IP nodes
	AttrNameMAC          string = "mac"          // for IP nodes
	AttrNameManufacturer string = "manufacturer" // device manufacturer
	AttrNameModel        string = "model"        // device model
	AttrNameName         string = "name"         // name of device, sensor
	AttrNamePowerSource  string = "powersource"  // battery, usb, mains
	AttrNameProduct      string = "product"      // device product or model name
	AttrNamePublicKey    string = "publickey"    // public key for encrypting sensitive configuration settings
	AttrNameVersion      string = "version"      // device/service firmware version

	// Configuration attributes for configuration of adapters, services, nodes or sensors
	AttrNameColor        string = "color"        // color in hex notation
	AttrNameFilename     string = "filename"     // filename to write images or other values to
	AttrNameLocation     string = "location"     // name of the location
	AttrNameLoginName    string = "loginname"    // login name to connect to the device. Value is not published
	AttrNameMax          string = "max"          // maximum value of sensor or config
	AttrNameMin          string = "min"          // minimum value of sensor or config
	AttrNameLatLon       string = "latlon"       // latitude, longitude of the device for display on a map r/w
	AttrNameLight        string = "light"        // light configuration of device (Camera IR, LED indicator)
	AttrNamePollInterval string = "pollinterval" // polling interval in seconds
	AttrNameSubnet       string = "subnet"       // IP subnets configuration
	AttrNamePassword     string = "password"     // password to connect. Value is not published.
)

// DataType of configuration values.
type DataType string

const (
	DataTypeBool   DataType = "boolean" // value is true/false, 1/0, on/off
	DataTypeBytes  DataType = "bytes"   // value is encoded byte array
	DataTypeDate   DataType = "date"    // ISO8601 date YYYY-MM-DDTHH:MM:SS.mmmZ
	DataTypeEnum   DataType = "enum"    // value is one of a predefined set of string values, published in the 'enum info field'
	DataTypeInt    DataType = "int"     // value is an integer number
	DataTypeNumber DataType = "number"  // value is a float number
	DataTypeSecret DataType = "secret"  // a secret string that is not published
	DataTypeString DataType = "string"  // value is a string
	DataTypeVector DataType = "vector"  // 3D vector (x, y, z) or (lat, lon, 0)
	DataTypeJSON   DataType = "json"    // value is a json object
)

// NodeType identifying  the purpose of the node
// Based on the primary role of the device.
type NodeType string

const (
	NodeTypeAlarm     = "alarm"     // an alarm emitter
	NodeTypeAVControl = "avcontrol" // Audio/Video controller
	NodeTypeBeacon    = "beacon"    // device is a location beacon
	NodeTypeButton    = "button"    // device is a physical button device with one or more buttons
	NodeTypeAdapter   = "adapter"   // software adapter, eg virtual device
	//NodeTypeController = "controller"    // software adapter, eg virtual device
	NodeTypePhone          = "phone"         // device is a phone
	NodeTypeCamera         = "camera"        // Node with camera
	NodeTypeComputer       = "computer"      // General purpose computer
	NodeTypeDimmer         = "dimmer"        // light dimmer
	NodeTypeGateway        = "gateway"       // Node is a gateway for other nodes (onewire, zwave, etc)
	NodeTypeKeyPad         = "keypad"        // Entry key pad
	NodeTypeLock           = "lock"          // Electronic door lock
	NodeTypeMultiSensor    = "multisensor"   // Node with multiple sensors
	NodeTypeNetRouter      = "networkrouter" // Node is a network router
	NodeTypeNetSwitch      = "networkswitch" // Node is a network switch
	NodeTypeNetWifiAP      = "wifiap"        // Node is a wireless access point
	NodeTypePowerMeter     = "powermeter"    // Node is a power meter
	NodeTypeRepeater       = "repeater"      // Node is a zwave or other signal repeater
	NodeTypeReceiver       = "receiver"      // Node is a (not so) smart radio/receiver/amp (eg, denon)
	NodeTypeSensor         = "sensor"        // Node is a single sensor (volt,...)
	NodeTypeSmartLight     = "smartlight"    // Node is a smart light, eg philips hue
	NodeTypeSwitch         = "switch"        // Node is a physical on/off switch
	NodeTypeThermometer    = "thermometer"   // Node is a temperature meter
	NodeTypeThermostat     = "thermostat"    // Node is a thermostat control unit
	NodeTypeTV             = "tv"            // Node is a (not so) smart TV
	NodeTypeUnknown        = "unknown"
	NodeTypeWallpaper      = "wallpaper"  // Node is a wallpaper montage of multiple images
	NodeTypeWaterValve     = "watervalve" // Water valve control unit
	NodeTypeWeatherStation = "weatherstation"
)

//
//// Data types of configuration values.
//type DataType string
//const (
//  DataTypeBool    DataType = "boolean" // value is true/false, 1/0, on/off
//  DataTypeBytes   DataType = "bytes"   // value is encoded byte array
//  DataTypeDate    DataType = "date"    // ISO8601 date YYYY-MM-DDTHH:MM:SS.mmmZ
//  DataTypeEnum    DataType = "enum"    // value is one of a predefined set of string values, published in the 'enum info field'
//  DataTypeNumber  DataType = "number"  // value is an integer or float number
//  DataTypeOnOff   DataType = "onoff"   // value is On or Off
//  DataTypeSecret  DataType = "secret"  // a secret string that is not published
//  DataTypeString  DataType = "string"  // value is a string
//  DataTypeVector  DataType = "vector"  // 3D vector (x, y, z) or (lat, lon, 0)
//  DataTypeJson    DataType = "json"    // value is a json object
//)

// RunState constants with Node running status
type RunState string

const (
	RunStateAlert        RunState = "Alert"        // Node is connected but something is wrong
	RunStateReady        RunState = "ready"        // Node is ready for use
	RunStateDisconnected RunState = "disconnected" // Node has cleanly disconnected
	RunStateFailed       RunState = "failed"       // Node failed to start
	RunStateInitializing RunState = "initializing" // Node is initializing
	RunStateLost         RunState = "lost"         // Node connection lost
	RunStateSleeping     RunState = "sleeping"     // Node has gone into sleep mode, often a battery powered devie
)

// Unit constants with unit names.
// These are defined with the sensor type
type Unit string

const (
	UnitNone            Unit = ""
	UnitAmp             Unit = "A"
	UnitCelcius         Unit = "C"
	UnitCandela         Unit = "cd"
	UnitCount           Unit = "#"
	UnitDegree          Unit = "Degree"
	UnitFahrenheit      Unit = "F"
	UnitFeet            Unit = "ft"
	UnitGallon          Unit = "Gal"
	UnitJpeg            Unit = "jpeg"
	UnitKelvin          Unit = "K"
	UnitKmPerHour       Unit = "Kph"
	UnitLiter           Unit = "L"
	UnitMercury         Unit = "hg"
	UnitMeter           Unit = "m"
	UnitMetersPerSecond Unit = "m/s"
	UnitMilesPerHour    Unit = "mph"
	UnitMillibar        Unit = "mbar"
	UnitMole            Unit = "mol"
	UnitPartsPerMillion Unit = "ppm"
	UnitPng             Unit = "png"
	UnitKWH             Unit = "KWh"
	UnitKG              Unit = "kg"
	UnitLux             Unit = "lux"
	UnitPascal          Unit = "Pa"
	UnitPercent         Unit = "%"
	UnitPounds          Unit = "lbs"
	UnitSpeed           Unit = "m/s"
	UnitPSI             Unit = "psi"
	UnitSecond          Unit = "s "
	UnitVolt            Unit = "V"
	UnitWatt            Unit = "W"
)

var (
	// UnitValuesAtmosphericPressure unit values for atmospheric pressure
	UnitValuesAtmosphericPressure = fmt.Sprintf("%s, %s, %s", UnitMillibar, UnitMercury, UnitPSI)
	UnitValuesImage               = fmt.Sprintf("%s, %s", UnitJpeg, UnitPng)
	UnitValuesLength              = fmt.Sprintf("%s, %s", UnitMeter, UnitFeet)
	UnitValuesSpeed               = fmt.Sprintf("%s, %s, %s", UnitMetersPerSecond, UnitKmPerHour, UnitMilesPerHour)
	UnitValuesTemperature         = fmt.Sprintf("%s, %s", UnitCelcius, UnitFahrenheit)
	UnitValuesWeight              = fmt.Sprintf("%s, %s", UnitKG, UnitPounds)
	UnitValuesVolume              = fmt.Sprintf("%s, %s", UnitLiter, UnitGallon)
)

// NodeInOutput and actuator types
// These determine the available units and the datatype.
const (
	IOTypeUnknown string = "" // Not a known property type

	IOTypeAcceleration           string = "acceleration"
	IOTypeAirQuality             string = "airquality"
	IOTypeAlarm                  string = "alarm"
	IOTypeAtmosphericPressure    string = "atmosphericpressure"
	IOTypeBattery                string = "battery"
	IOTypeCarbonDioxideLevel     string = "co2level"
	IOTypeCarbonMonoxideDetector string = "codetector"
	IOTypeCarbonMonoxideLevel    string = "colevel"
	IOTypeChannel                string = "avchannel"
	IOTypeColor                  string = "color"
	IOTypeColorTemperature       string = "colortemperature"
	IOTypeConnections            string = "connections"
	IOTypeContact                string = "contact"
	IOTypeCPULevel               string = "cpulevel"
	IOTypeCommand                string = "command"
	IOTypeDewpoint               string = "dewpoint"
	IOTypeDimmer                 string = "dimmer"
	IOTypeDoorWindowSensor       string = "doorwindowsensor"
	IOTypeElectricCurrent        string = "current"
	IOTypeElectricEnergy         string = "energy"
	IOTypeElectricPower          string = "power"
	IOTypeErrors                 string = "errors"
	IOTypeHeatIndex              string = "heatindex"
	IOTypeHue                    string = "hue"
	IOTypeHumidex                string = "humidex"
	IOTypeHumidity               string = "humidity"
	IOTypeImage                  string = "image"
	IOTypeLatency                string = "Latency"
	IOTypeLevel                  string = "level" // multilevel sensor
	IOTypeLocation               string = "location"
	IOTypeLock                   string = "lock"
	IOTypeLuminance              string = "luminance"
	IOTypeMotion                 string = "motion"
	IOTypeMute                   string = "avmute"
	IOTypeOnOffSwitch            string = "switch"
	IOTypePlay                   string = "avplay"
	IOTypePushButton             string = "pushbutton" // with nr of pushes
	IOTypeSaturation             string = "saturation"
	IOTypeScale                  string = "scale"
	IOTypeSignalStrength         string = "signalstrength"
	IOTypeSmokeDetector          string = "smokedetector"
	IOTypeSoundDetector          string = "sounddetector"
	IOTypeTemperature            string = "temperature"
	IOTypeUltraviolet            string = "ultraviolet"
	IOTypeVibrationDetector      string = "vibrationdetector"
	IOTypeValue                  string = "value" // generic value
	IOTypeVoltage                string = "voltage"
	IOTypeVolume                 string = "volume"
	IOTypeWaterLevel             string = "waterlevel"
	IOTypeWindHeading            string = "windheading"
	IOTypeWindSpeed              string = "windspeed"
)

// IOTypeMap defines data type and unit for an IOType
// Todo: option to download from file
var IOTypeMap = map[string]struct {
	DataType    DataType
	DefaultUnit Unit
	UnitValues  string
}{
	IOTypeUnknown:                {},
	IOTypeAcceleration:           {DataType: DataTypeNumber, DefaultUnit: "m/s2"},
	IOTypeAirQuality:             {DataType: DataTypeNumber, DefaultUnit: ""},
	IOTypeAlarm:                  {DataType: DataTypeString},
	IOTypeAtmosphericPressure:    {DataType: DataTypeNumber, DefaultUnit: UnitMillibar, UnitValues: UnitValuesAtmosphericPressure},
	IOTypeBattery:                {DataType: DataTypeNumber, DefaultUnit: UnitPercent},
	IOTypeCarbonDioxideLevel:     {DataType: DataTypeNumber, DefaultUnit: UnitPartsPerMillion},
	IOTypeCarbonMonoxideDetector: {DataType: DataTypeBool},
	IOTypeCarbonMonoxideLevel:    {DataType: DataTypeNumber, DefaultUnit: UnitPartsPerMillion},
	IOTypeChannel:                {DataType: DataTypeNumber},
	IOTypeColor:                  {DataType: DataTypeString},
	IOTypeColorTemperature:       {DataType: DataTypeNumber, DefaultUnit: UnitKelvin},
	IOTypeConnections:            {DataType: DataTypeNumber},
	IOTypeContact:                {DataType: DataTypeBool},
	IOTypeCPULevel:               {DataType: DataTypeNumber, DefaultUnit: UnitPercent},
	IOTypeCommand:                {DataType: DataTypeString},
	IOTypeDewpoint:               {DataType: DataTypeNumber, DefaultUnit: UnitCelcius, UnitValues: UnitValuesTemperature},
	IOTypeDimmer:                 {DataType: DataTypeNumber, DefaultUnit: UnitPercent},
	IOTypeDoorWindowSensor:       {DataType: DataTypeBool},
	IOTypeElectricCurrent:        {DataType: DataTypeBool, DefaultUnit: UnitAmp},
	IOTypeElectricEnergy:         {DataType: DataTypeNumber, DefaultUnit: UnitKWH},
	IOTypeElectricPower:          {DataType: DataTypeNumber, DefaultUnit: UnitWatt},
	IOTypeErrors:                 {DataType: DataTypeNumber},
	IOTypeHeatIndex:              {DataType: DataTypeNumber, DefaultUnit: UnitCelcius, UnitValues: "C, F"},
	IOTypeHue:                    {DataType: DataTypeString},
	IOTypeHumidex:                {DataType: DataTypeNumber, DefaultUnit: UnitCelcius, UnitValues: "C, F"},
	IOTypeHumidity:               {DataType: DataTypeNumber, DefaultUnit: UnitPercent},
	IOTypeImage:                  {DataType: DataTypeBytes, DefaultUnit: "", UnitValues: UnitValuesImage},
	IOTypeLatency:                {DataType: DataTypeNumber, DefaultUnit: UnitSecond},
	IOTypeLevel:                  {DataType: DataTypeNumber, DefaultUnit: UnitPercent},
	IOTypeLocation:               {DataType: DataTypeString},
	IOTypeLock:                   {DataType: DataTypeString},
	IOTypeLuminance:              {DataType: DataTypeNumber, DefaultUnit: UnitLux},
	IOTypeMotion:                 {DataType: DataTypeBool},
	IOTypeMute:                   {DataType: DataTypeBool},
	IOTypeOnOffSwitch:            {DataType: DataTypeBool},
	IOTypePlay:                   {DataType: DataTypeBool},
	IOTypePushButton:             {DataType: DataTypeNumber},
	IOTypeSaturation:             {DataType: DataTypeString},
	IOTypeScale:                  {DataType: DataTypeNumber, DefaultUnit: UnitKG, UnitValues: UnitValuesWeight},
	IOTypeSignalStrength:         {DataType: DataTypeNumber, DefaultUnit: "dBm"},
	IOTypeSmokeDetector:          {DataType: DataTypeBool},
	IOTypeSoundDetector:          {DataType: DataTypeBool},
	IOTypeTemperature:            {DataType: DataTypeNumber},
	IOTypeUltraviolet:            {DataType: DataTypeNumber},
	IOTypeVibrationDetector:      {DataType: DataTypeNumber},
	IOTypeValue:                  {DataType: DataTypeNumber},
	IOTypeVoltage:                {DataType: DataTypeNumber, DefaultUnit: UnitVolt},
	IOTypeVolume:                 {DataType: DataTypeNumber, DefaultUnit: UnitPercent},
	IOTypeWaterLevel:             {DataType: DataTypeNumber, DefaultUnit: UnitMeter, UnitValues: UnitValuesLength},
	IOTypeWindHeading:            {DataType: DataTypeNumber, DefaultUnit: UnitPercent},
	IOTypeWindSpeed:              {DataType: DataTypeNumber, DefaultUnit: UnitSpeed, UnitValues: UnitValuesSpeed},
}
