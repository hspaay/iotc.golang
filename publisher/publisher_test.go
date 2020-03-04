package publisher

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/hspaay/iotconnect.golang/messenger"
	"github.com/hspaay/iotconnect.golang/standard"
	"github.com/stretchr/testify/assert"
)

const node1ID = "node1"
const node1AliasID = "alias1"
const publisher1ID = "publisher1"
const publisher2ID = "publisher2"
const zone1ID = "$local"

var node1Base = fmt.Sprintf("%s/%s/%s", zone1ID, publisher1ID, node1ID)
var node1Alias = fmt.Sprintf("%s/%s/%s", zone1ID, publisher1ID, node1AliasID)
var node1Addr = node1Base + "/$node"
var node1 = standard.NewNode(zone1ID, publisher1ID, node1ID)
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

var node1Input1 = standard.NewInput(node1, "switch", "0")
var node1Output1 = standard.NewOutput(node1, "switch", "0")
var pubAddr = fmt.Sprintf("%s/%s/%s/$node", zone1ID, publisher1ID, standard.PublisherNodeID)

var pub2Addr = fmt.Sprintf("%s/%s/%s/$node", zone1ID, publisher2ID, standard.PublisherNodeID)
var pub2Node = standard.NewNode(zone1ID, publisher2ID, standard.PublisherNodeID)

// const node2 = new node.Node{}

// TestNew publisher instance
func TestNewPublisher(t *testing.T) {
	var testMessenger = messenger.NewDummyMessenger()
	pub1 := NewPublisher(zone1ID, publisher1ID, testMessenger)
	if !assert.NotNil(t, pub1, "Failed creating publisher") {
		return
	}
	tmpNode := pub1.GetNodeByAddress(pubAddr)
	if !(assert.NotNil(t, tmpNode, "Failed getting publisher node") &&
		assert.Equal(t, pubAddr, tmpNode.Address, "Retrieved publisher node not equal to expected node")) {
		return
	}
}

// TestDiscover tests if discovered nodes, input and output are propery accessible via the publisher
func TestDiscover(t *testing.T) {
	var testMessenger = messenger.NewDummyMessenger()
	pub1 := NewPublisher(zone1ID, publisher1ID, testMessenger)
	pub1.DiscoverNode(node1)
	tmpNode := pub1.GetNodeByAddress(node1Addr)
	if !(assert.NotNil(t, tmpNode, "Failed getting discovered node") &&
		assert.Equal(t, node1Addr, tmpNode.Address, "Retrieved node not equal to expected node")) {
		return
	}

	pub1.DiscoverInput(node1Input1)
	tmpIn := pub1.GetInput(node1, "switch", "0")
	if !(assert.NotNil(t, tmpIn, "Failed getting discovered input") &&
		assert.Equal(t, node1Input1.Address, tmpIn.Address, "Retrieved input 1 not equal to discovered input 1") &&
		assert.Equal(t, node1InputAddr, tmpIn.Address, "Input address incorrect")) {
		return
	}

	pub1.DiscoverOutput(node1Output1)
	tmpOut := pub1.GetOutput(node1, "switch", "0")
	if !(assert.NotNil(t, tmpOut, "Failed getting discovered output") &&
		assert.Equal(t, node1Output1.Address, tmpOut.Address, "Retrieved output 1 not equal to discovered output 1")) {
		return
	}
	assert.NotEqual(t, tmpIn.Address, tmpOut.Address, "Input and output addresses should not be equal")
	assert.Equal(t, tmpIn.IOType, tmpOut.IOType, "Input and output type should be equal")
	assert.Equal(t, tmpIn.Instance, tmpOut.Instance, "Input and output instance should be equal")
}

