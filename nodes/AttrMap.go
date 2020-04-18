package nodes

// AttrMap for use in node attributes and node status attributes
type AttrMap map[string]string

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
