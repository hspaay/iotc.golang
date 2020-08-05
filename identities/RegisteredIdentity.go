package identities

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/iotdomain/iotdomain-go/lib"
	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/sirupsen/logrus"
)

// valid for 1 year
const validDuration = time.Hour * 24 * 365

// IdentityFileSuffix to append to name of the file containing saved identity
const IdentityFileSuffix = "-identity.json"

// RegisteredIdentity for managing the publisher's full identity
type RegisteredIdentity struct {
	filename     string // identity filename under which it is saved
	domain       string // domain of the publisher creating this identity
	publisherID  string
	fullIdentity *types.PublisherFullIdentity
	dssPubKey    *ecdsa.PublicKey  // DSS pub key for verification (secure zones only)
	privateKey   *ecdsa.PrivateKey // private key from the new identity
	updated      bool              // flag, this identity has been updated and needs to be published/saved
}

// GetAddress returns the identity's publication address
func (regIdentity *RegisteredIdentity) GetAddress() string {
	return regIdentity.fullIdentity.Address
}

// GetPublicKey returns the identity's public key
// func (regIdentity *RegisteredIdentity) GetPublicKey() *ecdsa.PublicKey {
// 	return &regIdentity.privateKey.PublicKey
// }

// GetPrivateKey returns the identity's private key
func (regIdentity *RegisteredIdentity) GetPrivateKey() *ecdsa.PrivateKey {
	return regIdentity.privateKey
}

// GetFullIdentity returns the full identity with private key
func (regIdentity *RegisteredIdentity) GetFullIdentity() (fullIdentity *types.PublisherFullIdentity, privKey *ecdsa.PrivateKey) {
	return regIdentity.fullIdentity, regIdentity.privateKey
}

// LoadIdentity loads the publisher identity and private key from json file and
// verifies its content. See also VerifyIdentity for the criteria.
//
// Returns the identity with corresponding ECDSA private key, or nil if no identity is found
// If anything goes wrong, err will contain the error and nil identity is returned
// Use SaveIdentity to save updates to the identity
func (regIdentity *RegisteredIdentity) LoadIdentity(jsonFilename string) (
	fullIdentity *types.PublisherFullIdentity, privKey *ecdsa.PrivateKey, err error) {

	// load the identity
	identityJSON, err := ioutil.ReadFile(jsonFilename)
	if err != nil {
		return nil, nil, err
	}
	fullIdentity = &types.PublisherFullIdentity{}
	err = json.Unmarshal(identityJSON, fullIdentity)
	if err == nil {
		privKey = messaging.PrivateKeyFromPem(fullIdentity.PrivateKey)
	}
	if err == nil {
		// must match domain and publisher
		// We don't know the DSS signing key at this point
		err = VerifyFullIdentity(fullIdentity, regIdentity.domain, regIdentity.publisherID, nil)
	}
	// finaly, replace the identity with the loaded identity
	if err == nil {
		regIdentity.fullIdentity = fullIdentity
		regIdentity.privateKey = privKey
	}
	return fullIdentity, privKey, err
}

// SetDssKey sets the DSS public key. This is needed to allow the DSS to update the
// registered identity. Without it, any updates are refused. Intended to be set by
// the publisher when a verified DSS identity is received.
func (regIdentity *RegisteredIdentity) SetDssKey(dssSigningKey *ecdsa.PublicKey) {
	regIdentity.dssPubKey = dssSigningKey
}

// UpdateIdentity verifies and sets a new registered identity and saves it to the
// identity file.
func (regIdentity *RegisteredIdentity) UpdateIdentity(fullIdentity *types.PublisherFullIdentity) {

	err := VerifyFullIdentity(fullIdentity, regIdentity.domain, regIdentity.publisherID, regIdentity.dssPubKey)
	if err != nil {
		logrus.Errorf("UpdateIdentity: verification failed. Identity not updated.")
		return
	}
	privKey := messaging.PrivateKeyFromPem(fullIdentity.PrivateKey)
	regIdentity.privateKey = privKey
	regIdentity.fullIdentity = fullIdentity
	regIdentity.updated = true
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

// SaveIdentity save the full identity of the publisher to the given json filename.
// The identity is saved as a json file.
// see also https://stackoverflow.com/questions/21322182/how-to-store-ecdsa-private-key-in-go
func SaveIdentity(jsonFilename string, identity *types.PublisherFullIdentity) error {
	// save the identity as JSON. Remove the existing file first as they are read-only
	identityJSON, _ := json.MarshalIndent(identity, " ", " ")
	os.Remove(jsonFilename)
	err := ioutil.WriteFile(jsonFilename, identityJSON, 0400)
	if err != nil {
		return lib.MakeErrorf("SaveIdentity: Unable to save the publisher's identity at %s: %s", jsonFilename, err)
	}
	return err
}

// VerifyFullIdentity verifies the given full identity
// If the publisher joined with the DSS domain then a dssSigningKey is known and
// the identity MUST be signed by thep rovided DSS.
//
// verification  criteria:
//  - identity and keys were found, and
//  - the loaded identity is matches the domain/publisher of the publisher, and
//  - the loaded identity has valid keys, and
//  - the identity is not expired
//  - the identity signature matches the public identity
// If any of these conditions are not met then a new self-signed identity is created. When in a
// secured domain, the publisher must be re-added to the domain as the issuer is not the DSS.
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

// NewRegisteredIdentity creates a new registered identity
// Use LoadIdentity to load a previously saved identity
func NewRegisteredIdentity(domain string, publisherID string) (regIdent *RegisteredIdentity) {

	fullIdentity, privKey := CreateIdentity(domain, publisherID)

	regIdent = &RegisteredIdentity{
		domain:       domain,
		fullIdentity: fullIdentity,
		privateKey:   privKey,
		publisherID:  publisherID,
		updated:      true,
	}
	return regIdent
}
