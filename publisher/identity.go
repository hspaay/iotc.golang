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
	"time"

	"github.com/hspaay/iotc.golang/iotc"
	"github.com/hspaay/iotc.golang/messenger"
	"github.com/hspaay/iotc.golang/persist"
)

// LoadIdentity loads the publisher identity with public and private keys from pem files
// Returns the identity with corresponding ECDSA private and public key, or nil if no identity is found
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
		return identityMsg, nil, err
	}
	blockPriv, _ := pem.Decode(pemEncodedPriv)
	x509Encoded := blockPriv.Bytes
	privateKey, _ := x509.ParseECPrivateKey(x509Encoded)

	return identityMsg, privateKey, nil
}

// SaveIdentity save the identity of the publisher and its keys
// The identity is saved as json
// The keys are saved as folder/<publisherId>-private.pem
// The public key as it is embedded in the private key
// see also https://stackoverflow.com/questions/21322182/how-to-store-ecdsa-private-key-in-go
//
// folder is the location to save the key, preferably secured
// publisherID is used in the name of the key file
// identity is the identity message to save
// privKey is the identity's private key to save
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

// SetupPublisherIdentity loads the publisher identity and keys from file, or creates a new one
// if none yet exists.
// identityFolder contains the folder with the identity files, use "" for default config folder (.config/iotc)
//   if you're paranoid, this can be on a USB key that is inserted on startup and removed once running.
// domain for this identity
// publisherID for this identity
func (publisher *Publisher) SetupPublisherIdentity(identityFolder string, domain string, publisherID string) {
	var identity *iotc.PublisherIdentityMessage

	if identityFolder == "" {
		identityFolder = persist.DefaultConfigFolder
	}

	identity, privKey, err := LoadIdentity(identityFolder, publisherID)
	if err == nil {
		publisher.identity = identity
		publisher.privateKeySigning = privKey

	} else {
		// we don't have an identity yet, so create one

		// publisher.address = pubNode.Address
		timestampStr := time.Now().Format(iotc.TimeFormat)
		validUntil := time.Now().Add(time.Hour * 24 * 365) // valid for 1 year
		validUntilStr := validUntil.Format(iotc.TimeFormat)

		// generate private/public key for signing and store the public key in the publisher identity
		// TODO: store keys
		rng := rand.Reader
		curve := elliptic.P256()
		privKey, err := ecdsa.GenerateKey(curve, rng)
		publisher.privateKeySigning = privKey
		if err != nil {
			publisher.logger.Errorf("Publisher.NewPublisher: Failed to create keys for signing: %s", err)
		}
		privStr, pubStr := messenger.KeysToPem(privKey, &privKey.PublicKey)
		_ = privStr

		addr := publisher.Address()

		publisher.identity = &iotc.PublisherIdentityMessage{
			Address: addr,
			Identity: iotc.PublisherIdentity{
				Domain:       domain,
				IssuerName:   publisherID, // self issued, will be replaced by ZCAS
				Location:     "local",
				Organization: "", // todo: get from messenger configuration
				// PublicKeyCrypto:  cryptoStr,
				PublicKeySigning: pubStr,
				PublisherID:      publisherID,
				Timestamp:        timestampStr,
				ValidUntil:       validUntilStr,
			},
			IdentitySignature: "",
			SignerName:        publisherID,
			Timestamp:         timestampStr,
		}
		SaveIdentity(identityFolder, publisherID, publisher.identity, privKey)
	}
}