// TestNodePublication tests if discoveries are published.
func TestNodePublication(t *testing.T) {
	var testMessenger = messenger.NewDummyMessenger()
	pub1 := NewPublisher(zone1ID, publisher1ID, testMessenger)

	// Start synchroneous publications to verify publications in order
	pub1.Start(false, nil, nil)       // publisher is first publication [0]
	pub1.DiscoverNode(node1)          // 2nd [1]
	pub1.DiscoverInput(node1Input1)   // 3rd [2]
	pub1.DiscoverOutput(node1Output1) // 4th [3]
	pub1.Stop()

	if !assert.Len(t, testMessenger.Publications, 4, "Missing publication") {
		return
	}
	p0 := testMessenger.FindLastPublication(node1Addr)
	assert.NotNilf(t, p0, "Publication for publisher %s not found", pubAddr)
	assert.NotEmpty(t, p0.Signature, "Missing signature in publication")
	var p0Node standard.Node
	err := json.Unmarshal([]byte(p0.Message), &p0Node)
	assert.NoError(t, err, "Failed parsing node message publication")
	assert.Equal(t, node1Addr, p0Node.Address, "published node doesn't match address")

	p1 := testMessenger.FindLastPublication(node1Addr)
	assert.NotNilf(t, p1, "Publication for node %s not found", node1Addr)

	p2 := testMessenger.FindLastPublication(node1InputAddr)
	assert.NotNilf(t, p2, "Publication for input %s not found", node1InputAddr)

	p3 := testMessenger.FindLastPublication(node1Output1Addr)
	assert.NotNilf(t, p3, "Publication for output %s not found", node1Output1Addr)
}

// TestAlias tests if the the alias is used in the inout address publication
func TestAlias(t *testing.T) {
	var testMessenger = messenger.NewDummyMessenger()
	pub1 := NewPublisher(zone1ID, publisher1ID, testMessenger)

	// update the node alias and see if its output is published with alias' as node id
	pub1.Start(false, nil, nil)
	pub1.DiscoverNode(node1)
	time.Sleep(time.Second * 2) // process publisher discovery

	// var cm = &standard.ConfigureMessage{
	// 	Address: node1Addr,
	// 	Config:  map[string]string{"alias": node1AliasID},
	// 	Sender:  publisher.publisherNode.Address,
	// }
	// mcm, _ := json.Marshal(cm)
	// publication := &messenger.Publication{
	// 	Message: mcm,
	// }
	// publisher.handleNodeConfigCommand(node1Addr, publication)
	pub1.UpdateNodeConfigValue(node1Addr, map[string]string{"alias": node1AliasID})

	pub1.DiscoverOutput(node1Output1) // expect a publication with the alias
	pub1.Stop()

	p3 := testMessenger.FindLastPublication(node1AliasOutput1Addr) // the output discovery publication
	if !assert.NotNil(t, p3, "output discovery should use alias with address %s but no publication was found", node1AliasOutput1Addr) {
		return
	}

	var out standard.InOutput
	err := json.Unmarshal([]byte(p3.Message), &out)
	if !assert.NoError(t, err, "Failed to unmarshal published message") {
		return
	}
	assert.Equal(t, node1Output1Addr, out.Address, "published output has unexpected address")
	assert.Equal(t, standard.IOTypeOnOffSwitch, out.IOType, "published output has unexpected iotype")
}

// TestConfigure tests if the node configuration is handled
func TestConfigure(t *testing.T) {
	var testMessenger = messenger.NewDummyMessenger()
	pub1 := NewPublisher(zone1ID, publisher1ID, testMessenger)

	// update the node alias and see if its output is published with alias' as node id
	pub1.Start(false, nil, nil) // start to subscribe
	pub1.DiscoverNode(node1)
	config := standard.NewConfig("name", standard.DataTypeString, "Friendly Name", "")
	pub1.DiscoverNodeConfig(node1, config)

	time.Sleep(time.Second * 2) // process publisher discovery
	var message = fmt.Sprintf(`{"address":"%s", "sender": "%s", "timestamp": "%s", "config": {"name":"NewName"} }`,
		node1ConfigureAddr, pubAddr, time.Now().Format(standard.TimeFormat))
	// message = `{ "a": "Hello world" }`
	var m json.RawMessage
	m = json.RawMessage(message)
	signatureBase64 := standard.CreateEcdsaSignature(m, pub1.signPrivateKey)
	payload := fmt.Sprintf(`{"signature": "%s", "message": %s }`, signatureBase64, message)
	testMessenger.OnReceive(node1ConfigureAddr, []byte(payload))

	// config := map[string]string{"alias": "myalias"}
	// publisher.UpdateNodeConfig(node1Addr, config) // p2
	// publisher.DiscoverOutput(node1Output1)        // p3
	pub1.Stop()
	node1 := pub1.GetNodeByAddress(node1Addr)
	c := node1.Config["name"]
	if !assert.NotNil(t, c, "Can't find configuration for name") {
		return
	}
	assert.Equal(t, "NewName", c.Value, "Configuration wasn't applied")
}

