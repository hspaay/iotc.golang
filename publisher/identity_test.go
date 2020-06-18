package publisher

import (
	"crypto/x509"
	"testing"

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
	logger := logrus.New()
	logger.SetReportCaller(true) // publisher logging includes caller and file:line#

	ident, privKey := CreateIdentity(testDomain, testPublisherID)

	err := SaveIdentity(configFolder, testPublisherID, ident)
	assert.NoError(t, err, "Failed saving identity")

	// load and compare results
	ident2, privKey2, err := LoadIdentity(configFolder, testPublisherID)
	assert.NoError(t, err, "Failed loading identity")
	assert.NotNil(t, privKey2, "Unable to read private key")
	if assert.NotNil(t, ident2, "Unable to read identity") {
		assert.Equal(t, testPublisherID, ident2.Public.PublisherID)
		pe1, _ := x509.MarshalECPrivateKey(privKey)
		pe2, _ := x509.MarshalECPrivateKey(privKey2)
		assert.Equal(t, pe1, pe2, "public key not identical")
	}
}
