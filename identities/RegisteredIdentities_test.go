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
	var msgConfig *messaging.MessengerConfig = &messaging.MessengerConfig{Domain: domain}
	messenger := messaging.NewDummyMessenger(msgConfig)
	privKey := messaging.CreateAsymKeys()
	var pubKey *ecdsa.PublicKey = &privKey.PublicKey
	getPubKey := func(address string) *ecdsa.PublicKey {
		return pubKey
	}
	signer := messaging.NewMessageSigner(true, getPubKey, messenger, privKey)
	regIdentity := identities.NewRegisteredIdentity(domain, publisher1ID, privKey, signer)
	regIdentity.Start()

	// The DSS is the only one that can update an identity
	// dssPub := identities.NewPublisher(configFolder, cacheFolder, domain, types.DSSPublisherID, false, messenger)
	// dssPub.Start()
	// dssKeys := dssPub.GetIdentityKeys()
	// time.Sleep(time.Second)

	// Create the self-signed DSS identity
	dssIdent, dssKeys := identities.CreateIdentity(domain, dssID)
	dssIdent.IssuerName = "dss"
	dssIdent.Organization = "iotdomain.org"
	dssIdent.Sender = identities.MakePublisherIdentityAddress(domain, dssID)
	dssIdent.IdentitySignature = messaging.CreateIdentitySignature(&dssIdent.PublisherIdentityMessage, dssKeys)
	// payload, _ := json.MarshalIndent(dssIdent, " ", " ")
	// signedPMessage, err := messaging.CreateJWSSignature(string(payload), dssKeys)
	// encryptedMessage, err := messaging.EncryptMessage(signedPMessage, &privKey.PublicKey)
	// assert.NoErrorf(t, err, "Encryption of test message failed")
	//
	// next, create a new identity for this publisher and publish it
	newFullIdent, _ := identities.CreateIdentity(domain, publisher1ID)
	newFullIdent.IssuerName = "dss"
	newFullIdent.Organization = "tester"
	newFullIdent.Sender = dssIdent.Sender
	newFullIdent.IdentitySignature = messaging.CreateIdentitySignature(&newFullIdent.PublisherIdentityMessage, dssKeys)
	payload, _ := json.MarshalIndent(newFullIdent, " ", " ")
	signedMessage, _ := messaging.CreateJWSSignature(string(payload), dssKeys)
	encryptedMessage, _ := messaging.EncryptMessage(signedMessage, &privKey.PublicKey)

	// Signer to use dss for sig verification
	pubKey = &dssKeys.PublicKey
	regIdentity.HandleIdentityUpdate(newFullIdent.Address, encryptedMessage)
	// compare results
	assert.Equal(t, "tester", regIdentity.FullIdentity.Organization, "Identity not updated")

	// error cases - unsigned but encrypted
	encryptedMessage, _ = messaging.EncryptMessage(string(payload), &privKey.PublicKey)
	regIdentity.HandleIdentityUpdate(newFullIdent.Address, encryptedMessage)
	// error cases - signed but not encrypted
	regIdentity.HandleIdentityUpdate(newFullIdent.Address, signedMessage)
	// error cases - sender is not dss
	newFullIdent.Sender = "someoneelse"
	payload, _ = json.MarshalIndent(newFullIdent, " ", " ")
	signedMessage, _ = messaging.CreateJWSSignature(string(payload), dssKeys)
	encryptedMessage, _ = messaging.EncryptMessage(signedMessage, &privKey.PublicKey)
	regIdentity.HandleIdentityUpdate(newFullIdent.Address, encryptedMessage)
	assert.Equal(t, dssIdent.Sender, regIdentity.FullIdentity.Sender, "Identity with invalid sender should not be accepted")

	// update and re-publish, verification should fail as the signature no longer matches
	regIdentity.FullIdentity.Location = "here"
	regIdentity.PublishIdentity()

	regIdentity.Stop()

}

func TestVerifyIdentity(t *testing.T) {
	const domain = "test"

	const publisherID = "publisher1"
	// regIdent := identities.NewRegisteredIdentity(nil)
	ident, privKey := identities.CreateIdentity(domain, publisherID)
	_ = privKey
	assert.NotEmpty(t, ident, "Identity not created")
	assert.Equal(t, domain, ident.Domain)
	assert.Equal(t, publisherID, ident.PublisherID)

	err := identities.VerifyIdentity(ident, domain, publisherID, &privKey.PublicKey)
	assert.NoError(t, err, "Self signed signature should verify against the identity")

	// error case - missing public key in identity
	ident2 := *ident
	ident2.PublicKey = ""
	err = identities.VerifyIdentity(&ident2, domain, publisherID, nil)
	assert.Errorf(t, err, "Identity without public key should fail")

	// error case - identity signature doesn't match content
	ident2.Location = "not a location"
	err = identities.VerifyIdentity(&ident2, domain, publisherID, nil)
	assert.Errorf(t, err, "Identity signature should mispatch")

	// error case - identity expired after 366 days
	ident3 := *ident
	expiredTime := time.Now().Add(-time.Hour * 24 * 366)
	ident3.ValidUntil = expiredTime.Format(types.TimeFormat)
	err = identities.VerifyIdentity(&ident3, domain, publisherID, nil)
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
	err = identities.VerifyIdentity(&ident4, domain, publisherID, &privKey.PublicKey)
	assert.NoError(t, err, "Signature should verify against the identity")
	// but not when modified
	ident4.Location = "not a location"
	ident4.IdentitySignature = messaging.CreateIdentitySignature(&ident4.PublisherIdentityMessage, privKey)
	err = identities.VerifyIdentity(&ident4, domain, publisherID, &privKey.PublicKey)
	assert.Error(t, err, "Signature should fail against a modified identity")

	// mismatch in public/private key of identity
	ident4 = *ident
	newPrivKey := messaging.CreateAsymKeys()
	ident4.PrivateKey = messaging.PrivateKeyToPem(newPrivKey)
	err = identities.VerifyIdentity(&ident4, domain, publisherID, &privKey.PublicKey)
	assert.Error(t, err, "Signature should fail against a mismatched public/private key pem in the identity ")

}
