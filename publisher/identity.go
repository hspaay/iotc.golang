// Package publisher with functions for managing the publisher's identity
package publisher

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/hspaay/iotc.golang/iotc"
	"github.com/hspaay/iotc.golang/messenger"
	"github.com/hspaay/iotc.golang/nodes"
	"github.com/hspaay/iotc.golang/persist"
)

// CreateIdentity creates a new identity for a domain publisher
// The validity is 1 year
func CreateIdentity(domain string, publisherID string) (identityMessage *iotc.PublisherIdentityMessage, privSigningKey *ecdsa.PrivateKey) {
	// No identity could be loaded, Create a new one and sign it.
	timestampStr := time.Now().Format(iotc.TimeFormat)
	validUntil := time.Now().Add(time.Hour * 24 * 365) // valid for 1 year
	validUntilStr := validUntil.Format(iotc.TimeFormat)

	// generate private/public key for signing and store the public key in the publisher identity in PEM format
	rng := rand.Reader
	curve := elliptic.P256()
	privKey, err := ecdsa.GenerateKey(curve, rng)
	if err != nil {
		panic("Unable to generate a private signing key. Can't continue without it.")
	}

	pubSigningStr := messenger.PublicKeyToPem(&privKey.PublicKey)

	addr := nodes.MakePublisherIdentityAddress(domain, publisherID)

	identity := iotc.PublisherIdentity{
		Domain:       domain,
		IssuerName:   publisherID, // self issued, will be replaced by ZCAS
		Location:     "local",
		Organization: "", // todo: get from messenger configuration
		// PublicKeyCrypto:  pubCryptoStr,
		PublicKey:   pubSigningStr,
		PublisherID: publisherID,
		Timestamp:   timestampStr,
		ValidUntil:  validUntilStr,
	}
	// self signed identity
	identitySignature := messenger.SignEncodeIdentity(&identity, privKey)
	identityMessage = &iotc.PublisherIdentityMessage{
		Address:           addr,
		Identity:          identity,
		IdentitySignature: identitySignature,
		SignerName:        publisherID,
		Timestamp:         timestampStr,
	}
	return identityMessage, privKey
}

// IsIdentityExpired tests if the given identity is expired
func IsIdentityExpired(identity *iotc.PublisherIdentity) bool {
	timestampStr := time.Now().Format(iotc.TimeFormat)
	nowIsGreater := strings.Compare(timestampStr, identity.ValidUntil)
	return (nowIsGreater > 0)
}

// LoadIdentity loads the publisher identity and private key from file in the given folder.
// The expected identity file is named <publisherID>-identity.json, the private key file is
// named <publisherID>-private.pem.
// Returns the identity with corresponding ECDSA private key, or nil if no identity is found
// If anything goes wrong, err will contain the error and nil identity is returned
func LoadIdentity(folder string, publisherID string) (identityMsg *iotc.PublisherIdentityMessage, privKey *ecdsa.PrivateKey, err error) {
	identityFile := fmt.Sprintf("%s/%s-identity.json", folder, publisherID)
	privFile := fmt.Sprintf("%s/%s-private.pem", folder, publisherID)

	// load the identity
	identityJSON, err := ioutil.ReadFile(identityFile)
	if err != nil {
		return nil, nil, err
	}
	identityMsg = &iotc.PublisherIdentityMessage{}
	err = json.Unmarshal(identityJSON, identityMsg)
	if err != nil {
		msg := fmt.Sprintf("Error unmarshalling identity file: %s", err)
		print(msg)
		return nil, nil, err
	}

	// load the private key pem file
	pemEncodedPriv, err := ioutil.ReadFile(privFile)
	if err != nil {
		return nil, nil, err
	}
	blockPriv, _ := pem.Decode(pemEncodedPriv)
	x509Encoded := blockPriv.Bytes
	privateKey, _ := x509.ParseECPrivateKey(x509Encoded)

	return identityMsg, privateKey, nil
}

// SaveIdentity save the identity message of the publisher and its keys in the given folder.
// The identity is saved as a json file. The keys are saved as <publisherId>-private.pem.
// see also https://stackoverflow.com/questions/21322182/how-to-store-ecdsa-private-key-in-go
func SaveIdentity(folder string, publisherID string,
	identity *iotc.PublisherIdentityMessage, privKey *ecdsa.PrivateKey) error {
	privFile := fmt.Sprintf("%s/%s-private.pem", folder, publisherID)
	identityFile := fmt.Sprintf("%s/%s-identity.json", folder, publisherID)

	// save the identity as JSON. Remove first as they are read-only
	identityJSON, _ := json.MarshalIndent(identity, " ", " ")
	os.Remove(identityFile)
	err := ioutil.WriteFile(identityFile, identityJSON, 0400)
	if err != nil {
		err := fmt.Errorf("SaveIdentity: Unable to save the publisher's identity at %s: %s", identityFile, err)
		return err
	}

	// save the private key pem file
	x509Encoded, _ := x509.MarshalECPrivateKey(privKey)
	pemEncodedPriv := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: x509Encoded})
	os.Remove(privFile)
	err = ioutil.WriteFile(privFile, pemEncodedPriv, 0400)
	if err != nil {
		err := fmt.Errorf("SaveIdentity: Unable to save the publisher's identity private key at %s: %s", privFile, err)
		panic(err)
	}

	return err
}

// SetupPublisherIdentity loads the publisher identity and keys from file in the identityFolder.
// If no identity and keys are found, a self signed identity is created.
// See SaveIdentity for info on how the identity is saved.
//
// identityFolder contains the folder with the identity files, use "" for default config folder (.config/iotc)
//   if you're paranoid, this can be on a USB key that is inserted on startup and removed once running.
// domain and publisherID are used to define the identity address
func SetupPublisherIdentity(identityFolder string, domain string, publisherID string) (identityMessage *iotc.PublisherIdentityMessage, privateKey *ecdsa.PrivateKey) {
	// var identity *iotc.PublisherIdentityMessage

	if identityFolder == "" {
		identityFolder = persist.DefaultConfigFolder
	}
	// If an identity is saved, load it
	identityMessage, privKey, err := LoadIdentity(identityFolder, publisherID)
	if err == nil {
		return identityMessage, privKey
	}

	identityMessage, privKey = CreateIdentity(domain, publisherID)

	SaveIdentity(identityFolder, publisherID, identityMessage, privKey)

	return identityMessage, privKey
}
