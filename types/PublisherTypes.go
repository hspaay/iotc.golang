// Package types with publisher message type definitions
package types

// DSSPublisherID defines the publisherID of a domain's security service
// the DSS is responsible for renewal of keys in a secured domain.
const DSSPublisherID = "$dss"

// PublisherRunState indicates the operating status of the publisher. Used in LWT.
type PublisherRunState string

// PublisherState values
const (
	PublisherRunStateConnected    PublisherRunState = "connected"    // Publisher is connected and working
	PublisherRunStateDisconnected PublisherRunState = "disconnected" // Publisher has cleanly disconnected
	PublisherRunStateFailed       PublisherRunState = "failed"       // Publisher failed to start
	PublisherRunStateInitializing PublisherRunState = "initializing" // Publisher is initializing
	PublisherRunStateLost         PublisherRunState = "lost"         // Publisher unexpectedly disconnected
)

// PublisherIdentityMessage contains the public identity of a publisher
type PublisherIdentityMessage struct {
	Address           string `json:"address"`               // publication address of this identity, eg domain/publisherId/\$identity
	Certificate       string `json:"certificate,omitempty"` // optional x509 cert base64 encoded
	Domain            string `json:"domain"`                // IoT domain name for this publisher
	IssuerID          string `json:"issuerId"`              // Issuer of the identity, the DSS, publisherId or CA
	Location          string `json:"location,omitempty"`    // city, province, country
	Organization      string `json:"organization"`          // publishing organization
	PublicKey         string `json:"publicKey"`             // public key in PEM format for signature verification and encryption
	PublisherID       string `json:"publisherId"`           // This publisher's ID for this domain
	ValidUntil        string `json:"validUntil"`            // timestamp this identity expires
	IdentitySignature string `json:"signature"`             // base64 encoded signature of this identity
	Timestamp         string `json:"timestamp"`             // timestamp this message was created
}

// PublisherFullIdentity containing the public identity, DSS signature and private key
// Also used by the DSS to renew a publisher's identity.
// This message MUST be encrypted and signed by the DSS
type PublisherFullIdentity struct {
	PublisherIdentityMessage
	PrivateKey string `json:"privateKey"` // private key for signing (PEM format)
	Sender     string `json:"sender"`     // sender of this update, usually the DSS
}

// PublisherStatusMessage containing 'alive' status, used in LWT
type PublisherStatusMessage struct {
	Address string            `json:"address"` // publication address of this message
	Status  PublisherRunState `json:"status"`
}
