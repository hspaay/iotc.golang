// Package iotc with IoTConnect node message type definitions
package iotc

// PublisherNodeID to use when node is a publisher
const PublisherNodeID = "$publisher" // reserved node ID for publishers

// Predefined node attribute names that describe the node.
// When they are configurable they also appear in Node Config section.
// See also NodeStatusXxx below for attributes that describe the state of the node
const (
	NodeAttrAddress         NodeAttr = "address"         // the node's internal address. Can be used as the nodeID
	NodeAttrAlias           NodeAttr = "alias"           // node alias for publishing inputs and outputs
	NodeAttrColor           NodeAttr = "color"           // color in hex notation
	NodeAttrDescription     NodeAttr = "description"     // device description
	NodeAttrDisabled        NodeAttr = "disabled"        // device or sensor is disabled
	NodeAttrFilename        NodeAttr = "filename"        // filename to write images or other values to
	NodeAttrGatewayAddress  NodeAttr = "gatewayAddress"  // the node gateway address
	NodeAttrHostname        NodeAttr = "hostname"        // network device hostname
	NodeAttrIotcVersion     NodeAttr = "iotcVersion"     // IotConnect version
	NodeAttrLocalIP         NodeAttr = "localIP"         // for IP nodes
	NodeAttrLatLon          NodeAttr = "latlon"          // latitude, longitude of the device for display on a map r/w
	NodeAttrLocationName    NodeAttr = "locationName"    // name of a location
	NodeAttrLoginName       NodeAttr = "loginName"       // login name to connect to the device. Value is not published
	NodeAttrMAC             NodeAttr = "mac"             // MAC address for IP nodes
	NodeAttrManufacturer    NodeAttr = "manufacturer"    // device manufacturer
	NodeAttrMax             NodeAttr = "max"             // maximum value of sensor or config
	NodeAttrMin             NodeAttr = "min"             // minimum value of sensor or config
	NodeAttrModel           NodeAttr = "model"           // device model
	NodeAttrName            NodeAttr = "name"            // name of device, sensor
	NodeAttrNetmask         NodeAttr = "netmask"         // IP network mask
	NodeAttrPassword        NodeAttr = "password"        // password to connect. Value is not published.
	NodeAttrPollInterval    NodeAttr = "pollInterval"    // polling interval in seconds
	NodeAttrPowerSource     NodeAttr = "powerSource"     // battery, usb, mains
	NodeAttrProduct         NodeAttr = "product"         // device product or model name
	NodeAttrPublicKey       NodeAttr = "publicKey"       // public key for encrypting sensitive configuration settings
	NodeAttrSoftwareVersion NodeAttr = "softwareVersion" // version of the software running the node
	NodeAttrSubnet          NodeAttr = "subnet"          // IP subnets configuration
)

// Various NodeStatus attributes
const (
	NodeStatusErrorCount    NodeStatus = "errorCount"    // nr of errors reported on this device
	NodeStatusHealth        NodeStatus = "health"        // health status of the device 0-100%
	NodeStatusLastError     NodeStatus = "lastError"     // most recent error message
	NodeStatusLastSeen      NodeStatus = "lastSeen"      // ISO time the device was last seen
	NodeStatusLatencyMSec   NodeStatus = "latencyMSec"   // duration connect to sensor in milliseconds
	NodeStatusNeighborCount NodeStatus = "neighborCount" // mesh network nr of neighbors
	NodeStatusNeighborIDs   NodeStatus = "neighborIDs"   // mesh network device neighbors ID list [id,id,...]
	NodeStatusRxCount       NodeStatus = "rxCount"       // Nr of messages received from device
	NodeStatusTxCount       NodeStatus = "txCount"       // Nr of messages send to device
	NodeStatusRunState      NodeStatus = "runState"      // Node runstate as per below
)

// Various Running States
const (
	NodeRunStateError        string = "error"        // Node needs servicing
	NodeRunStateDisconnected string = "disconnected" // Node has cleanly disconnected
	NodeRunStateFailed       string = "failed"       // Node failed to start
	NodeRunStateInitializing string = "initializing" // Node is initializing
	NodeRunStateReady        string = "ready"        // Node is ready for use
	NodeRunStateSleeping     string = "sleeping"     // Node has gone into sleep mode, often a battery powered devie
)

// NodeType identifying  the purpose of the node
// Based on the primary role of the device.
type NodeType string

