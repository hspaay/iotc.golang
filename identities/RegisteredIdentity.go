package identities

import (
	"crypto/ecdsa"
	"fmt"
	"strings"
	"time"

	"github.com/iotdomain/iotdomain-go/lib"
	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/iotdomain/iotdomain-go/persist"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/sirupsen/logrus"
)

// valid for 1 year
const validDuration = time.Hour * 24 * 365

// RegisteredIdentity for managing the publisher's full identity
type RegisteredIdentity struct {
	domain        string // domain of the publisher creating this identity
	publisherID   string
	FullIdentity  *types.PublisherFullIdentity
	privateKey    *ecdsa.PrivateKey
	messageSigner *messaging.MessageSigner
}

// HandleIdentityUpdate handles the set command for an update to this publisher identity.
// The message must be encrypted and signed by the DSS or it will be discarded.
func (regIdentity *RegisteredIdentity) HandleIdentityUpdate(address string, message string) error {
	var fullIdentity types.PublisherFullIdentity

	isEncrypted, isSigned, err := regIdentity.messageSigner.DecodeMessage(message, &fullIdentity)

	if err != nil {
		return lib.MakeErrorf("HandleIdentityUpdate: Message to %s. Error %s'. Message discarded.", address, err)
	} else if !isEncrypted {
		return lib.MakeErrorf("HandleIdentityUpdate: Identity update '%s' is not encrypted. Message discarded.", address)
	} else if !isSigned {
		return lib.MakeErrorf("HandleIdentityUpdate: Identity update '%s' is not signed. Message discarded.", address)
	}

	dssAddress := MakePublisherIdentityAddress(regIdentity.domain, types.DSSPublisherID)
	if fullIdentity.Sender != dssAddress {
		return lib.MakeErrorf("HandleIdentityUpdate: Sender is %s instead of the DSS %s. Identity update discarded.",
			fullIdentity.Sender, dssAddress)
	}

	privKey := messaging.PrivateKeyFromPem(fullIdentity.PrivateKey)
	regIdentity.privateKey = privKey
	regIdentity.FullIdentity = &fullIdentity
	return nil
}

// PublishIdentity publishes this identity
func (regIdentity *RegisteredIdentity) PublishIdentity() {
	publicIdentity := &regIdentity.FullIdentity.PublisherIdentityMessage
	logrus.Infof("PublishIdentity: publish identity: %s", publicIdentity.Address)
	regIdentity.messageSigner.PublishObject(publicIdentity.Address, true, publicIdentity, nil)
}

// // SetIdentity sets the registered identity
// func (regIdentity *RegisteredIdentity) SetIdentity(fullIdent *types.PublisherFullIdentity) {
// 	panic("not implemented")
// }

// Start listening for updates to the registered identity
// Intended to receive new keys from the DSS
func (regIdentity *RegisteredIdentity) Start() {
	addr := MakePublisherIdentityAddress(
		regIdentity.domain, regIdentity.publisherID)
	regIdentity.messageSigner.Subscribe(addr, regIdentity.HandleIdentityUpdate)
}

// Stop listening
func (regIdentity *RegisteredIdentity) Stop() {
	addr := MakePublisherIdentityAddress(
		regIdentity.FullIdentity.Domain, regIdentity.FullIdentity.PublisherID)
	regIdentity.messageSigner.Unsubscribe(addr, regIdentity.HandleIdentityUpdate)
}

// CreateIdentity creates and self-sign a new identity for the publisher
// This creates a base64encoded signature of the public identity using the given
// private key.
// The validity is 1 year.
func CreateIdentity(domain string, publisherID string) (
	fullIdentity *types.PublisherFullIdentity, signingPrivKey *ecdsa.PrivateKey) {
	// Create a new one and sign it.
	timestampStr := time.Now().Format(types.TimeFormat)
	validUntil := time.Now().Add(validDuration)
	validUntilStr := validUntil.Format(types.TimeFormat)

	// generate private/public key for signing and store the public key in the publisher identity in PEM format
	identityPrivKey := messaging.CreateAsymKeys()

	identityPubPem := messaging.PublicKeyToPem(&identityPrivKey.PublicKey)
	identityPrivPem := messaging.PrivateKeyToPem(identityPrivKey)
	addr := MakePublisherIdentityAddress(domain, publisherID)

	// self signed identity
	publicIdentity := types.PublisherIdentityMessage{
		Address:           addr,
		IdentitySignature: "",
		Domain:            domain,
		IssuerName:        publisherID, // self issued, will be replaced by DSS
		Location:          "local",
		Organization:      "", // todo: get from messenger configuration
		// PublicKeyCrypto:  pubCryptoStr,
		PublicKey:   identityPubPem,
		PublisherID: publisherID,
		Timestamp:   timestampStr,
		ValidUntil:  validUntilStr,
	}
	// self signed identity.
	identitySignature := messaging.CreateIdentitySignature(&publicIdentity, identityPrivKey)
	publicIdentity.IdentitySignature = identitySignature

	fullIdentity = &types.PublisherFullIdentity{
		PublisherIdentityMessage: publicIdentity,
		PrivateKey:               identityPrivPem,
	}
	return fullIdentity, identityPrivKey
}

