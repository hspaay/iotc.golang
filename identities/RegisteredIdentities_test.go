package identities_test

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/json"
	"testing"
	"time"

	"github.com/iotdomain/iotdomain-go/identities"
	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/iotdomain/iotdomain-go/persist"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const configFolder = "../test"

// TestSave identity, saves and reloads the publisher identity with its private/public key
func TestPersistIdentity(t *testing.T) {
	const domain = "test"
	const publisherID = "publisher1"
	// regIdent := identities.NewRegisteredIdentity(nil)
	ident, privKey := identities.CreateIdentity(domain, publisherID)
	assert.NotEmpty(t, ident, "Identity not created")
	assert.Equal(t, domain, ident.Domain)
	assert.Equal(t, publisherID, ident.PublisherID)

	err := persist.SaveIdentity(configFolder, publisherID, ident)
	assert.NoError(t, err, "Failed saving identity")

	// load and compare results
	ident2, privKey2, err := persist.LoadIdentity(configFolder, publisherID)
	assert.NoError(t, err, "Failed loading identity")
	assert.NotNil(t, privKey2, "Unable to read private key")
	require.NotNil(t, ident2, "Unable to read identity")
	assert.Equal(t, publisherID, ident2.PublisherID)
	pe1, _ := x509.MarshalECPrivateKey(privKey)
	pe2, _ := x509.MarshalECPrivateKey(privKey2)
	assert.Equal(t, pe1, pe2, "public key not identical")

	// now lets load it all using SetupPublisherIdentity
	ident3, privKey3 := identities.SetupPublisherIdentity(configFolder, domain, publisherID, nil)
	require.NotNil(t, ident3, "Unable to read identity")
	require.NotNil(t, privKey3, "Unable to get identity keys")

	// error case using default identity folder and a not yet existing identity
	ident4, privKey4 := identities.SetupPublisherIdentity("", domain, publisherID, nil)
	require.NotNil(t, ident4, "Unable to create identity")
	require.NotNil(t, privKey4, "Unable to create identity keys")

}

func TestUpdateIdentity(t *testing.T) {
	const domain = "test"
	const publisher1ID = "pub1"
	const dssID = types.DSSPublisherID
	var pubKeys = make(map[string]*ecdsa.PublicKey)

	regIdentity, privKey := identities.NewRegisteredIdentity(configFolder, domain, publisher1ID)
	pubKeys[regIdentity.GetAddress()] = &privKey.PublicKey

	// privKey := messaging.CreateAsymKeys()
	// var pubKey *ecdsa.PublicKey = &privKey.PublicKey
	getPubKey := func(address string) *ecdsa.PublicKey {
		return pubKeys[address]
	}

	var msgConfig *messaging.MessengerConfig = &messaging.MessengerConfig{Domain: domain}
	messenger := messaging.NewDummyMessenger(msgConfig)
	signer := messaging.NewMessageSigner(messenger, privKey, getPubKey)
	rxIdent := identities.NewReceiveRegisteredIdentityUpdate(regIdentity, signer)
	rxIdent.Start()

	// The DSS is the only one that can update an identity
	// dssPub := identities.NewPublisher(configFolder, cacheFolder, domain, types.DSSPublisherID, false, messenger)
	// dssPub.Start()
	// dssKeys := dssPub.GetIdentityKeys()
	// time.Sleep(time.Second)

	// Create the self-signed DSS identity who will publish the new identity
	dssIdent, dssKeys := identities.CreateIdentity(domain, dssID)
	pubKeys[dssIdent.Address] = &dssKeys.PublicKey
	dssIdent.IssuerID = dssIdent.PublisherID
	dssIdent.Organization = "iotdomain.org"
	dssIdent.Sender = identities.MakePublisherIdentityAddress(domain, dssID)
	messaging.SignIdentity(&dssIdent.PublisherIdentityMessage, dssKeys)
	regIdentity.SetDssKey(&dssKeys.PublicKey)

	// next, create a new identity to publish
	newFullIdent, _ := identities.CreateIdentity(domain, publisher1ID)
	newFullIdent.IssuerID = dssIdent.PublisherID
	newFullIdent.Organization = "tester1"
	newFullIdent.Sender = dssIdent.Sender
	messaging.SignIdentity(&newFullIdent.PublisherIdentityMessage, dssKeys)
	payload, _ := json.MarshalIndent(newFullIdent, " ", " ")

	// test update the identity directly, if that doesn't work then the rest will fail too
	regIdentity.UpdateIdentity(newFullIdent)
	ident2, _ := regIdentity.GetIdentity()
	require.Equal(t, "tester1", ident2.Organization, "Identity not updated")

	// test through publishing
	newFullIdent.Organization = "tester2"
	messaging.SignIdentity(&newFullIdent.PublisherIdentityMessage, dssKeys)
	payload, _ = json.MarshalIndent(newFullIdent, " ", " ")
	signedMessage, _ := messaging.CreateJWSSignature(string(payload), dssKeys)
	encryptedMessage, _ := messaging.EncryptMessage(signedMessage, &privKey.PublicKey)

	// publish and receive the identity
	rxIdent.ReceiveIdentityUpdate(newFullIdent.Address, encryptedMessage)
	ident2, _ = regIdentity.GetIdentity()
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
	ident3, _ := regIdentity.GetIdentity()
	assert.Equal(t, dssIdent.Sender, ident3.Sender, "Identity with invalid sender should not be accepted")

	// error case - domain doesn't match
	newFullIdent.Sender = dssIdent.Sender
	newFullIdent.Domain = "wrong"
	payload, _ = json.MarshalIndent(newFullIdent, " ", " ")
	signedMessage, _ = messaging.CreateJWSSignature(string(payload), dssKeys)
	encryptedMessage, _ = messaging.EncryptMessage(signedMessage, &privKey.PublicKey)
	rxIdent.ReceiveIdentityUpdate(newFullIdent.Address, encryptedMessage)
	ident4, _ := regIdentity.GetIdentity()
	assert.Equal(t, domain, ident4.Domain, "Identity with invalid domain should not be accepted")

	// update and re-publish, verification should fail as the signature no longer matches
	ident3.Location = "here"
	regIdentity.PublishIdentity(signer)

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