// Various Types of Nodes
const (
	NodeTypeAlarm          NodeType = "alarm"          // an alarm emitter
	NodeTypeAVControl      NodeType = "avControl"      // Audio/Video controller
	NodeTypeAVReceiver     NodeType = "avReceiver"     // Node is a (not so) smart radio/receiver/amp (eg, denon)
	NodeTypeBeacon         NodeType = "beacon"         // device is a location beacon
	NodeTypeButton         NodeType = "button"         // device is a physical button device with one or more buttons
	NodeTypeAdapter        NodeType = "adapter"        // software adapter or service, eg virtual device
	NodeTypePhone          NodeType = "phone"          // device is a phone
	NodeTypeCamera         NodeType = "camera"         // Node with camera
	NodeTypeComputer       NodeType = "computer"       // General purpose computer
	NodeTypeDimmer         NodeType = "dimmer"         // light dimmer
	NodeTypeGateway        NodeType = "gateway"        // Node is a gateway for other nodes (onewire, zwave, etc)
	NodeTypeKeypad         NodeType = "keypad"         // Entry key pad
	NodeTypeLock           NodeType = "lock"           // Electronic door lock
	NodeTypeMultisensor    NodeType = "multisensor"    // Node with multiple sensors
	NodeTypePublisher      NodeType = "publisher"      // Node is a zwave or other network repeater
	NodeTypeNetRepeater    NodeType = "netRepeater"    // Node is a zwave or other network repeater
	NodeTypeNetRouter      NodeType = "netRouter"      // Node is a network router
	NodeTypeNetSwitch      NodeType = "netSwitch"      // Node is a network switch
	NodeTypeNetWifiAP      NodeType = "wifiAP"         // Node is a wifi access point
	NodeTypeOnOffSwitch    NodeType = "onOffSwitch"    // Node is a physical on/off switch
	NodeTypePowerMeter     NodeType = "powerMeter"     // Node is a power meter
	NodeTypeSensor         NodeType = "sensor"         // Node is a single sensor (volt,...)
	NodeTypeSmartlight     NodeType = "smartlight"     // Node is a smart light, eg philips hue
	NodeTypeThermometer    NodeType = "thermometer"    // Node is a temperature meter
	NodeTypeThermostat     NodeType = "thermostat"     // Node is a thermostat control unit
	NodeTypeTV             NodeType = "tv"             // Node is a (not so) smart TV
	NodeTypeUnknown        NodeType = "unknown"        // type not identified
	NodeTypeWallpaper      NodeType = "wallpaper"      // Node is a wallpaper montage of multiple images
	NodeTypeWaterValve     NodeType = "waterValve"     // Water valve control unit
	NodeTypeWeatherService NodeType = "weatherService" // Node is a service providing current and forecasted weather
	NodeTypeWeatherStation NodeType = "weatherStation" // Node is a weatherstation device
	NodeTypeWeighScale     NodeType = "weighScale"     // Node is an electronic weight scale
)

// NodeAttr with predefined names of node attributes and configuration
type NodeAttr string

// NodeAttrMap for storing node attributes
type NodeAttrMap map[NodeAttr]string

// NodeStatus various node status attributes
type NodeStatus string

// ConfigAttrMap for storing node configuration
type ConfigAttrMap map[string]ConfigAttr

// ConfigAttr describing the configuration of the device/service or sensor
type ConfigAttr struct {
	Datatype    DataType `json:"datatype,omitempty"`    // Data type of the attribute. [integer, float, boolean, string, bytes, enum, ...]
	Default     string   `json:"default,omitempty"`     // Default value
	Description string   `json:"description,omitempty"` // Description of the attribute
	Enum        []string `json:"enum,omitempty"`        // Possible valid enum values
	ID          NodeAttr `json:"id,omitempty"`          // Unique ID of this config
	Max         float64  `json:"max,omitempty"`         // Max value for numbers
	Min         float64  `json:"min,omitempty"`         // Min value for numbers
	Secret      bool     `json:"secret,omitempty"`      // the attribute value was set encrypted. Don't publish the value.
	Value       string   `json:"value,omitempty"`       // Current value of the attribute. Could be a string or map
}

// NodeConfigureMessage with values to update a node configuration
type NodeConfigureMessage struct {
	Address   string      `json:"address"` // zone/publisher/node/$configure
	Attr      NodeAttrMap `json:"attr"`    // configuration attributes
	Sender    string      `json:"sender"`  // sending node: zone/publisher/node
	Timestamp string      `json:"timestamp"`
}

// PublisherIdentity for nodes that are publishers
type PublisherIdentity struct {
	Address          string `json:"address"`          // discovery address of the publisher (zone/pub/\$publisher/\$node)
	Expires          string `json:"expires"`          // timestamp this identity expires
	Location         string `json:"location"`         // city, province, country
	Organization     string `json:"organization"`     // publishing organization
	PublicKeyCrypto  string `json:"publicKeyCrypto"`  // public key for encrypting messages to this publisher
	PublicKeySigning string `json:"publicKeySigning"` // public key for verifying signature of messages published by this publisher
	Publisher        string `json:"publisher"`        // publisher ID
	Timestamp        string `json:"timestamp"`        // timestamp this identity was last renewed/verified
	URL              string `json:"url"`              // Web URL related to the publisher identity, if applicable
	Zone             string `json:"zone"`             // Zone in which publisher lives
}

// NodeDiscoveryMessage definition published in node discovery
type NodeDiscoveryMessage struct {
	// ID                string                  `json:"id"`                          // Node's immutable ID
	Address           string                  `json:"address"`                     // Node discovery address
	Attr              NodeAttrMap             `json:"attr,omitempty"`              // Node/service info attributes
	Config            map[NodeAttr]ConfigAttr `json:"config,omitempty"`            // Node/service configuration.
	Identity          *PublisherIdentity      `json:"identity,omitempty"`          // Identity if node is a publisher
	IdentitySignature string                  `json:"identitySignature,omitempty"` // optional signature of the identity by the ZSAS
	NodeID            string                  `json:"nodeId"`                      // The node ID
	// PublisherID       string                  `json:"publisher"`                   // publisher ID
	Status map[NodeStatus]string `json:"status,omitempty"` // additional node status details
	Type   NodeType              `json:"type"`             // node type
	// Zone              string                  `json:"zone"`                        // Zone in which node lives
}
