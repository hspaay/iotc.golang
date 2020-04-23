// Package messaging with message types used in message addressing
package messaging

// MessageType used in message addressing
type MessageType string

// Available message types from the standard
const (
	CommandConfigure       = "$configure" // node configuration, payload is ConfigureMessage
	CommandCreate          = "$create"    // create node command
	CommandDelete          = "$delete"    // delete node command
	CommandEvent           = "$event"     // node outputs event, payload is EventMessage
	CommandForecast        = "$forecast"  // output forecast, payload is HistoryMessage
	CommandHistory         = "$history"   // output history, payload is HistoryMessage
	CommandInputDiscovery  = "$input"     // input discovery, payload is InOutput object
	CommandLatest          = "$latest"    // latest output, payload is latest message
	CommandNodeDiscovery   = "$node"      // node discovery, payload is Node object
	CommandOutputDiscovery = "$output"    // output discovery, payload output definition
	CommandSet             = "$set"       // control input command, payload is input value
	CommandUpgrade         = "$upgrade"   // perform firmware upgrade, payload is UpgradeMessage
	CommandValue           = "$value"     // raw output value
	// LocalZone ID for local-only zones (eg, no sharing outside this zone)
	LocalZoneID = "local"
	// PublisherNodeID to use when none is provided
	// PublisherNodeID = "$publisher" // reserved node ID for publishers
)
