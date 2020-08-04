package nodes_test

import (
	"crypto/ecdsa"
	"encoding/json"
	"testing"

	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/iotdomain/iotdomain-go/nodes"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var dummyConfig = &messaging.MessengerConfig{}

func TestNewDomainNodes(t *testing.T) {
	const Source1ID = "source1"
	const domain = "test"
	const publisherID = "pub2"
	const nodeID = "node1"
	const TestConfigID = "test"
	const TestConfigDefault = "testDefault"
	const node1Addr = domain + "/" + publisherID + "/" + nodeID
	privKey := messaging.CreateAsymKeys()
	getPubKey := func(address string) *ecdsa.PublicKey {
		return &privKey.PublicKey
	}
	messenger := messaging.NewDummyMessenger(dummyConfig)
	signer := messaging.NewMessageSigner(messenger, privKey, getPubKey)

	collection := nodes.NewDomainNodes(signer)
	require.NotNil(t, collection, "Failed creating registered node collection")

	node := nodes.NewNode(domain, publisherID, nodeID, types.NodeTypeAdapter)
	node.Attr[types.NodeAttrName] = "bob"
	node.Config[TestConfigID] = *nodes.NewNodeConfig(types.DataTypeString, "testing", "")
	collection.AddNode(node)

	// must be able to get the newly created node
	addr := nodes.MakeNodeDiscoveryAddress(domain, publisherID, nodeID)
	node2 := collection.GetNodeByAddress(addr)
	require.NotNil(t, node2, "Failed getting created node")
	node3 := collection.GetNodeByAddress("fake/node/address/test/test")
	require.Nil(t, node3, "Unexpected seeing an node3 here")

	pubNodes := collection.GetPublisherNodes(domain + "/" + publisherID)
	assert.Equal(t, 1, len(pubNodes))

	//	node attr
	name := collection.GetNodeAttr(addr, types.NodeAttrName)
	assert.NotEmpty(t, name, "Missing name attribute")
	name = collection.GetNodeAttr("test/noaddress", types.NodeAttrName)
	assert.Empty(t, name, "Missing name attribute")

	// node config
	confValue, err := collection.GetNodeConfigValue(addr, types.NodeAttrName, "default")
	assert.NoError(t, err)
	assert.Equal(t, "bob", confValue, "No default for config attribute")
	confValue, err = collection.GetNodeConfigValue(addr, TestConfigID, TestConfigDefault)
	assert.NoError(t, err)
	assert.Equal(t, TestConfigDefault, confValue)

	confValue, err = collection.GetNodeConfigValue(addr, types.NodeAttrDescription, "default")
	assert.Error(t, err, "Expected error for config not existing")
	name, err = collection.GetNodeConfigValue("test/noaddress", types.NodeAttrName, "default")
	assert.Error(t, err, "Expected error for node not existing")
	assert.Equal(t, "default", name, "Missing name attribute")

	// remove the node
	collection.RemoveNode(node2.Address)
	allNodes := collection.GetAllNodes()
	require.NotNil(t, allNodes, "Failed getting all nodes")
	assert.Equal(t, 0, len(allNodes), "Expected no nodes in GetAllNodes")
}

func TestDiscoverDomainNodes(t *testing.T) {
	const Source1ID = "source1"
	const domain = "test"
	const publisherID = "pub2"
	const nodeID = "node1"
	const node1Addr = domain + "/" + publisherID + "/" + nodeID
	privKey := messaging.CreateAsymKeys()

	getPubKey := func(address string) *ecdsa.PublicKey {
		return &privKey.PublicKey
	}
	messenger := messaging.NewDummyMessenger(dummyConfig)
	signer := messaging.NewMessageSigner(messenger, privKey, getPubKey)

	collection := nodes.NewDomainNodes(signer)
	require.NotNil(t, collection, "Failed creating registered node collection")
	collection.Start()

	node := nodes.NewNode("domain2", "publisher2", "node55", types.NodeTypeAVControl)
	nodeAsBytes, err := json.Marshal(node)
	require.NoErrorf(t, err, "Failed serializing node discovery message")
	messenger.Publish(node.Address, false, string(nodeAsBytes))

	inList := collection.GetAllNodes()
	assert.Equal(t, 1, len(inList), "Expected 1 discovered node. Got %d", len(inList))
	collection.Stop()
}
