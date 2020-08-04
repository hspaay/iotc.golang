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
	domain       string // domain of the publisher creating this identity
	publisherID  string
	fullIdentity *types.PublisherFullIdentity
	dssPubKey    *ecdsa.PublicKey  // DSS pub key for verification (secure zones only)
	privateKey   *ecdsa.PrivateKey // private key from the new identity
}

// GetAddress returns the identity's publication address
func (regIdentity *RegisteredIdentity) GetAddress() string {
	return regIdentity.fullIdentity.Address
}

// GetIdentity returns the full identity with private key
func (regIdentity *RegisteredIdentity) GetIdentity() (fullIdentity *types.PublisherFullIdentity, privKey *ecdsa.PrivateKey) {
	return regIdentity.fullIdentity, regIdentity.privateKey
}

// UpdateIdentity verifies and sets a new registered identity
func (regIdentity *RegisteredIdentity) UpdateIdentity(fullIdentity *types.PublisherFullIdentity) {

	err := VerifyFullIdentity(fullIdentity, regIdentity.domain, regIdentity.publisherID, regIdentity.dssPubKey)
	if err != nil {
		logrus.Errorf("UpdateIdentity: verification failed. Identity not updated.")
		return
	}
	privKey := messaging.PrivateKeyFromPem(fullIdentity.PrivateKey)
	regIdentity.privateKey = privKey
	regIdentity.fullIdentity = fullIdentity
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
		IssuerID:          publisherID, // self issued, will be replaced by DSS
		Location:          "local",
		Organization:      "", // todo: get from messenger configuration
		// PublicKeyCrypto:  pubCryptoStr,
		PublicKey:   identityPubPem,
		PublisherID: publisherID,
		Timestamp:   timestampStr,
		ValidUntil:  validUntilStr,
	}
	// self signed identity.
	messaging.SignIdentity(&publicIdentity, identityPrivKey)

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

// MakePublisherIdentityAddress generates the address of a publisher:
//   domain/publisherID/$identity
// Intended for lookup of nodes in the node list.
// domain of the domain the node lives in.
// publisherID of the publisher for this node, unique for the domain
func MakePublisherIdentityAddress(domain string, publisherID string) string {
	address := fmt.Sprintf("%s/%s/%s", domain, publisherID, types.MessageTypeIdentity)
	return address
}

// PublishIdentity signs and publishes the public part of this identity using
// the given message signer.
func (regIdentity *RegisteredIdentity) PublishIdentity(signer *messaging.MessageSigner) {
	publicIdentity := &regIdentity.fullIdentity.PublisherIdentityMessage
	logrus.Infof("PublishIdentity: publish identity: %s", publicIdentity.Address)

	signer.PublishObject(publicIdentity.Address, true, publicIdentity, nil)
}

// SetDssKey sets the DSS public key. This is needed to allow the DSS to update the
// registered identity. Without it, any updates are refused. Intended to be set by
// the publisher when a verified DSS identity is received.
func (regIdentity *RegisteredIdentity) SetDssKey(dssSigningKey *ecdsa.PublicKey) {
	regIdentity.dssPubKey = dssSigningKey
}

// VerifyFullIdentity verifies the given full identity is correctly signed,
// matches the given domain and publisher, and is not expired.
// If the publisher joined with the DSS domain then a dssSigningKey is known and
// the identity MUST be signed by the DSS.
func VerifyFullIdentity(ident *types.PublisherFullIdentity, domain string,
	publisherID string, dssSigningKey *ecdsa.PublicKey) error {

	// the public identity must verify first
	err := VerifyPublisherIdentity(ident.Address, &ident.PublisherIdentityMessage, dssSigningKey)
	if err != nil {
		return err
	}

	// public key in identity must be the PEM key that belongs to the private key
	identPrivateKey := messaging.PrivateKeyFromPem(ident.PrivateKey)
	publicPem := messaging.PublicKeyToPem(&identPrivateKey.PublicKey)
	if publicPem != ident.PublicKey {
		return lib.MakeErrorf("VerifyFullIdentity: Public key in signed identity '%s' doesn't belong to the identity private key", ident.Address)
	}
	// identity is valid
	return nil
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
	identityFolder string, domain string, publisherID string,
	dssSigningKey *ecdsa.PublicKey) (
	fullIdentity *types.PublisherFullIdentity, privKey *ecdsa.PrivateKey) {

	// If an identity is saved, load it
	fullIdentity, privKey, err := persist.LoadIdentity(identityFolder, publisherID)
	// must match domain and publisher
	if err == nil {
		err = VerifyFullIdentity(fullIdentity, domain, publisherID, dssSigningKey)
	}
	// Recreate identity if invalid
	if err != nil { // invalid identity or none exists, create a new one
		fullIdentity, privKey = CreateIdentity(domain, publisherID)
		persist.SaveIdentity(identityFolder, publisherID, fullIdentity)
	}

	return fullIdentity, privKey
}

// NewRegisteredIdentity creates a new instance of the registered identity management,
// including handling of updates from the DSS.
// This first loads a previously saved identity if available and generates a new
// identity when no valid identity is known.
func NewRegisteredIdentity(identityFolder string, domain string, publisherID string,
) (regIdent *RegisteredIdentity, privKey *ecdsa.PrivateKey) {

	fullIdentity, privKey := SetupPublisherIdentity(
		identityFolder, domain, publisherID, nil)

	regIdent = &RegisteredIdentity{
		domain:       domain,
		fullIdentity: fullIdentity,
		privateKey:   privKey,
		publisherID:  publisherID,
	}
	return regIdent, privKey
}
