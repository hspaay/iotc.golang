// Package types with publisher message type definitions
package types

// DSSPublisherID contains the publisherID of a domain's security service
const DSSPublisherID = "$dss"

// LWTStatusConnected when currently connectioned
const LWTStatusConnected = "connected"

// LWTStatusDisconnected when connection is purposefully ended
const LWTStatusDisconnected = "disconnected"

// LWTStatusLost when connection unexpectedly drops
const LWTStatusLost = "lost"

// PublisherPublicIdentity public record
type PublisherPublicIdentity struct {
	Certificate  string `json:"certificate,omitempty"` // optional x509 cert base64 encoded
	Domain       string `json:"domain"`                // IoT domain name for this publisher
	IssuerName   string `json:"issuerName"`            // Issuer of the identity, usually the DSS
	Location     string `json:"location,omitempty"`    // city, province, country
	Organization string `json:"organization"`          // publishing organization
	PublicKey    string `json:"publicKey"`             // public key in PEM format for signature verification and encryption
	PublisherID  string `json:"publisherId"`           // This publisher's ID for this domain
	Timestamp    string `json:"timestamp"`             // timestamp this identity was signed
	ValidUntil   string `json:"validUntil"`            // timestamp this identity expires
}

// PublisherIdentityMessage for publisher discovery
type PublisherIdentityMessage struct {
	Address           string                  `json:"address"`   // publication address of this identity, eg domain/publisherId/\$identity
	Public            PublisherPublicIdentity `json:"public"`    // public identity
	IdentitySignature string                  `json:"signature"` // base64 encoded signature of this identity
	SignerName        string                  `json:"signer"`    // name of the signer of this identity, eg 'DSS' or 'Lets Encrypt'
	Timestamp         string                  `json:"timestamp"` // timestamp this message was created
}

// PublisherFullIdentity containing the public identity, DSS signature and private key
// Also used by the DSS to renew a publisher's identity.
// This message MUST be encrypted and signed by the DSS
type PublisherFullIdentity struct {
	PublisherIdentityMessage
	PrivateKey string `json:"privateKey"` // private key for signing (PEM format)
	Sender     string `json:"sender"`     // sender of this update, usually the DSS
}

// PublisherLWTMessage containing 'alive' status
type PublisherLWTMessage struct {
	Address string `json:"address"` // publication address of this message
	Status  string `json:"status"`  //  LWTStatusXxx
}
