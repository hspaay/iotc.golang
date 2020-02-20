// Package nodes with definition for nodes, inputs and outputs
package nodes

import "fmt"

// Node definition
type Node struct {
	Address           string        `yaml:"address"            json:"address"`                              // Node discovery address
	Attr              AttrMap       `yaml:"attr,omitempty"     json:"attr,omitempty"`                       // Node/service specific info attributes
	Config            ConfigAttrMap `yaml:"config,omitempty"   json:"config,omitempty"`                     // Node/service configuration.
	Identity          *Identity     `yaml:"identity,omitempty" json:"identity,omitempty"`                   // Identity if node is a publisher
	IdentitySignature string        `yaml:"identitySignature,omitempty" json:"identitySignature,omitempty"` // optional signature of the identity by the ZSAS
	Status            AttrMap       `yaml:"status,omitempty"   json:"status,omitempty"`                     // include status at time of discovery
	// Inputs      map[string]*InOutput `yaml:"inputs,omitempty"  json:"inputs,omitempty"`  // inputs by their type.instance
	// Outputs     map[string]*InOutput `yaml:"outputs,omitempty" json:"outputs,omitempty"` // outputs by their type.instance
	// historySize int                  `yaml:"-" json:"-"`                                 // size of history for inputs and outputs
	// repeatDelay int                  `yaml:"-" json:"-"`
	// modified    bool                 `yaml:"-" json:"-"` // Node/service attribute or config has been updated
}

// AttrMap for use in node attributes and node status attributes
type AttrMap map[string]interface{}

// ConfigAttrMap for use in node and in/output configuration
type ConfigAttrMap map[string]*ConfigAttr

// ConfigAttr describing the configuration of the device/service or sensor
type ConfigAttr struct {
	DataType    DataType `yaml:"datatype,omitempty" json:"datatype,omitempty"`       // Data type of the attribute. [integer, float, boolean, string, bytes, enum, ...]
	Default     string   `yaml:"default,omitempty" json:"default,omitempty"`         // Default value
	Description string   `yaml:"description,omitempty" json:"description,omitempty"` // Description of the attribute
	Enum        []string `yaml:"enum,omitempty" json:"enum,omitempty"`               // Possible valid enum values
	Max         float64  `yaml:"max,omitempty" json:"max,omitempty"`                 // Max value for numbers
	Min         float64  `yaml:"min,omitempty" json:"min,omitempty"`                 // Min value for numbers
	//Name        string   `yaml:"name,omitempty" json:"name,omitempty"`               // Friendly name
	Secret bool   `yaml:"secret,omitempty" json:"secret,omitempty"` // the attribute value was set encrypted. Don't publish the value.
	Value  string `yaml:"value,omitempty" json:"value,omitempty"`   // Current value of the attribute. Could be a string or map
	//ValueList   []interface{} `yaml:"valuelist,omitempty" json:"valuelist,omitempty"`             // Current value of the attribute. Could be a string or map
	//ValueList   []interface{} `yaml:"-" json:"-"`             // Current value of the attribute. Could be a string or map
}

// Identity contains the identity information of a publisher node
type Identity struct {
	Address          string `yaml:"address"   json:"address"`                   // discovery address of the publisher (zone/pub/\$publisher/\$node)
	Expires          string `yaml:"expires"   json:"expires"`                   // timestamp this identity expires
	Location         string `yaml:"location"  json:"location"`                  // city, province, country
	Organization     string `yaml:"organization"    json:"organization"`        // publishing organization
	PublicKeyCrypto  string `yaml:"publicKeyCrypto"    json:"publicKeyCrypto"`  // public key for encrypting messages to this publisher
	PublicKeySigning string `yaml:"publicKeySigning"   json:"publicKeySigning"` // public key for verifying signature of messages published by this publisher
	Publisher        string `yaml:"publisher"    json:"publisher"`              // publisher ID
	Timestamp        string `yaml:"timestamp"    json:"timestamp"`              // timestamp this identity was last renewed/verified
	URL              string `yaml:"url"      json:"url"`                        // Web URL related to the publisher identity, if applicable
	Zone             string `yaml:"zone"     json:"zone"`                       // Zone in which publisher lives
}

// NewNode instance
func NewNode(zoneID string, publisherID string, nodeID string) *Node {
	address := fmt.Sprintf("%s/%s/%s/$node", zoneID, publisherID, nodeID)
	return &Node{
		Address: address,
		Attr:    AttrMap{},
		Config:  ConfigAttrMap{},
		Status:  AttrMap{},
	}
}
