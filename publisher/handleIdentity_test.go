package publisher_test

import (
	"crypto/x509"
	"encoding/json"
	"testing"
	"time"

	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/iotdomain/iotdomain-go/publisher"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const testDomain = "test"
const testPublisherID = "publisher1"
const configFolder = "../test"

// type TestConfig struct {
// 	ConfigString string `yaml:"cstring"`
// 	ConfigNumber int    `yaml:"cnumber"`
// }
// type MessengerConfig struct {
// 	Server string `yaml:"server"`         // Messenger hostname or ip address
// 	Port   uint16 `yaml:"port,omitempty"` // optional port, default is 8883 for TLS
// }

// TestSave identity, saves and reloads the publisher identity with its private/public key
func TestPersistIdentity(t *testing.T) {
	logrus.SetReportCaller(true) // publisher logging includes caller and file:line#

	ident, privKey := publisher.CreateIdentity(testDomain, testPublisherID)
	assert.NotEmpty(t, ident, "Identity not created")

	err := publisher.SaveIdentity(configFolder, testPublisherID, ident)
	assert.NoError(t, err, "Failed saving identity")

	// load and compare results
	ident2, privKey2, err := publisher.LoadIdentity(configFolder, testPublisherID)
	assert.NoError(t, err, "Failed loading identity")
	assert.NotNil(t, privKey2, "Unable to read private key")
	if assert.NotNil(t, ident2, "Unable to read identity") {
		assert.Equal(t, testPublisherID, ident2.PublisherID)
		pe1, _ := x509.MarshalECPrivateKey(privKey)
		pe2, _ := x509.MarshalECPrivateKey(privKey2)
		assert.Equal(t, pe1, pe2, "public key not identical")
	}
}

// TestUpdateIdentity, saves and reloads the publisher identity with its private/public key
func TestUpdateIdentity(t *testing.T) {
	logrus.SetReportCaller(true) // publisher logging includes caller and file:line#
	var msgConfig *messaging.MessengerConfig = &messaging.MessengerConfig{Domain: domain}
	messenger := messaging.NewDummyMessenger(msgConfig)
	pub := publisher.NewPublisher(configFolder, cacheFolder, domain, testPublisherID, false, messenger)
	pub.Start()
	privKey := pub.GetIdentityKeys()

	// The DSS is the only one that can update an identity
	dssPub := publisher.NewPublisher(configFolder, cacheFolder, domain, types.DSSPublisherID, false, messenger)
	dssPub.Start()
	dssKeys := dssPub.GetIdentityKeys()
	time.Sleep(time.Second)

	// Create a new and updated identity for the publisher, signed by the DSS
	newIdent := pub.FullIdentity()
	newIdent.IssuerName = "dss"
	newIdent.Organization = "iotdomain.org"
	newIdent.IdentitySignature = ""
	identSignature := messaging.SignEncodeIdentity(&newIdent.PublisherIdentityMessage, dssKeys)
	newIdent.IdentitySignature = identSignature
	newIdent.Sender = dssPub.Address()
	payload, _ := json.MarshalIndent(newIdent, " ", " ")
	signedPMessage, err := messaging.CreateJWSSignature(string(payload), dssKeys)
	encryptedMessage, err := messaging.EncryptMessage(signedPMessage, &privKey.PublicKey)
	assert.NoErrorf(t, err, "Encryption of test message failed")
	//
	pub.HandleIdentityUpdate(newIdent.Address, encryptedMessage)

	// compare results
	receivedIdent := pub.FullIdentity()
	assert.Equal(t, "iotdomain.org", receivedIdent.Organization, "Identity not updated")

	pub.Stop()
	dssPub.Stop()

}
