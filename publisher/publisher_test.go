package publisher

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"iotzone/messenger"
	"iotzone/standard"

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
	publisher := NewPublisher(zone1ID, publisher1ID, testMessenger)
	if !assert.NotNil(t, publisher, "Failed creating publisher") {
		return
	}
	tmpNode := publisher.GetNode(pubAddr)
	if !(assert.NotNil(t, tmpNode, "Failed getting publisher node") &&
		assert.Equal(t, pubAddr, tmpNode.Address, "Retrieved publisher node not equal to expected node")) {
		return
	}
}

// TestDiscover tests if discovered nodes, input and output are propery accessible via the publisher
func TestDiscover(t *testing.T) {
	var testMessenger = messenger.NewDummyMessenger()
	publisher := NewPublisher(zone1ID, publisher1ID, testMessenger)
	publisher.DiscoverNode(node1)
	tmpNode := publisher.GetNode(node1Addr)
	if !(assert.NotNil(t, tmpNode, "Failed getting discovered node") &&
		assert.Equal(t, node1Addr, tmpNode.Address, "Retrieved node not equal to expected node")) {
		return
	}

	publisher.DiscoverInput(node1Input1)
	tmpIn := publisher.GetInput(node1Input1.Address)
	if !(assert.NotNil(t, tmpIn, "Failed getting discovered input") &&
		assert.Equal(t, node1Input1.Address, tmpIn.Address, "Retrieved input 1 not equal to discovered input 1") &&
		assert.Equal(t, node1InputAddr, tmpIn.Address, "Input address incorrect")) {
		return
	}

	publisher.DiscoverOutput(node1Output1)
	tmpOut := publisher.GetOutput(node1Output1.Address)
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
	publisher := NewPublisher(zone1ID, publisher1ID, testMessenger)

	// Start synchroneous publications to verify publications in order
	publisher.Start(false, nil, nil)       // publisher is first publication [0]
	publisher.DiscoverNode(node1)          // 2nd [1]
	publisher.DiscoverInput(node1Input1)   // 3rd [2]
	publisher.DiscoverOutput(node1Output1) // 4th [3]
	publisher.Stop()

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
	publisher := NewPublisher(zone1ID, publisher1ID, testMessenger)

	// update the node alias and see if its output is published with alias' as node id
	publisher.Start(true, nil, nil)
	publisher.DiscoverNode(node1)

	config := map[string]string{"alias": node1AliasID}
	publisher.UpdateNodeConfigValue(node1Addr, config)

	publisher.DiscoverOutput(node1Output1) // expect a publication with the alias
	publisher.Stop()

	p3 := testMessenger.FindLastPublication(node1AliasOutput1Addr) // the output discovery publication
	assert.NotNil(t, p3, "output discovery should use alias with address %s but no publication was found", node1AliasOutput1Addr)

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
	publisher := NewPublisher(zone1ID, publisher1ID, testMessenger)

	// update the node alias and see if its output is published with alias' as node id
	publisher.Start(true, nil, nil) // start to subscribe
	publisher.DiscoverNode(node1)
	publisher.DiscoverNodeConfig(node1, "name", &standard.ConfigAttr{Description: "Friendly Name"})

	var message = fmt.Sprintf(`{"address":"%s", "sender": "%s", "timestamp": "%s", "config": {"name":"NewName"} }`,
		node1ConfigureAddr, pubAddr, time.Now().Format(standard.TimeFormat))
	// message = `{ "a": "Hello world" }`
	payload := fmt.Sprintf(`{"signature": "123", "message": %s }`, message)
	testMessenger.OnReceive(node1ConfigureAddr, []byte(payload))

	// config := map[string]string{"alias": "myalias"}
	// publisher.UpdateNodeConfig(node1Addr, config) // p2
	// publisher.DiscoverOutput(node1Output1)        // p3
	publisher.Stop()
	node1 := publisher.GetNode(node1Addr)
	c := node1.Config["name"]
	if !assert.NotNil(t, c, "Can't find configuration for name") {
		return
	}
	assert.Equal(t, "NewName", c.Value, "Configuration wasn't applied")
}

// TestOutputValue tests publication of output values
func TestOutputValue(t *testing.T) {
	var testMessenger = messenger.NewDummyMessenger()
	publisher := NewPublisher(zone1ID, publisher1ID, testMessenger)

	// update the node alias and see if its output is published with alias' as node id
	publisher.Start(false, nil, nil)       // p0
	publisher.DiscoverOutput(node1Output1) // p3
	publisher.UpdateOutputValue(node1Output1Addr, "true")
	publisher.Stop()

	// test raw $value publication
	p1 := testMessenger.FindLastPublication(node1valueAddr)
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
	publisher := NewPublisher(zone1ID, publisher1ID, testMessenger)

	// update the node alias and see if its output is published with alias' as node id
	publisher.Start(false, nil, func(input *standard.InOutput, message *standard.SetMessage) {
		publisher.Logger.Infof("Received message: %s", message.Value)
		// find the corresponding output address
		output := publisher.GetOutput(input.Address)
		if !assert.NotNilf(t, output, "No corresponding output for address %s", input.Address) {
			return
		}
		standard.UpdateValue(output, message.Value)
	})
	publisher.DiscoverNode(node1) // p1
	publisher.DiscoverInput(node1Input1)
	publisher.DiscoverOutput(node1Output1)

	var message = fmt.Sprintf(`{"address":"%s", "sender": "%s", "timestamp": "%s", "value": "true" }`,
		node1InputSetAddr, pubAddr, time.Now().Format(standard.TimeFormat))
	// message = `{ "a": "Hello world" }`
	payload := fmt.Sprintf(`{"signature": "123", "message": %s }`, message)
	testMessenger.OnReceive(node1InputSetAddr, []byte(payload))

	val := standard.GetOutputValue(node1Output1)
	assert.Equal(t, "true", val, "Input value didn't update the output")

	publisher.Stop()
}

// TestDiscoveryPublishers tests receiving other publishers
func TestDiscoveryPublishers(t *testing.T) {
	var testMessenger = messenger.NewDummyMessenger()
	publisher := NewPublisher(zone1ID, publisher1ID, testMessenger)

	// update the node alias and see if its output is published with alias' as node id
	publisher.Start(false, nil, nil)

	publisher2 := NewPublisher(zone1ID, publisher2ID, testMessenger)
	publisher2.Start(false, nil, nil)
	// wait for incoming messages to be processed
	time.Sleep(time.Second * 2)
	publisher2.Stop()
	publisher.Stop()

	// publisher 1 and 2 should have been discovered
	assert.Len(t, publisher.zonePublishers, 2, "Should have discovered 2 publishers")

}
