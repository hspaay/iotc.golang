// Package identities with handling of registered identity update command
package identities

import (
	"github.com/iotdomain/iotdomain-go/lib"
	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/sirupsen/logrus"
)

// ReceiveDomainPublisherIdentities listens for publisher identities on the domain.
// The domain identities are used to verify the signature of messages from a publisher
// In secured domains the domain identity must be signed by the DSS.
type ReceiveDomainPublisherIdentities struct {
	domainIdentities *DomainPublisherIdentities
	messageSigner    *messaging.MessageSigner // subscription to command
	dssAddress       string                   // the DSS address for this domain
}

// Start listening for updates to the registered identity
// Intended to receive new keys from the DSS
func (rxIdentity *ReceiveDomainPublisherIdentities) Start() {
	// subscription address for all identities domain/publisherID/$identity
	addr := MakePublisherIdentityAddress("+", "+")
	rxIdentity.messageSigner.Subscribe(addr, rxIdentity.ReceiveDomainIdentity)
}

// Stop listening
func (rxIdentity *ReceiveDomainPublisherIdentities) Stop() {
	addr := MakePublisherIdentityAddress("+", "+")
	rxIdentity.messageSigner.Unsubscribe(addr, rxIdentity.ReceiveDomainIdentity)

}

// ReceiveDomainIdentity handles receiving published identities of the domain.
// This:
// - verifies if the sender signature is valid
// - verifies that the identity is signed by the DSS when in a secure domain
// - passes the update to the domain identity collection
func (rxIdentity *ReceiveDomainPublisherIdentities) ReceiveDomainIdentity(address string, rawMessage string) error {
	var newIdentity types.PublisherIdentityMessage

	// Handle the DSS publisher separately
	// anyDssIdentitySuffix := types.DSSPublisherID + "/" + types.MessageTypeIdentity
	// isDSS := (address == rxIdentity.dssAddress || strings.HasSuffix(address, anyDssIdentitySuffix))

	logrus.Infof("ReceiveDomainIdentity: %s", address)

	// decode the message and determine the sender.
	isSigned, err := messaging.VerifySenderJWSSignature(rawMessage, &newIdentity, nil)
	if !isSigned || err != nil {
		return lib.MakeErrorf("ReceiveDomainIdentity: Invalid identity message on '%s': %s", address, err)
	}

	// Determine the key to verify the identity with
	if newIdentity.IssuerID == newIdentity.PublisherID {
		// self signed identity
		issuerKey := messaging.PublicKeyFromPem(newIdentity.PublicKey)
		err = VerifyPublisherIdentity(address, &newIdentity, issuerKey)
	} else if newIdentity.IssuerID == types.DSSPublisherID {
		// DSS signed identity. DSS Must be known.
		issuerAddress := newIdentity.Domain + "/" + newIdentity.IssuerID
		issuerKey := rxIdentity.domainIdentities.GetPublisherKey(issuerAddress)
		err = VerifyPublisherIdentity(address, &newIdentity, issuerKey)
	} else {
		// TODO: assume a CA signed identity. Not yet supported
		err = lib.MakeErrorf("Unknown Issuer %s for domain %s", newIdentity.IssuerID, newIdentity.Domain)
	}
	if err != nil {
		return lib.MakeErrorf("ReceiveDomainIdentity: Publisher identity signature verification failed for %s", address)
	}

	rxIdentity.domainIdentities.AddIdentity(&newIdentity)
	return nil
}

// NewReceivePublisherIdentities listens for publisher identity updates of the domain
// Run Start() to start listening.
func NewReceivePublisherIdentities(domain string, domainIdentities *DomainPublisherIdentities,
	messageSigner *messaging.MessageSigner) *ReceiveDomainPublisherIdentities {

	rxIdent := &ReceiveDomainPublisherIdentities{
		dssAddress:       MakePublisherIdentityAddress(domain, types.DSSPublisherID),
		domainIdentities: domainIdentities,
		messageSigner:    messageSigner,
	}
	return rxIdent
}
