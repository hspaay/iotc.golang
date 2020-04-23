// Package messaging with message pulication type
package messaging

import "encoding/json"

// Publication struct to  encapsulate messages with a signature
type Publication struct {
	// Message json.RawMessage `json:"message"`
	Message   json.RawMessage `json:"message"`
	Signature string          `json:"signature"`
}
