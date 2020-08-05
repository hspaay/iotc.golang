package identities_test

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/iotdomain/iotdomain-go/identities"
	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const configFolder = "../test"

// TestSave identity, saves and reloads the publisher identity with its private/public key
func TestPersistIdentity(t *testing.T) {
	const domain = "test"
	const domain3 = "test3"
	const publisherID = "publisher1"
	const publisherID3 = "publisher3"
	identityFile := configFolder + "/testpersistidentity.json"

	// setup - create and save an identity
	ident, privKey := identities.CreateIdentity(domain, publisherID)
	require.NotEmpty(t, ident, "Identity not created")
	require.Equal(t, domain, ident.Domain)
	require.Equal(t, publisherID, ident.PublisherID)
	err := identities.SaveIdentity(identityFile, ident)
	require.NoError(t, err, "Failed saving identity")

	// load and compare identity
	regIdent2 := identities.NewRegisteredIdentity(domain, publisherID)
	ident2, privKey2, err := regIdent2.LoadIdentity(identityFile)
	require.NoError(t, err, "Failed loading identity")
	require.NotNil(t, ident2, "Unable to read identity")
	assert.Equal(t, domain, ident2.Domain, "Loaded identity doesn't match saved identity")
	assert.Equal(t, publisherID, ident2.PublisherID, "Loaded identity doesn't match saved identity")
	pe1, _ := x509.MarshalECPrivateKey(privKey)
	pe2, _ := x509.MarshalECPrivateKey(privKey2)
	assert.Equal(t, pe1, pe2, "public key not identical")

	// error case using default identity folder and a not yet existing identity
	regIdent3 := identities.NewRegisteredIdentity(domain3, publisherID3)
	ident3, privKey3, err := regIdent3.LoadIdentity("")
	require.NotNil(t, err, "Expected error loading non existing identity")
	require.Nil(t, ident3, "Unable to create identity")
	require.Nil(t, privKey3, "Unable to create identity keys")
	assert.Equal(t, domain, ident2.Domain, "Domain should be unchanged")
	assert.Equal(t, publisherID, ident2.PublisherID, "Publisher should be unchanged")

	// error case - invalid file
	err = identities.SaveIdentity("/root/nofileaccess", ident3)
	assert.Error(t, err, "shoulf fail saving to root")
	// cleanup
	os.Remove(identityFile)
}

func TestUpdateIdentity(t *testing.T) {
	const domain = "test"
	const publisher1ID = "pub1"
	const dssID = types.DSSPublisherID
	var pubKeys = make(map[string]*ecdsa.PublicKey)
	// identityFile := configFolder + "/testpersistidentity.json"

	regIdentity := identities.NewRegisteredIdentity(domain, publisher1ID)
	privKey := regIdentity.GetPrivateKey()
	pubKeys[regIdentity.GetAddress()] = &privKey.PublicKey

	// privKey := messaging.CreateAsymKeys()
	// var pubKey *ecdsa.PublicKey = &privKey.PublicKey
	getPubKey := func(address string) *ecdsa.PublicKey {
		return pubKeys[address]
	}
	// setup the receiver for identity updates
	var msgConfig *messaging.MessengerConfig = &messaging.MessengerConfig{Domain: domain}
	messenger := messaging.NewDummyMessenger(msgConfig)
	signer := messaging.NewMessageSigner(messenger, regIdentity.GetPrivateKey(), getPubKey)
	rxIdent := identities.NewReceiveRegisteredIdentityUpdate(regIdentity, signer)
	rxIdent.Start()

	// The DSS is the only one that can update an identity
	// Create the self-signed DSS identity who will publish the new identity
	dssIdent, dssKeys := identities.CreateIdentity(domain, dssID)
	pubKeys[dssIdent.Address] = &dssKeys.PublicKey
	dssIdent.IssuerID = dssIdent.PublisherID
	dssIdent.Organization = "iotdomain.org"
	dssIdent.Sender = identities.MakePublisherIdentityAddress(domain, dssID)
	messaging.SignIdentity(&dssIdent.PublisherIdentityMessage, dssKeys)
	regIdentity.SetDssKey(&dssKeys.PublicKey)

	// next, create a new identity to publish by the DSS
	newFullIdent, _ := identities.CreateIdentity(domain, publisher1ID)
	newFullIdent.IssuerID = dssIdent.PublisherID
	newFullIdent.Organization = "tester1"
	newFullIdent.Sender = dssIdent.Sender
	messaging.SignIdentity(&newFullIdent.PublisherIdentityMessage, dssKeys)
	payload, _ := json.MarshalIndent(newFullIdent, " ", " ")

	// test update the identity directly, if that doesn't work then the rest will fail too
	regIdentity.UpdateIdentity(newFullIdent)
	ident2, _ := regIdentity.GetFullIdentity()
	require.Equal(t, "tester1", ident2.Organization, "Identity not updated")

	// test through publishing
	newFullIdent.Organization = "tester2"
	messaging.SignIdentity(&newFullIdent.PublisherIdentityMessage, dssKeys)
	payload, _ = json.MarshalIndent(newFullIdent, " ", " ")
	signedMessage, _ := messaging.CreateJWSSignature(string(payload), dssKeys)
	encryptedMessage, _ := messaging.EncryptMessage(signedMessage, &privKey.PublicKey)

	// publish and receive the identity
	rxIdent.ReceiveIdentityUpdate(newFullIdent.Address, encryptedMessage)
	ident2, _ = regIdentity.GetFullIdentity()
	assert.Equal(t, "tester2", ident2.Organization, "Identity not updated")

	// error case - unsigned but encrypted
	encryptedMessage, _ = messaging.EncryptMessage(string(payload), &privKey.PublicKey)
	rxIdent.ReceiveIdentityUpdate(newFullIdent.Address, encryptedMessage)

	// error case - signed but not encrypted
	rxIdent.ReceiveIdentityUpdate(newFullIdent.Address, signedMessage)

	// error case - sender is not the dss
	newFullIdent.Sender = "someoneelse"
	pubKeys[newFullIdent.Sender] = &dssKeys.PublicKey
	payload, _ = json.MarshalIndent(newFullIdent, " ", " ")
	signedMessage, _ = messaging.CreateJWSSignature(string(payload), dssKeys)
	encryptedMessage, _ = messaging.EncryptMessage(signedMessage, &privKey.PublicKey)
	rxIdent.ReceiveIdentityUpdate(newFullIdent.Address, encryptedMessage)
	ident3, _ := regIdentity.GetFullIdentity()
	assert.Equal(t, dssIdent.Sender, ident3.Sender, "Identity with invalid sender should not be accepted")

	// error case - domain doesn't match
	newFullIdent.Sender = dssIdent.Sender
	newFullIdent.Domain = "wrong"
	payload, _ = json.MarshalIndent(newFullIdent, " ", " ")
	signedMessage, _ = messaging.CreateJWSSignature(string(payload), dssKeys)
	encryptedMessage, _ = messaging.EncryptMessage(signedMessage, &privKey.PublicKey)
	rxIdent.ReceiveIdentityUpdate(newFullIdent.Address, encryptedMessage)
	ident4, _ := regIdentity.GetFullIdentity()
	assert.Equal(t, domain, ident4.Domain, "Identity with invalid domain should not be accepted")

	// update and re-publish, verification should fail as the signature no longer matches
	ident3.Location = "here"
	identities.PublishIdentity(&ident3.PublisherIdentityMessage, signer)

	rxIdent.Stop()

}

