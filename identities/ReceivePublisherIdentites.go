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

// ReceiveDSSIdentity discoveres the identity of the DSS, domain security service.
// This only applies to secured domains where the DSS issues a signed identity to
// each publisher. The identity is signed by the DSS and can be verified with DSS public key.
// Without a DSS, all publishers are unverified.
// func (rxIdentity *ReceiveDomainPublisherIdentities) ReceiveDSSIdentity(address string, rawMessage string) error {
// 	var newIdentity types.PublisherIdentityMessage

// 	logrus.Infof("ReceiveDSSIdentity: %s", address)

// 	// DSS identities aren't encrypted and their verification is based on either
// 	// message bus ACL's or a signed certificate from lets encrypt.
// 	_, err := messaging.VerifySenderJWSSignature(rawMessage, &newIdentity, nil)

// 	if err != nil {
// 		return lib.MakeErrorf("ReceiveDSSIdentity: Identity message from %s. Error %s'. Message discarded.", address, err)
// 	}

// 	// TODO: CA support. For now assume address protection is used so this is trusted.

// 	// dssSigningPem := dssIdentity.Identity.PublicKeySigning
// 	// dssSigningKey := messaging.PublicKeyFromPem(dssSigningPem)
// 	// publisher.dssSigningKey = dssSigningKey
// 	rxIdentity.domainIdentities.AddIdentity(&newIdentity)
// 	return nil
// }

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

	// Publishers self sign their JWS. Verify now we know the publisher.
	pubKey := messaging.PublicKeyFromPem(newIdentity.PublicKey)
	_, err = messaging.VerifyJWSMessage(rawMessage, pubKey)
	if err != nil {
		return lib.MakeErrorf("ReceiveDomainIdentity: Publisher message not self-signed '%s'. Identity ignored", address)
	}

	// Determine the key to verify the identity with
	issuerKey := pubKey
	if newIdentity.IssuerID == newIdentity.PublisherID {
		// self signed identity
		issuerKey = messaging.PublicKeyFromPem(newIdentity.PublicKey)
		err = VerifyPublisherIdentity(address, &newIdentity, issuerKey)
	} else if newIdentity.IssuerID == types.DSSPublisherID {
		// DSS signed identity. DSS Must be known.
		issuerAddress := newIdentity.Domain + "/" + newIdentity.IssuerID
		issuerKey = rxIdentity.domainIdentities.GetPublisherKey(issuerAddress)
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
