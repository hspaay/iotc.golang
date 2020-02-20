package publisher

import (
	"fmt"
	"testing"

	"myzone/messenger"
	"myzone/nodes"

	"github.com/stretchr/testify/assert"
)

const zone1ID = "$local"
const publisher1ID = "publisher1"

var pubAddr = fmt.Sprintf("%s/%s/%s/$node", zone1ID, publisher1ID, PublisherNodeID)

const node1ID = "node1"

var node1Addr = fmt.Sprintf("%s/%s/%s/$node", zone1ID, publisher1ID, node1ID)
var node1 = nodes.NewNode(zone1ID, publisher1ID, node1ID)
var node1InputAddr = fmt.Sprintf("%s/%s/%s/$input/switch/0", zone1ID, publisher1ID, node1ID)
var node1Output1Addr = fmt.Sprintf("%s/%s/%s/$output/switch/0", zone1ID, publisher1ID, node1ID)
var node1Input1 = nodes.NewInput(node1, "switch", "0")
var node1Output1 = nodes.NewOutput(node1, "switch", "0")

var testMessenger = messenger.NewDummyMessenger()

// const node2 = new node.Node{}

// TestNew publisher instance
func TestNewPublisher(t *testing.T) {
	pub := NewPublisher(zone1ID, publisher1ID, testMessenger)
	if !assert.NotNil(t, pub, "Failed creating publisher") {
		return
	}
	// should be able to get the publisher node
	pubNode := pub.GetNode(pubAddr)
	assert.NotNil(t, pubNode, "Publisher's node not found")
}

// Test discovery of node and input
func TestDiscover(t *testing.T) {
	pub := NewPublisher(zone1ID, publisher1ID, testMessenger)
	if !assert.NotNil(t, pub, "Failed creating publisher") {
		return
	}

	pub.DiscoverNode(node1)
	tmpNode := pub.GetNode(node1Addr)
	if !(assert.NotNil(t, tmpNode, "Failed getting publisher") &&
		assert.Equal(t, node1.Address, tmpNode.Address, "Retrieved node 1 not equal to discovered node 1")) {
		return
	}

	pub.DiscoverInput(node1Input1)
	tmpIn := pub.GetInput(node1Input1.Address)
	if !(assert.NotNil(t, tmpIn, "Failed getting discovered input") &&
		assert.Equal(t, node1Input1.Address, tmpIn.Address, "Retrieved input 1 not equal to discovered input 1") &&
		assert.Equal(t, node1InputAddr, tmpIn.Address, "Input address incorrect")) {
		return
	}

	pub.DiscoverOutput(node1Output1)
	tmpOut := pub.GetOutput(node1Output1.Address)
	if !(assert.NotNil(t, tmpOut, "Failed getting discovered output") &&
		assert.Equal(t, node1Output1.Address, tmpOut.Address, "Retrieved output 1 not equal to discovered output 1")) {
		return
	}
	assert.NotEqual(t, tmpIn.Address, tmpOut.Address, "Input and output addresses should not be equal")
	assert.Equal(t, tmpIn.IOType, tmpOut.IOType, "Input and output type should be equal")
	assert.Equal(t, tmpIn.Instance, tmpOut.Instance, "Input and output instance should be equal")
}

// TestNodePublication tests if node discovery is published
func TestNodePublication(t *testing.T) {
	pub := NewPublisher(zone1ID, publisher1ID, testMessenger)
	pub.Start(true)
	pub.DiscoverNode(node1)
	pub.Stop()

	if !assert.NotEmpty(t, testMessenger.Publications, "Missing publication") {
		return
	}
	p1 := testMessenger.Publications[0]
	assert.Equal(t, pubAddr, p1.Address, "Publication has different address")
	assert.NotEmpty(t, p1.Signature, "Missing signature in publication")
}

// TestAliasConfig tests if the node configuration is updated and if the alias is used in the inout address
func TestAliasConfig(t *testing.T) {
	// update the node alias and see if its output is published with alias' as node id
	pub := NewPublisher(zone1ID, publisher1ID, testMessenger)
	pub.Start(true)         // p0
	pub.DiscoverNode(node1) // p1

	c := map[string]string{"alias": "myalias"}
	pub.UpdateNodeConfig(node1, c)      // p2
	pub.DiscoverOutput(node1Output1)    // p3
	p3 := testMessenger.Publications[3] // the output discovery publication
	assert.Equal(t, "$local/publisher1/myalias/$output/switch/0", p3.Address, "output discovery address should use myalias, got %s instead", p3.Address)

	// node1.Config["alias"] = &nodes.ConfigAttr{Value: "alias"}

}
