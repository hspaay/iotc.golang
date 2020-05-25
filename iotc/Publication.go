// Package iotc with IoTConnect message pulication type
package iotc

// Publication struct to  encapsulate messages with a signature
type Publication struct {
	// Dont use RawMessage as marshalling removes spaces and newlines of the message
	// Message json.RawMessage `json:"message"`
	Message   string `json:"message"` // JSON text message or base64 encoded raw message
	Signature string `json:"signature"`
}
