// Package iotc with IoTConnect message pulication type
package iotc

import "encoding/json"

// Publication struct to  encapsulate messages with a signature
type Publication struct {
	// Message json.RawMessage `json:"message"`
	Message   json.RawMessage `json:"message"`
	Signature string          `json:"signature"`
}