// IsIdentityExpired tests if the given identity is expired
func IsIdentityExpired(identity *types.PublisherIdentityMessage) bool {
	timestampStr := time.Now().Format(types.TimeFormat)
	nowIsGreater := strings.Compare(timestampStr, identity.ValidUntil)
	return (nowIsGreater > 0)
}

// VerifyIdentity verifies the given identity is correctly signed, matches the given
// domain and publisher, and is not expired.
func VerifyIdentity(ident *types.PublisherFullIdentity, domain string,
	publisherID string, dssSigningKey *ecdsa.PublicKey) error {

	// sanity check in case the file was edited
	addr := MakePublisherIdentityAddress(domain, publisherID)

	if ident.Address != addr ||
		ident.Domain != domain ||
		ident.PublisherID != publisherID ||
		ident.PublicKey == "" ||
		ident.IdentitySignature == "" ||
		ident.PrivateKey == "" {
		err := lib.MakeErrorf("Identity file for %s/%s is invalid.", domain, publisherID)
		return err
	}

	expired := IsIdentityExpired(&ident.PublisherIdentityMessage)
	if expired {
		err := lib.MakeErrorf("Identity is expired")
		return err
	}
	// identity signature must verify against its signer, eg using the DSS public key
	if dssSigningKey != nil {
		err := messaging.VerifyIdentitySignature(&ident.PublisherIdentityMessage, dssSigningKey)
		if err != nil {
			err := lib.MakeErrorf("Identity signature mismatch: %s", err)
			return err
		}
	}
	// public key in identity must be the PEM key that belongs to the private key
	identPrivateKey := messaging.PrivateKeyFromPem(ident.PrivateKey)
	publicPem := messaging.PublicKeyToPem(&identPrivateKey.PublicKey)
	if publicPem != ident.PublicKey {
		err := lib.MakeErrorf("Public key in signed identity doesn't belong to the identity private key")
		return err
	}

	return nil
}

// MakePublisherIdentityAddress generates the address of a publisher:
//   domain/publisherID/$identity
// Intended for lookup of nodes in the node list.
// domain of the domain the node lives in.
// publisherID of the publisher for this node, unique for the domain
func MakePublisherIdentityAddress(domain string, publisherID string) string {
	address := fmt.Sprintf("%s/%s/%s", domain, publisherID, types.MessageTypeIdentity)
	return address
}

// SetupPublisherIdentity loads the publisher identity and keys from file in the identityFolder.
// It returns the full identity with its private/public keypair.
//
// The identity is discarded and a new identity is created on any of the following conditions:
//  - no identity and keys are found, or
//  - the loaded identity is invalid due to a domain/publisher/address mismatch, or
//  - the loaded identity is missing its keys, or
//  - the identity is expired
//  - the identity DSS signature verification fails
// If any of these conditions are met in a secured domain then the publisher must
// be re-added to the domain.
func SetupPublisherIdentity(
	identityFolder string, domain string, publisherID string, dssSigningKey *ecdsa.PublicKey) (
	fullIdentity *types.PublisherFullIdentity, privKey *ecdsa.PrivateKey) {

	// If an identity is saved, load it
	fullIdentity, privKey, err := persist.LoadIdentity(identityFolder, publisherID)
	// must match domain and publisher
	if err == nil {
		err = VerifyIdentity(fullIdentity, domain, publisherID, dssSigningKey)
	}
	// Recreate identity if invalid
	if err != nil { // invalid identity or none exists, create a new one
		fullIdentity, privKey = CreateIdentity(domain, publisherID)
		persist.SaveIdentity(identityFolder, publisherID, fullIdentity)
	}

	return fullIdentity, privKey
}

// NewRegisteredIdentity creates a new instance of the registered identity management, including handling
// of updates from the DSS.
func NewRegisteredIdentity(domain string, publisherID string,
	privateKey *ecdsa.PrivateKey, signer *messaging.MessageSigner) *RegisteredIdentity {
	regIdent := &RegisteredIdentity{
		messageSigner: signer,
		domain:        domain,
		FullIdentity:  &types.PublisherFullIdentity{},
		privateKey:    privateKey,
		publisherID:   publisherID,
	}
	return regIdent
}