func TestVerifyIdentity(t *testing.T) {
	const domain = "test"

	const publisherID = "publisher1"
	// regIdent := identities.NewRegisteredIdentity(nil)
	// create and verify self-signed identity
	ident, privKey := identities.CreateIdentity(domain, publisherID)
	_ = privKey
	assert.NotEmpty(t, ident, "Identity not created")
	assert.Equal(t, domain, ident.Domain)
	assert.Equal(t, publisherID, ident.PublisherID)

	err := identities.VerifyFullIdentity(ident, domain, publisherID, &privKey.PublicKey)
	assert.NoError(t, err, "Self signed signature should verify against the identity")

	// error case - missing public key in identity
	ident2 := *ident
	ident2.PublicKey = ""
	err = identities.VerifyFullIdentity(&ident2, domain, publisherID, nil)
	assert.Errorf(t, err, "Identity without public key should fail")

	// error case - identity signature doesn't match content
	ident2.Location = "not a location"
	err = identities.VerifyFullIdentity(&ident2, domain, publisherID, nil)
	assert.Errorf(t, err, "Identity signature should mispatch")

	// error case - identity expired after 366 days
	ident3 := *ident
	expiredTime := time.Now().Add(-time.Hour * 24 * 366)
	ident3.ValidUntil = expiredTime.Format(types.TimeFormat)
	err = identities.VerifyFullIdentity(&ident3, domain, publisherID, nil)
	assert.Errorf(t, err, "Identity is expired")

	// error case - identity public key must match its private key
	pubKeyPem := messaging.PublicKeyToPem(&privKey.PublicKey)
	assert.Equal(t, pubKeyPem, ident.PublicKey)

	// error case - identity signature must verify against its signer
	ident4 := *ident
	ident4.IdentitySignature = ""
	payload, _ := json.Marshal(&ident4.PublisherIdentityMessage)
	sig := messaging.CreateEcdsaSignature(payload, privKey)
	err = messaging.VerifyEcdsaSignature(payload, sig, &privKey.PublicKey)
	assert.NoError(t, err)
	ident4.IdentitySignature = sig
	err = identities.VerifyFullIdentity(&ident4, domain, publisherID, &privKey.PublicKey)
	assert.NoError(t, err, "Signature should verify against the identity")
	// verification fails when identity is modified
	ident4.Location = "not a location"
	err = identities.VerifyFullIdentity(&ident4, domain, publisherID, &privKey.PublicKey)
	assert.Error(t, err, "Signature should fail against a modified identity")

	// mismatch in public/private key of identity
	ident4 = *ident
	newPrivKey := messaging.CreateAsymKeys()
	ident4.PrivateKey = messaging.PrivateKeyToPem(newPrivKey)
	err = identities.VerifyFullIdentity(&ident4, domain, publisherID, &privKey.PublicKey)
	assert.Error(t, err, "Signature should fail against a mismatched public/private key pem in the identity ")

}
