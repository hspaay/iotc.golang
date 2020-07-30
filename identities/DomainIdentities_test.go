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
	signer := messaging.NewMessageSigner(true, getPubKey, messenger, nil)

	collection := nodes.NewDomainNodes(signer)
	require.NotNil(t, collection, "Failed creating registered publisher collection")
}

func TestDiscoverDomainPublishers(t *testing.T) {
	const Source1ID = "source1"
	const domain = "test"
	const publisher2ID = "pub2"
	var pubKeyToGet = make(map[string]*ecdsa.PublicKey)

	getPubKey := func(address string) *ecdsa.PublicKey {
		return pubKeyToGet[address]
	}
	privKey := messaging.CreateAsymKeys()
	messenger := messaging.NewDummyMessenger(dummyConfig)
	signer1 := messaging.NewMessageSigner(true, getPubKey, messenger, privKey)
	collection := identities.NewDomainIdentities(signer1)
	require.NotNil(t, collection, "Failed creating domain identity collection")
	collection.Start()

	// publish an identity as publisher 2. It should be received and verified by the handler
	// of the collection.
	pub2Ident, pub2Keys := identities.CreateIdentity(domain, publisher2ID)
	signer2 := messaging.NewMessageSigner(true, getPubKey, messenger, pub2Keys)
	addr2 := identities.MakePublisherIdentityAddress(domain, publisher2ID)
	pubKeyToGet[addr2] = &pub2Keys.PublicKey
	signer2.PublishObject(addr2, false, pub2Ident.PublisherIdentityMessage, nil)

	inList := collection.GetAllPublishers()
	assert.Equal(t, 1, len(inList), "Expected 1 discovered publisher. Got %d", len(inList))
	collection.Stop()

	pub2b := collection.GetPublisherByAddress(addr2)
	assert.Equal(t, pub2Ident.PublisherID, pub2b.PublisherID)

	pub2Key := collection.GetPublisherKey(addr2)
	assert.NotNil(t, pub2Key, "Expected the discovered publisher 2 key")
	// error cases
	pubNoKey := collection.GetPublisherKey(domain + "." + publisher2ID)
	assert.Nil(t, pubNoKey, "Too short address still gets publisher 2 key")

}

func TestDSSiscovery(t *testing.T) {
	const Source1ID = "source1"
	const domain = "test"
	var pubKeyToGet = make(map[string]*ecdsa.PublicKey)

	getPubKey := func(address string) *ecdsa.PublicKey {
		return pubKeyToGet[address]
	}
	privKey := messaging.CreateAsymKeys()
	messenger := messaging.NewDummyMessenger(dummyConfig)
	signer1 := messaging.NewMessageSigner(true, getPubKey, messenger, privKey)
	collection := identities.NewDomainIdentities(signer1)
	require.NotNil(t, collection, "Failed creating domain identity collection")
	collection.Start()

	// Publish a dss identity
	// Create the self-signed DSS identity
	dssIdent, dssKeys := identities.CreateIdentity(domain, types.DSSPublisherID)
	dssSigner := messaging.NewMessageSigner(true, getPubKey, messenger, dssKeys)
	pubKeyToGet[dssIdent.Address] = &dssKeys.PublicKey
	dssIdent.IssuerName = "dss"
	dssIdent.Organization = "iotdomain.org"
	dssIdent.Sender = identities.MakePublisherIdentityAddress(domain, types.DSSPublisherID)
	dssIdent.IdentitySignature = messaging.CreateIdentitySignature(&dssIdent.PublisherIdentityMessage, dssKeys)
	err := dssSigner.PublishObject(dssIdent.Address, false, dssIdent, nil)
	assert.NoError(t, err, "Publishing DSS identity failed")

	// get identity
	dss2 := collection.GetDSSIdentity(domain)
	assert.NotNil(t, dss2, "Unable to get the DSS identity")
	// error
	dss3 := collection.GetDSSIdentity("nodomain")
	assert.Nil(t, dss3, "Unexpectedly getting a  DSS identity")

	collection.Stop()
}
