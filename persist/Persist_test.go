// Package persist - Load and save publisher application configuration
package persist_test

import (
	"crypto/x509"
	"testing"

	"github.com/iotdomain/iotdomain-go/identities"
	"github.com/iotdomain/iotdomain-go/nodes"
	"github.com/iotdomain/iotdomain-go/persist"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const PublisherID = "publisher1"
const cacheFolder = "./test"
const configFolder = "./test"

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

	appConfig := TestConfig{}
	err := persist.LoadAppConfig(configFolder, PublisherID, &appConfig)
	assert.NoError(t, err, "Failed loading app config")
	assert.Equal(t, appConfig.ConfigNumber, 5, "AppConfig does not contain number 5")

	// default config folder, file not found error
	err = persist.LoadAppConfig("", PublisherID, &appConfig)
	assert.Error(t, err)
	// unmarshal error
	inputConfig := "sss"
	err = persist.LoadAppConfig(configFolder, PublisherID, &inputConfig)
	assert.Error(t, err)

}

// TestMessengerConfig load configuration
func TestMessengerConfig(t *testing.T) {
	messengerConfig := MessengerConfig{}
	err := persist.LoadMessengerConfig(configFolder, &messengerConfig)
	assert.NoError(t, err, "Failed loading app config")
	assert.Equal(t, "localhost", messengerConfig.Server, "Messenger does not contain server address")
}

func TestSaveNodes(t *testing.T) {
	nodeList := make([]*types.NodeDiscoveryMessage, 0)
	nodeList2 := make([]*types.NodeDiscoveryMessage, 0)
	nodeList = append(nodeList, nodes.NewNode("zone1", "publisher1", "node1", types.NodeTypeAdapter))

	err := persist.SaveNodes(cacheFolder, PublisherID, nodeList)
	assert.NoError(t, err, "Failed saving config")

	err = persist.LoadNodes(cacheFolder, PublisherID, &nodeList2)
	assert.NoError(t, err, "Failed loading app config")
	nodeList2Node1 := nodeList2[0]
	require.NotNil(t, nodeList2Node1)

	assert.Equal(t, "node1", nodeList2Node1.NodeID)
	// assert.Equal(t, "zone1", nodeList2Node1.Zone)
	assert.Equal(t, "adapter", string(nodeList2Node1.NodeType))

	// default cache folder
	err = persist.LoadNodes("", PublisherID, &nodeList2)
	assert.Error(t, err, "Failed loading app config")
	// file not found
	err = persist.LoadNodes("", "unknownPub", &nodeList2)
	assert.Error(t, err)
	// unmarshal error
	err = persist.LoadNodes(cacheFolder, PublisherID, nil)
	assert.Error(t, err)

	// marshal error (cant marshal a function)
	err = persist.SaveToJSON("/doesntexist", PublisherID, TestMessengerConfig)
	assert.Error(t, err)
	// write error, cant write to home
	err = persist.SaveToJSON("/doesntexist", PublisherID, nodeList)
	assert.Error(t, err)

}

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
