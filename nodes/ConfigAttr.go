// Package nodes with configuration map for nodes, inputs or outputs
package nodes

// ConfigAttrMap for use in node and in/output configuration
type ConfigAttrMap map[string]ConfigAttr

// ConfigAttr describing the configuration of the device/service or sensor
type ConfigAttr struct {
	DataType    DataType `json:"datatype,omitempty"`    // Data type of the attribute. [integer, float, boolean, string, bytes, enum, ...]
	Default     string   `json:"default,omitempty"`     // Default value
	Description string   `json:"description,omitempty"` // Description of the attribute
	Enum        []string `json:"enum,omitempty"`        // Possible valid enum values
	ID          string   `json:"id,omitempty"`          // Unique ID of this config
	Max         float64  `json:"max,omitempty"`         // Max value for numbers
	Min         float64  `json:"min,omitempty"`         // Min value for numbers
	Secret      bool     `json:"secret,omitempty"`      // the attribute value was set encrypted. Don't publish the value.
	Value       string   `json:"value,omitempty"`       // Current value of the attribute. Could be a string or map
}

// DataType of configuration values.
type DataType string

// Various data types
const (
	// DataTypeBool value is true/false, 1/0, on/off
	DataTypeBool DataType = "boolean"
	// DataTypeBytes value is encoded byte array
	DataTypeBytes DataType = "bytes"
	// DataTypeDate ISO8601 date YYYY-MM-DDTHH:MM:SS.mmmZ
	DataTypeDate DataType = "date"
	// DataTypeEnum value is one of a predefined set of string values, published in the 'enum info field'
	DataTypeEnum DataType = "enum"
	// DataTypeInt value is an integer number
	DataTypeInt DataType = "int"
	// value is a float number
	DataTypeNumber DataType = "number"
	// a secret string that is not published
	DataTypeSecret DataType = "secret"
	DataTypeString DataType = "string"
	// 3D vector (x, y, z) or (lat, lon, 0)
	DataTypeVector DataType = "vector"
	// value is a json object
	DataTypeJSON DataType = "json"
)

// NewConfigAttr instance for holding node configuration
func NewConfigAttr(id string, dataType DataType, description string, value string) *ConfigAttr {
	config := ConfigAttr{
		ID:          id,
		DataType:    dataType,
		Description: description,
		Value:       value,
	}
	return &config
}
