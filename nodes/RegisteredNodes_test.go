package nodes_test

import (
	"fmt"
	"testing"

	"github.com/iotdomain/iotdomain-go/nodes"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const domain = "test"

const node1ID = "node1"
const node1AliasID = "alias1"
const publisher1ID = "publisher1"
const publisher2ID = "publisher2"

var node1Base = fmt.Sprintf("%s/%s/%s", domain, publisher1ID, node1ID)
var node1Alias = fmt.Sprintf("%s/%s/%s", domain, publisher1ID, node1AliasID)
var node1Addr = node1Base + "/$node"

var node1ConfigureAddr = node1Base + "/$configure"
var node1InputAddr = node1Base + "/switch/0/$input"
var node1InputSetAddr = node1Base + "/switch/0/$set"

var node1Output1Addr = node1Base + "/switch/0/$output"
var node1Output1Type = "switch"
var node1Output1Instance = "0"

var node1AliasOutput1Addr = node1Alias + "/switch/0/$output"
var node1valueAddr = node1Base + "/switch/0/$value"
var node1latestAddr = node1Base + "/switch/0/$latest"
var node1historyAddr = node1Base + "/switch/0/$history"

// const node2 = new node.Node{}

// TestNewNode instance
func TestNewNode(t *testing.T) {
	nodeList := nodes.NewRegisteredNodes(domain, publisher1ID)
	node := nodeList.CreateNode(node1ID, types.NodeTypeUnknown)

	if !assert.NotNil(t, node, "Failed creating node") {
		return
	}
	node2 := nodeList.GetNodeByAddress(node1Addr)
	if !(assert.NotNil(t, node2, "Failed getting created node")) {
		return
	}
}

// Test updating of node atributes and status
func TestAttrStatus(t *testing.T) {
	nodeList := nodes.NewRegisteredNodes(domain, publisher1ID)
	nodeList.CreateNode(node1ID, types.NodeTypeUnknown)

	newAttr := map[types.NodeAttr]string{"Manufacturer": "Bob"}
	nodeList.UpdateNodeAttr(node1ID, newAttr)

	newStatus := map[types.NodeStatus]string{"LastUpdated": "now"}
	nodeList.UpdateNodeStatus(node1ID, newStatus)

	node1 := nodeList.GetNodeByAddress(node1Addr)
	val, hasAttr := node1.Attr["Manufacturer"]
	if !assert.True(t, hasAttr, "Can't find attribute Manufacturer") {
		return
	}
	assert.Equal(t, "Bob", val, "Attribute change wasn't applied")
	val, hasAttr = node1.Status["LastUpdated"]
	if !assert.True(t, hasAttr, "Can't find status attribute LastUpdated") {
		return
	}
	assert.Equal(t, "now", val, "Status 'LastUpdated' wasn't applied")
}

// TestConfigure tests if the node configuration is handled
func TestConfigure(t *testing.T) {
	nodeList := nodes.NewRegisteredNodes(domain, publisher1ID)
	node := nodeList.CreateNode(node1ID, types.NodeTypeUnknown)

	config := nodeList.NewNodeConfig(node1ID, types.NodeAttrName, types.DataTypeString, "Friendly Name", "")
	nodeList.UpdateNodeConfig(node1ID, types.NodeAttrName, config)

	newValues := map[types.NodeAttr]string{types.NodeAttrName: "NewName"}
	nodeList.UpdateNodeConfigValues(node1ID, newValues)

	// node1 must match the newly added node
	node = nodeList.GetNodeByAddress(node1Addr)
	require.NotNil(t, node, "Node %s not found", node1Addr)

	config2 := node.Config[types.NodeAttrName]
	value2 := node.Attr[types.NodeAttrName]
	if !assert.NotNil(t, config2, "Can't find configuration for name") {
		return
	}
	assert.Equal(t, "NewName", value2, "Configuration wasn't applied")
}

// TODO more tests for node management and concurrency
