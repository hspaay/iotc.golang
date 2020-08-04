package identities_test

import (
	"crypto/ecdsa"
	"testing"

	"github.com/iotdomain/iotdomain-go/identities"
	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/iotdomain/iotdomain-go/nodes"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var dummyConfig = &messaging.MessengerConfig{}

func TestNewDomainIdentities(t *testing.T) {
	// const Source1ID = "source1"
	// const domain = "test"
	// const publisherID = "pub2"
	// const TestConfigID = "test"
	// const TestConfigDefault = "testDefault"
	privKey := messaging.CreateAsymKeys()
	getPubKey := func(address string) *ecdsa.PublicKey {
		return &privKey.PublicKey
	}
	messenger := messaging.NewDummyMessenger(dummyConfig)
	signer := messaging.NewMessageSigner(messenger, nil, getPubKey)

	collection := nodes.NewDomainNodes(signer)
	require.NotNil(t, collection, "Failed creating registered publisher collection")
}

func TestDiscoverDomainPublishers(t *testing.T) {
	const Source1ID = "source1"
	const domain = "test"
	const publisher2ID = "pub2"
	var err error

	collection := identities.NewDomainPublisherIdentities()
	privKey := messaging.CreateAsymKeys()
	messenger := messaging.NewDummyMessenger(dummyConfig)
	signer1 := messaging.NewMessageSigner(messenger, privKey, collection.GetPublisherKey)
	receiver := identities.NewReceivePublisherIdentities(domain, collection, signer1)
	require.NotNil(t, collection, "Failed creating domain identity collection")
	receiver.Start()

	// publish a self-signed identity as publisher 2. It should be received and verified by the handler
	// of the collection.
	pub2Ident, pub2Keys := identities.CreateIdentity(domain, publisher2ID)
	signer2 := messaging.NewMessageSigner(messenger, pub2Keys, collection.GetPublisherKey)
	addr2 := identities.MakePublisherIdentityAddress(domain, publisher2ID)
	// collection.AddIdentity(&pub2Ident.PublisherIdentityMessage)
	signer2.PublishObject(addr2, false, pub2Ident.PublisherIdentityMessage, nil)

	inList := collection.GetAllPublishers()
	assert.Equal(t, 1, len(inList), "Expected 1 discovered publisher. Got %d", len(inList))
	receiver.Stop()
	pub2b := collection.GetPublisherByAddress(addr2)
	require.NotNil(t, pub2b, "Expected the discovered publisher 2")
	assert.Equal(t, pub2Ident.PublisherID, pub2b.PublisherID)
	pub2Key := collection.GetPublisherKey(addr2)
	assert.NotNil(t, pub2Key, "Expected the discovered publisher 2 key")

	// error case - publisher2 publishes its dss signed identity. Should fail as pub2 is spoofing to be the dss.
	dssKeys := messaging.CreateAsymKeys()
	pub2Copy := *pub2Ident
	pub2Copy.IssuerID = types.DSSPublisherID
	messaging.SignIdentity(&pub2Copy.PublisherIdentityMessage, dssKeys)
	signer2.PublishObject(addr2, false, pub2Copy.PublisherIdentityMessage, nil)
	pub2b = collection.GetPublisherByAddress(addr2)
	assert.Equal(t, pub2Ident.PublisherID, pub2Ident.IssuerID, "Publisher2 should not be able to publish a DSS issued identity")

	// error case - signer is not DSS or publisher itself
	pub2Copy = *pub2Ident
	pub2Copy.IssuerID = "someoneelse"
	messaging.SignIdentity(&pub2Copy.PublisherIdentityMessage, dssKeys)
	signer2.PublishObject(addr2, false, pub2Copy.PublisherIdentityMessage, nil)
	pub2b = collection.GetPublisherByAddress(addr2)
	assert.Equal(t, pub2Ident.PublisherID, pub2Ident.IssuerID, "Publisher2 should not be able to publish a DSS issued identity")

	// error case - can't send a publisher identity from another publisher
	pub2Copy = *pub2Ident
	pub2Copy.IssuerID = pub2Copy.PublisherID
	pub2Copy.Location = "newlocation"
	messaging.SignIdentity(&pub2Copy.PublisherIdentityMessage, pub2Keys)
	signer1.PublishObject(addr2, false, pub2Copy.PublisherIdentityMessage, nil)
	pub2b = collection.GetPublisherByAddress(addr2)
	assert.NotEqual(t, pub2Ident.Location, "newlocation", "Publisher2 should not be able to publish a DSS issued identity")

	// error case - modified identity
	pub2Ident.Location = "modified location"
	err = identities.VerifyPublisherIdentity(pub2Ident.PublisherIdentityMessage.Address,
		&pub2Ident.PublisherIdentityMessage, &dssKeys.PublicKey)
	assert.Errorf(t, err, "modified identity (signed by DSS) should not verify")

	// error case - identity not signed by dss or self
	pub2Ident.IssuerID = "someoneelse"
	messaging.SignIdentity(&pub2Ident.PublisherIdentityMessage, dssKeys)
	err = identities.VerifyPublisherIdentity(pub2Ident.PublisherIdentityMessage.Address,
		&pub2Ident.PublisherIdentityMessage, &dssKeys.PublicKey)
	assert.Errorf(t, err, "Identity not signed by DSS or self must fail")

	// error case - address too short
	pubNoKey := collection.GetPublisherKey(domain + "." + publisher2ID)
	assert.Nil(t, pubNoKey, "Too short address still gets publisher 2 key")

}

func TestDSSiscovery(t *testing.T) {
	const Source1ID = "source1"
	const domain = "test"
	const domain2 = "test2"
	const ident2Addr = domain2 + "/" + types.DSSPublisherID + "/" + types.MessageTypeIdentity

	privKey := messaging.CreateAsymKeys()
	collection := identities.NewDomainPublisherIdentities()
	messenger := messaging.NewDummyMessenger(dummyConfig)
	signer1 := messaging.NewMessageSigner(messenger, privKey, collection.GetPublisherKey)
	receiver := identities.NewReceivePublisherIdentities(domain, collection, signer1)
	require.NotNil(t, collection, "Failed creating domain identity collection")
	receiver.Start()

	// Publish a dss identity
	// Create the self-signed DSS identity
	dssIdent, dssKeys := identities.CreateIdentity(domain, types.DSSPublisherID)
	dssSigner := messaging.NewMessageSigner(messenger, dssKeys, collection.GetPublisherKey)
	dssIdent.IssuerID = types.DSSPublisherID
	dssIdent.Organization = "iotdomain.org"
	dssIdent.Sender = identities.MakePublisherIdentityAddress(domain, types.DSSPublisherID)
	messaging.SignIdentity(&dssIdent.PublisherIdentityMessage, dssKeys)

	// collection.AddIdentity(&dssIdent.PublisherIdentityMessage)
	err := dssSigner.PublishObject(dssIdent.Address, false, dssIdent, nil)
	assert.NoError(t, err, "Publishing DSS identity failed")

	// get identity
	dss2 := collection.GetDSSIdentity(domain)
	assert.NotNil(t, dss2, "Unable to get the DSS identity")
	// error case - get dss for different domain
	dss3 := collection.GetDSSIdentity("nodomain")
	assert.Nil(t, dss3, "Unexpectedly getting a  DSS identity")

	// error case - receive DSS identity for different domain
	dssIdent2, dssKeys2 := identities.CreateIdentity(domain2, types.DSSPublisherID)
	dssIdent2.IssuerID = types.DSSPublisherID
	dssIdent2.Organization = "iotdomain.org"
	dssIdent.Sender = identities.MakePublisherIdentityAddress(domain2, types.DSSPublisherID)
	messaging.SignIdentity(&dssIdent2.PublisherIdentityMessage, dssKeys2)
	dssSigner2 := messaging.NewMessageSigner(messenger, dssKeys2, collection.GetPublisherKey)
	err = dssSigner2.PublishObject(dssIdent2.Address, false, dssIdent2, nil)
	assert.NoError(t, err, "Publishing DSS2 identity failed")

	// Each domain should return their respective DSS identity
	dss4 := collection.GetDSSIdentity(domain)
	assert.Equal(t, domain, dss4.Domain, "")
	dss4b := collection.GetDSSIdentity(domain2)
	require.NotNil(t, dss4b, "DSS identity for domain 2 not updated")
	assert.Equal(t, domain2, dss4b.Domain, "")

	dss5 := collection.GetPublisherByAddress(ident2Addr)
	assert.NotNil(t, dss5, "Should get the DSS of domain2 as a regular publisher")

	receiver.Stop()
}