// TestOutputValue tests publication of output values
func TestOutputValue(t *testing.T) {
	var testMessenger = messenger.NewDummyMessenger()
	pub1 := NewPublisher(zone1ID, publisher1ID, testMessenger)

	// assert.Nilf(t, node1.Config["alias"], "Alias set for node 1, unexpected")
	node1 = standard.NewNode(zone1ID, publisher1ID, node1ID)

	// update the node alias and see if its output is published with alias' as node id
	pub1.Start(false, nil, nil)
	pub1.DiscoverNode(node1)
	pub1.DiscoverOutput(node1Output1)
	pub1.UpdateOutputValue(node1, node1Output1Type, node1Output1Instance, "true")
	time.Sleep(time.Second * 2)
	pub1.Stop()

	// test raw $value publication
	p1 := testMessenger.FindLastPublication(node1valueAddr)
	if !assert.NotNilf(t, p1, "Unable to find published value on address", node1valueAddr) {
		return
	}
	p1Str := string(p1.Message)
	assert.Equal(t, "true", p1Str, "Published $value differs")

	// test $latest publication
	p2 := testMessenger.FindLastPublication(node1latestAddr)
	var latest standard.LatestMessage
	if !assert.NotNil(t, p2.Message) {
		return
	}
	json.Unmarshal([]byte(p2.Message), &latest)
	assert.Equal(t, "true", latest.Value, "Published $latest differs")

	// test $history publication
	p3 := testMessenger.FindLastPublication(node1historyAddr)
	var history standard.HistoryMessage
	json.Unmarshal([]byte(p3.Message), &history)
	assert.Len(t, history.History, 1, "History length differs")
}

// TestReceiveInput tests receiving input control commands
func TestReceiveInput(t *testing.T) {
	var testMessenger = messenger.NewDummyMessenger()
	pub1 := NewPublisher(zone1ID, publisher1ID, testMessenger)

	// update the node alias and see if its output is published with alias' as node id
	pub1.Start(false, nil, func(input *standard.InOutput, message *standard.SetMessage) {
		pub1.Logger.Infof("Received message: '%s'", message.Value)
		pub1.UpdateOutputValue(node1, input.IOType, input.Instance, message.Value)
	})
	pub1.DiscoverNode(node1) // p1
	pub1.DiscoverInput(node1Input1)
	pub1.DiscoverOutput(node1Output1)
	// process background messages
	time.Sleep(time.Second * 2)

	var message = fmt.Sprintf(`{"address":"%s", "sender": "%s", "timestamp": "%s", "value": "true" }`,
		node1InputSetAddr, pubAddr, time.Now().Format(standard.TimeFormat))
	signatureBase64 := standard.CreateEcdsaSignature([]byte(message), pub1.signPrivateKey)
	// publicKey := &publisher.signPrivateKey.PublicKey
	// test := publisher.ecdsaVerify([]byte(message), signatureBase64, publicKey)
	// _ = test

	payload := fmt.Sprintf(`{"signature": "%s", "message": %s }`, signatureBase64, message)
	testMessenger.OnReceive(node1InputSetAddr, []byte(payload))

	val := pub1.GetOutputValue(node1, node1Output1Type, node1Output1Instance)
	assert.Equal(t, "true", val, "Input value didn't update the output")

	pub1.Stop()
}

// TestDiscoveryPublishers tests receiving other publishers
func TestDiscoveryPublishers(t *testing.T) {
	var testMessenger = messenger.NewDummyMessenger()
	pub1 := NewPublisher(zone1ID, publisher1ID, testMessenger)

	// update the node alias and see if its output is published with alias' as node id
	pub1.Start(false, nil, nil)

	publisher2 := NewPublisher(zone1ID, publisher2ID, testMessenger)
	publisher2.Start(false, nil, nil)
	// wait for incoming messages to be processed
	time.Sleep(time.Second * 2)
	publisher2.Stop()
	pub1.Stop()

	// publisher 1 and 2 should have been discovered
	assert.Len(t, pub1.zonePublishers, 2, "Should have discovered 2 publishers")

}
