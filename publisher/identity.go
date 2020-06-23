// Package publisher with functions for managing the publisher's identity
package publisher

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/iotdomain/iotdomain-go/messenger"
	"github.com/iotdomain/iotdomain-go/nodes"
	"github.com/iotdomain/iotdomain-go/persist"
	"github.com/iotdomain/iotdomain-go/types"
)

// CreateIdentity creates a new identity for a domain publisher
// The validity is 1 year
func CreateIdentity(domain string, publisherID string) (
	fullIdentity *types.PublisherFullIdentity, privKey *ecdsa.PrivateKey) {
	// No identity could be loaded, Create a new one and sign it.
	timestampStr := time.Now().Format(types.TimeFormat)
	validUntil := time.Now().Add(time.Hour * 24 * 365) // valid for 1 year
	validUntilStr := validUntil.Format(types.TimeFormat)

	// generate private/public key for signing and store the public key in the publisher identity in PEM format
	rng := rand.Reader
	curve := elliptic.P256()
	privKey, err := ecdsa.GenerateKey(curve, rng)
	if err != nil {
		panic("Unable to generate a private signing key. Can't continue without it.")
	}

	pubSigningStr := messenger.PublicKeyToPem(&privKey.PublicKey)

	addr := nodes.MakePublisherIdentityAddress(domain, publisherID)

	publicIdentity := types.PublisherPublicIdentity{
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
	identitySignature := messenger.SignEncodeIdentity(&publicIdentity, privKey)
	fullIdentity = &types.PublisherFullIdentity{
		PublisherIdentityMessage: types.PublisherIdentityMessage{
			Address:           addr,
			Public:            publicIdentity,
			IdentitySignature: identitySignature,
			SignerName:        publisherID,
			Timestamp:         timestampStr,
		},
		PrivateKey: messenger.PrivateKeyToPem(privKey),
	}
	return fullIdentity, privKey
}

// IsIdentityExpired tests if the given identity is expired
func IsIdentityExpired(identity *types.PublisherPublicIdentity) bool {
	timestampStr := time.Now().Format(types.TimeFormat)
	nowIsGreater := strings.Compare(timestampStr, identity.ValidUntil)
	return (nowIsGreater > 0)
}

// LoadIdentity loads the publisher identity and private key from file in the given folder.
// The expected identity file is named <publisherID>-identity.json.
// Returns the identity with corresponding ECDSA private key, or nil if no identity is found
// If anything goes wrong, err will contain the error and nil identity is returned
func LoadIdentity(folder string, publisherID string) (fullIdentity *types.PublisherFullIdentity, privateKey *ecdsa.PrivateKey, err error) {
	identityFile := fmt.Sprintf("%s/%s-identity.json", folder, publisherID)

	// load the identity
	identityJSON, err := ioutil.ReadFile(identityFile)
	if err != nil {
		return nil, nil, err
	}
	fullIdentity = &types.PublisherFullIdentity{}
	err = json.Unmarshal(identityJSON, fullIdentity)
	if err != nil {
		msg := fmt.Sprintf("Error unmarshalling identity file: %s", err)
		print(msg)
		return nil, nil, err
	}
	// sanity check in case the file was edited
	addr := nodes.MakePublisherIdentityAddress(fullIdentity.Public.Domain, publisherID)
	if fullIdentity.Public.Domain == "" ||
		fullIdentity.Public.PublisherID != publisherID ||
		fullIdentity.Address != addr ||
		fullIdentity.Public.PublicKey == "" ||
		fullIdentity.PrivateKey == "" {
		msg := fmt.Sprintf("Identity file is inconsistent. Maybe it was edited")
		return nil, nil, errors.New(msg)
	}
	// TODO verify signature with public part
	privateKey = messenger.PrivateKeyFromPem(fullIdentity.PrivateKey)
	return fullIdentity, privateKey, nil
}

// SaveIdentity save the full identity of the publisher and its keys in the given folder.
// The identity is saved as a json file.
// see also https://stackoverflow.com/questions/21322182/how-to-store-ecdsa-private-key-in-go
func SaveIdentity(folder string, publisherID string, identity *types.PublisherFullIdentity) error {
	identityFile := fmt.Sprintf("%s/%s-identity.json", folder, publisherID)

	// save the identity as JSON. Remove first as they are read-only
	identityJSON, _ := json.MarshalIndent(identity, " ", " ")
	os.Remove(identityFile)
	err := ioutil.WriteFile(identityFile, identityJSON, 0400)
	if err != nil {
		err := fmt.Errorf("SaveIdentity: Unable to save the publisher's identity at %s: %s", identityFile, err)
		return err
	}
	return err
}

// SetupPublisherIdentity loads the publisher identity and keys from file in the identityFolder.
// If no identity and keys are found, a self signed identity is created. If the loaded identity is invalid,
// due to a domain/publisher/address mismatch, or its public key is missing, a new identity is also created.
// See SaveIdentity for info on how the identity is saved.
//
// identityFolder contains the folder with the identity files, use "" for default config folder (.config/iotc)
// domain and publisherID are used to define the identity address
func SetupPublisherIdentity(identityFolder string, domain string, publisherID string) (
	fullIdentity *types.PublisherFullIdentity, privKey *ecdsa.PrivateKey) {

	if identityFolder == "" {
		identityFolder = persist.DefaultConfigFolder
	}
	// If an identity is saved, load it
	fullIdentity, privKey, err := LoadIdentity(identityFolder, publisherID)
	identityAddress := nodes.MakePublisherIdentityAddress(domain, publisherID)

	// validity check on identity, recreate a new one if changed
	if err != nil ||
		fullIdentity.Public.Domain != domain ||
		fullIdentity.Public.PublisherID != publisherID ||
		fullIdentity.Address != identityAddress ||
		fullIdentity.Public.PublicKey == "" {
		// invalid identity or none exists, create a new one
		fullIdentity, privKey = CreateIdentity(domain, publisherID)
		SaveIdentity(identityFolder, publisherID, fullIdentity)
	} else {
		expired := IsIdentityExpired(&fullIdentity.Public)
		if expired {
			// assume the DSS will re-issue an updated identitiy
		}
	}

	return fullIdentity, privKey
}
