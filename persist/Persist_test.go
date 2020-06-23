// Package persist - Load and save publisher application configuration
package persist

import (
	"testing"

	"github.com/iotdomain/iotdomain-go/nodes"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
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
	logger := logrus.New()
	logger.SetReportCaller(true) // publisher logging includes caller and file:line#

	appConfig := TestConfig{}
	err := LoadAppConfig(configFolder, PublisherID, &appConfig)
	assert.NoError(t, err, "Failed loading app config")
	assert.Equal(t, appConfig.ConfigNumber, 5, "AppConfig does not contain number 5")
}

// TestMessengerConfig load configuration
func TestMessengerConfig(t *testing.T) {
	logger := logrus.New()
	logger.SetReportCaller(true) // publisher logging includes caller and file:line#

	messengerConfig := MessengerConfig{}
	err := LoadMessengerConfig(configFolder, &messengerConfig)
	assert.NoError(t, err, "Failed loading app config")
	assert.Equal(t, "localhost", messengerConfig.Server, "Messenger does not contain server address")
}

// TestSave configuration
func TestSave(t *testing.T) {
	logger := logrus.New()
	logger.SetReportCaller(true) // publisher logging includes caller and file:line#

	nodeList := make([]*types.NodeDiscoveryMessage, 0)
	nodeList2 := make([]*types.NodeDiscoveryMessage, 0)
	nodeList = append(nodeList, nodes.NewNode("zone1", "publisher1", "node1", types.NodeTypeAdapter))

	err := SaveNodesToCache(cacheFolder, PublisherID, nodeList)
	assert.NoError(t, err, "Failed saving config")

	err = LoadNodesFromCache(cacheFolder, PublisherID, &nodeList2)
	assert.NoError(t, err, "Failed loading app config")
	nodeList2Node1 := nodeList2[0]
	if assert.NotNil(t, nodeList2Node1) {
		return
	}
	assert.Equal(t, "node1", nodeList2Node1.NodeID)
	// assert.Equal(t, "zone1", nodeList2Node1.Zone)
	assert.Equal(t, "adapter", nodeList2Node1.Type)

}
