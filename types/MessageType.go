// Package types with IoTDomain message types used in message addressing
package types

// MessageType used in message addressing
type MessageType string

// Available message types from the standard
const (
	MessageTypeConfigure       = "$configure" // node configuration, payload is NodeConfigureMessage
	MessageTypeCreate          = "$create"    // create node command
	MessageTypeDelete          = "$delete"    // delete node command
	MessageTypeEvent           = "$event"     // node outputs event, payload is EventMessage
	MessageTypeForecast        = "$forecast"  // output forecast, payload is HistoryMessage
	MessageTypeHistory         = "$history"   // output history, payload is HistoryMessage
	MessageTypeIdentity        = "$identity"  // publisher identity
	MessageTypeInputDiscovery  = "$input"     // input discovery, payload is InOutput object
	MessageTypeLatest          = "$latest"    // latest output, payload is latest message
	MessageTypeLWT             = "$lwt"       // publisher last will and testament, if supported
	MessageTypeNodeDiscovery   = "$node"      // node discovery, payload is Node object
	MessageTypeNodeAlias       = "$alias"     // set node alias, payload is NodeAliasMessage
	MessageTypeOutputDiscovery = "$output"    // output discovery, payload output definition
	MessageTypeSet             = "$set"       // command to set input value, payload is input value
	MessageTypeUpgrade         = "$upgrade"   // perform firmware upgrade, payload is UpgradeMessage
	MessageTypeRaw             = "$raw"       // raw output value
	// LocaldomainID for local-only domains (eg, no sharing outside this domain)
	LocalDomainID = "local" // local area domain
	TestDomainID  = "test"  // Domain to use in testing
)
