// Package persist with configuration for publishers and/or subscribers
package persist

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/iotdomain/iotdomain-go/lib"
	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/iotdomain/iotdomain-go/types"
)

// IdentityFileSuffix to append to name of the file containing saved identity
const IdentityFileSuffix = "-identity.json"

// LoadIdentity loads the publisher identity and private key from file in the given folder.
// The expected identity file is named <publisherID>-identity.json.
// Returns the identity with corresponding ECDSA private key, or nil if no identity is found
// If anything goes wrong, err will contain the error and nil identity is returned
func LoadIdentity(folder string, publisherID string) (fullIdentity *types.PublisherFullIdentity, privateKey *ecdsa.PrivateKey, err error) {
	identityFile := fmt.Sprintf("%s/%s%s", folder, publisherID, IdentityFileSuffix)

	// load the identity
	identityJSON, err := ioutil.ReadFile(identityFile)
	if err != nil {
		return nil, nil, err
	}
	fullIdentity = &types.PublisherFullIdentity{}
	err = json.Unmarshal(identityJSON, fullIdentity)
	if err == nil {
		privateKey = messaging.PrivateKeyFromPem(fullIdentity.PrivateKey)
	}
	return fullIdentity, privateKey, err
}

// SaveIdentity save the full identity of the publisher in the given folder.
// The identity is saved as a json file.
// see also https://stackoverflow.com/questions/21322182/how-to-store-ecdsa-private-key-in-go
func SaveIdentity(folder string, publisherID string, identity *types.PublisherFullIdentity) error {
	identityFile := fmt.Sprintf("%s/%s%s", folder, publisherID, IdentityFileSuffix)

	// save the identity as JSON. Remove first as they are read-only
	identityJSON, _ := json.MarshalIndent(identity, " ", " ")
	os.Remove(identityFile)
	err := ioutil.WriteFile(identityFile, identityJSON, 0400)
	if err != nil {
		return lib.MakeErrorf("SaveIdentity: Unable to save the publisher's identity at %s: %s", identityFile, err)
	}
	return err
}
