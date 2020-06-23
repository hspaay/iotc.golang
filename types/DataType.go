// Package types with IoTDomain data types used in input and output messages
package types

// DataType of configuration, input and ouput values.
type DataType string

// Available data types
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
