package nodes

import (
	"fmt"

	"github.com/hspaay/iotconnect.golang/standard"
)

// Node definition.
type Node struct {
	ID                string        `json:"id"`
	Address           string        `json:"address"`                     // Node discovery address
	Attr              AttrMap       `json:"attr,omitempty"`              // Node/service specific info attributes
	Config            ConfigAttrMap `json:"config,omitempty"`            // Node/service configuration.
	Identity          *Identity     `json:"identity,omitempty"`          // Identity if node is a publisher
	IdentitySignature string        `json:"identitySignature,omitempty"` // optional signature of the identity by the ZSAS
	PublisherID       string        `json:"publisher"`                   // publisher ID
	Status            AttrMap       `json:"status,omitempty"`            // include status at time of discovery
	Zone              string        `json:"zone"`                        // Zone in which node lives
	HistorySize       int           `json:"historySize"`                 // size of history for inputs and outputs, default automatically for 24 hours
	RepeatDelay       int           `json:"repeatDelay"`                 // delay in seconds before repeating the same value, default 1 hour
}

// Identity contains the identity information of a publisher node
type Identity struct {
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

// NodeType identifying  the purpose of the node
// Based on the primary role of the device.
type NodeType string

// PublisherNodeID to use when node is a publisher
const PublisherNodeID = "$publisher" // reserved node ID for publishers

// Various Types of Nodes
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

// Clone returns a copy of the node with new Attr, Config and Status maps
// Intended for updating the node in a concurrency safe manner in combination with UpdateNode()
// This does not perform a deep copy of the  maps. Any updates to the map must use new instances of the values
func (node *Node) Clone() *Node {
	newNode := *node

	newNode.Attr = make(AttrMap)
	for key, value := range node.Attr {
		newNode.Attr[key] = value
	}
	newNode.Config = make(ConfigAttrMap)
	for key, value := range node.Config {
		newNode.Config[key] = value
	}
	newNode.Status = make(AttrMap)
	for key, value := range node.Status {
		newNode.Status[key] = value
	}
	return &newNode
}

// GetAlias returns the alias or node ID if no alias is set
func (node *Node) GetAlias() (alias string, hasAlias bool) {
	hasAlias = false
	alias = node.ID
	aliasConfig, attrExists := node.Config[AttrNameAlias]
	if attrExists && aliasConfig.Value != "" {
		alias = aliasConfig.Value
	}
	return alias, hasAlias
}

// GetConfigValue returns the node configuration value
// This retuns the 'default' value if no value is set
func (node *Node) GetConfigValue(attrName string) (string, bool) {
	config, configExists := node.Config[attrName]
	if !configExists {
		return "", configExists
	}
	if config.Value == "" {
		return config.Default, configExists
	}
	return config.Value, configExists
}

// MakeNodeDiscoveryAddress for publishing
func MakeNodeDiscoveryAddress(zoneID string, publisherID string, nodeID string) string {
	address := fmt.Sprintf("%s/%s/%s/"+standard.CommandNodeDiscovery, zoneID, publisherID, nodeID)
	return address
}

// NewNode instance
func NewNode(zoneID string, publisherID string, nodeID string) *Node {
	address := MakeNodeDiscoveryAddress(zoneID, publisherID, nodeID)
	return &Node{
		Address:     address,
		Attr:        AttrMap{},
		Config:      ConfigAttrMap{},
		ID:          nodeID,
		PublisherID: publisherID,
		Status:      AttrMap{},
		Zone:        zoneID,
	}
}
