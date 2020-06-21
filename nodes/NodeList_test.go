package nodes

import (
	"fmt"
	"testing"

	"github.com/hspaay/iotc.golang/iotc"
	"github.com/stretchr/testify/assert"
)

const node1ID = "node1"
const node1AliasID = "alias1"
const publisher1ID = "publisher1"
const publisher2ID = "publisher2"
const domain1ID = iotc.LocalDomainID

var node1Base = fmt.Sprintf("%s/%s/%s", domain1ID, publisher1ID, node1ID)
var node1Alias = fmt.Sprintf("%s/%s/%s", domain1ID, publisher1ID, node1AliasID)
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
	nodeList := NewNodeList()
	node := nodeList.NewNode(domain1ID, publisher1ID, node1ID, iotc.NodeTypeUnknown)

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
	nodeList := NewNodeList()
	nodeList.NewNode(domain1ID, publisher1ID, node1ID, iotc.NodeTypeUnknown)

	newAttr := map[iotc.NodeAttr]string{"Manufacturer": "Bob"}
	nodeList.SetNodeAttr(node1Addr, newAttr)

	newStatus := map[iotc.NodeStatus]string{"LastUpdated": "now"}
	nodeList.SetNodeStatus(node1Addr, newStatus)

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
	nodeList := NewNodeList()
	nodeAddr := nodeList.NewNode(domain1ID, publisher1ID, node1ID, iotc.NodeTypeUnknown)

	config := NewNodeConfig(iotc.DataTypeString, "Friendly Name", "")
	nodeList.UpdateNodeConfig(nodeAddr, iotc.NodeAttrName, config)

	newValues := map[iotc.NodeAttr]string{iotc.NodeAttrName: "NewName"}
	nodeList.SetNodeConfigValues(nodeAddr, newValues)
	// node1 must match the newly added node
	node := nodeList.GetNodeByAddress(node1Addr)
	config2 := node.Config[iotc.NodeAttrName]
	value2 := node.Attr[iotc.NodeAttrName]
	if !assert.NotNil(t, config2, "Can't find configuration for name") {
		return
	}
	assert.Equal(t, "NewName", value2, "Configuration wasn't applied")
}

// TODO more tests for node management and concurrency
