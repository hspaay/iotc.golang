// Package iotc with IoTConnect publisher message type definitions
package iotc

// DSSPublisherID contains the publisherID of a domain's security service
const DSSPublisherID = "dss"

// LWTStatusConnected when currently connectioned
const LWTStatusConnected = "connected"

// LWTStatusDisconnected when connection is purposefully ended
const LWTStatusDisconnected = "disconnected"

// LWTStatusLost when connection unexpectedly drops
const LWTStatusLost = "lost"

// PublisherIdentity public record
type PublisherIdentity struct {
	Certificate  string `json:"certificate,omitempty"` // optional x509 cert base64 encoded
	Domain       string `json:"domain"`                // IoT domain name for this publisher
	IssuerName   string `json:"issuerName"`            // Issuer of the identity, usually the ZCAS
	Location     string `json:"location,omitempty"`    // city, province, country
	Organization string `json:"organization"`          // publishing organization
	PublicKey    string `json:"publicKey"`             // public key for signing and and encryption
	PublisherID  string `json:"publisherId"`           // This publisher's ID for this zone
	Timestamp    string `json:"timestamp"`             // timestamp this identity was signed
	ValidUntil   string `json:"validUntil"`            // timestamp this identity expires
}

// PublisherIdentityMessage for publisher discovery
type PublisherIdentityMessage struct {
	Address           string            `json:"address"`   // publication address of this identity, eg domain/publisherId/\$identity
	Identity          PublisherIdentity `json:"identity"`  // public identity
	IdentitySignature string            `json:"signature"` // base64 encoded signature of this identity
	SignerName        string            `json:"signer"`    // name of the signer of this identity, eg 'DSS' or 'Lets Encrypt'
	Timestamp         string            `json:"timestamp"` // timestamp this message was created
}

// PublisherLWTMessage containing 'alive' status
type PublisherLWTMessage struct {
	Address string `json:"address"` // publication address of this message
	Status  string `json:"status"`  //  LWTStatusXxx
}
