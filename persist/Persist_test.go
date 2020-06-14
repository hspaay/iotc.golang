// Package persist - Load and save publisher application configuration
package persist

import (
	"crypto/x509"
	"testing"

	"github.com/hspaay/iotc.golang/iotc"
	"github.com/hspaay/iotc.golang/messenger"
	"github.com/hspaay/iotc.golang/nodes"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const PublisherID = "publisher1"
const ConfigFolder = "./test"

type TestConfig struct {
	ConfigString string `yaml:"cstring"`
	ConfigNumber int    `yaml:"cnumber"`
}
type MessengerConfig struct {
	Server string `yaml:"server"`         // Messenger hostname or ip address
	Port   uint16 `yaml:"port,omitempty"` // optional port, default is 8883 for TLS
}

// TestAppConfig configuration
func TestAppConfig(t *testing.T) {
	logger := logrus.New()
	logger.SetReportCaller(true) // publisher logging includes caller and file:line#

	appConfig := TestConfig{}
	err := LoadAppConfig(ConfigFolder, PublisherID, &appConfig)
	assert.NoError(t, err, "Failed loading app config")
	assert.Equal(t, appConfig.ConfigNumber, 5, "AppConfig does not contain number 5")
}

// TestMessengerConfig load configuration
func TestMessengerConfig(t *testing.T) {
	logger := logrus.New()
	logger.SetReportCaller(true) // publisher logging includes caller and file:line#

	messengerConfig := MessengerConfig{}
	err := LoadMessengerConfig(ConfigFolder, &messengerConfig)
	assert.NoError(t, err, "Failed loading app config")
	assert.Equal(t, "localhost", messengerConfig.Server, "Messenger does not contain server address")
}

// TestSave configuration
func TestSave(t *testing.T) {
	logger := logrus.New()
	logger.SetReportCaller(true) // publisher logging includes caller and file:line#

	nodeList := make([]*iotc.NodeDiscoveryMessage, 0)
	nodeList2 := make([]*iotc.NodeDiscoveryMessage, 0)
	nodeList = append(nodeList, nodes.NewNode("zone1", "publisher1", "node1", iotc.NodeTypeAdapter))

	err := SaveNodes(ConfigFolder, PublisherID, nodeList)
	assert.NoError(t, err, "Failed saving config")

	err = LoadNodes(ConfigFolder, PublisherID, &nodeList2)
	assert.NoError(t, err, "Failed loading app config")
	nodeList2Node1 := nodeList2[0]
	if assert.NotNil(t, nodeList2Node1) {
		return
	}
	assert.Equal(t, "node1", nodeList2Node1.NodeID)
	// assert.Equal(t, "zone1", nodeList2Node1.Zone)
	assert.Equal(t, "adapter", nodeList2Node1.Type)

}

// TestSave identity, saves and reloads the publisher identity with its private/public key
func TestPersistIdentity(t *testing.T) {
	logger := logrus.New()
	logger.SetReportCaller(true) // publisher logging includes caller and file:line#
	privKey := messenger.CreateAsymKeys()

	ident := iotc.PublisherIdentityMessage{
		Address: "a",
		Identity: iotc.PublisherIdentity{
			Domain:           "mydomain",
			IssuerName:       "test",
			Organization:     "iotc",
			PublisherID:      PublisherID,
			PublicKeySigning: "pubsigning",
			PublicKeyCrypto:  "pubcrypto",
		},
		IdentitySignature: "identSig",
		SignerName:        "test",
	}
	err := SaveIdentity(ConfigFolder, PublisherID, &ident, privKey)
	assert.NoError(t, err, "Failed saving identity")

	// load and compare results
	ident2, privKey2, err := LoadIdentity(ConfigFolder, PublisherID)
	assert.NoError(t, err, "Failed loading identity")
	assert.NotNil(t, ident2, "Unable to read identity")
	assert.NotNil(t, privKey2, "Unable to read private key")
	assert.Equal(t, PublisherID, ident2.Identity.PublisherID)

	pe1, _ := x509.MarshalECPrivateKey(privKey)
	pe2, _ := x509.MarshalECPrivateKey(privKey2)
	assert.Equal(t, pe1, pe2, "public key not identical")
}
