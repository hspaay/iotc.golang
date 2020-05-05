package nodes

import (
	"fmt"
	"testing"

	"github.com/hspaay/iotconnect.golang/messaging"
	"github.com/stretchr/testify/assert"
)

const node1ID = "node1"
const node1AliasID = "alias1"
const publisher1ID = "publisher1"
const publisher2ID = "publisher2"
const zone1ID = messaging.LocalZoneID

var node1Base = fmt.Sprintf("%s/%s/%s", zone1ID, publisher1ID, node1ID)
var node1Alias = fmt.Sprintf("%s/%s/%s", zone1ID, publisher1ID, node1AliasID)
var node1Addr = node1Base + "/$node"

var node1ConfigureAddr = node1Base + "/$configure"
var node1InputAddr = node1Base + "/$input/switch/0"
var node1InputSetAddr = node1Base + "/$set/switch/0"

var node1Output1Addr = node1Base + "/$output/switch/0"
var node1Output1Type = "switch"
var node1Output1Instance = "0"

var node1AliasOutput1Addr = node1Alias + "/$output/switch/0"
var node1valueAddr = node1Base + "/$value/switch/0"
var node1latestAddr = node1Base + "/$latest/switch/0"
var node1historyAddr = node1Base + "/$history/switch/0"

// const node2 = new node.Node{}

// TestNewNode instance
func TestNewNode(t *testing.T) {
	nodeList := NewNodeList()
	node := NewNode(zone1ID, publisher1ID, node1ID, messaging.NodeTypeUnknown)
	nodeList.UpdateNode(node)

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
	node := NewNode(zone1ID, publisher1ID, node1ID, messaging.NodeTypeUnknown)
	nodeList.UpdateNode(node)

	newAttr := map[messaging.NodeAttr]string{"Manufacturer": "Bob"}
	nodeList.SetNodeAttr(node1Addr, newAttr)

	newStatus := map[messaging.NodeStatus]string{"LastUpdated": "now"}
	nodeList.SetNodeStatus(node, newStatus)

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
	node := NewNode(zone1ID, publisher1ID, node1ID, messaging.NodeTypeUnknown)
	nodeList.UpdateNode(node)

	config := NewConfigAttr(messaging.NodeAttrName, messaging.DataTypeString, "Friendly Name", "")
	nodeList.SetNodeConfig(node1Addr, config)

	newValues := map[messaging.NodeAttr]string{messaging.NodeAttrName: "NewName"}
	nodeList.SetNodeConfigValues(node1Addr, newValues)

	node = nodeList.GetNodeByAddress(node1Addr)
	c := node.Config[messaging.NodeAttrName]
	if !assert.NotNil(t, c, "Can't find configuration for name") {
		return
	}
	assert.Equal(t, "NewName", c.Value, "Configuration wasn't applied")
}

// TODO more tests for node management and concurrency
