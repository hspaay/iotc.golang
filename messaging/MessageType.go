// Package messaging with message types used in message addressing
package messaging

// MessageType used in message addressing
type MessageType string

// Available message types from the standard
const (
	MessageTypeConfigure       = "$configure" // node configuration, payload is ConfigureMessage
	MessageTypeCreate          = "$create"    // create node command
	MessageTypeDelete          = "$delete"    // delete node command
	MessageTypeEvent           = "$event"     // node outputs event, payload is EventMessage
	MessageTypeForecast        = "$forecast"  // output forecast, payload is HistoryMessage
	MessageTypeHistory         = "$history"   // output history, payload is HistoryMessage
	MessageTypeInputDiscovery  = "$input"     // input discovery, payload is InOutput object
	MessageTypeLatest          = "$latest"    // latest output, payload is latest message
	MessageTypeNodeDiscovery   = "$node"      // node discovery, payload is Node object
	MessageTypeOutputDiscovery = "$output"    // output discovery, payload output definition
	MessageTypeSet             = "$set"       // control input command, payload is input value
	MessageTypeUpgrade         = "$upgrade"   // perform firmware upgrade, payload is UpgradeMessage
	MessageTypeValue           = "$value"     // raw output value
	// LocalZone ID for local-only zones (eg, no sharing outside this zone)
	LocalZoneID = "local"
	// PublisherNodeID to use when none is provided
	// PublisherNodeID = "$publisher" // reserved node ID for publishers
)
