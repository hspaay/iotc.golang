package publisher

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/hspaay/iotc.golang/iotc"
	"github.com/hspaay/iotc.golang/messenger"
	"github.com/hspaay/iotc.golang/nodes"
	"github.com/stretchr/testify/assert"
)

const node1ID = "node1"
const node1AliasID = "alias1"
const publisher1ID = "publisher1"
const publisher2ID = "publisher2"
const domain1ID = "test"
const identityFolder = "../test"
const cacheFolder = "../test/cache"

var node1Base = fmt.Sprintf("%s/%s/%s", domain1ID, publisher1ID, node1ID)
var node1Alias = fmt.Sprintf("%s/%s/%s", domain1ID, publisher1ID, node1AliasID)
var node1Addr = node1Base + "/$node"
var node1 = nodes.NewNode(domain1ID, publisher1ID, node1ID, iotc.NodeTypeUnknown)
var node1ConfigureAddr = node1Base + "/$configure"
var node1InputAddr = node1Base + "/switch/0/$input"
var node1InputSetAddr = node1Base + "/switch/0/$set"

var node1Output1Addr = node1Base + "/switch/0/$output"
var node1Output1Type = "switch"
var node1Output1Instance = "0"

var node1AliasOutput1Addr = node1Alias + "/switch/0/$output"
var node1valueAddr = node1Base + "/switch/0/$raw"
var node1latestAddr = node1Base + "/switch/0/$latest"
var node1historyAddr = node1Base + "/switch/0/$history"

var node1Input1 = nodes.NewInput(node1Addr, "switch", "0")
var node1Output1 = nodes.NewOutput(node1Addr, "switch", "0")
var pubAddr = fmt.Sprintf("%s/%s/$identity", domain1ID, publisher1ID)

var pub2Addr = fmt.Sprintf("%s/%s/$identity", domain1ID, publisher2ID)

var msgConfig *messenger.MessengerConfig = &messenger.MessengerConfig{Domain: domain1ID}

// TestNew publisher instance
func TestNewPublisher(t *testing.T) {
	var testMessenger = messenger.NewDummyMessenger(msgConfig, nil)
	pub1 := NewPublisher(identityFolder, cacheFolder, domain1ID, publisher1ID, testMessenger)
	if !assert.NotNil(t, pub1, "Failed creating publisher") {
		return
	}
	assert.NotNil(t, pub1.identity, "Missing publisher identity")
	assert.NotEmpty(t, pub1.Address(), "Missing publisher address")
	assert.NotEmpty(t, pub1.identity.Address, "Missing publisher identity address")
}

// TestDiscover tests if discovered nodes, input and output are propery accessible via the publisher
func TestDiscover(t *testing.T) {
	var testMessenger = messenger.NewDummyMessenger(msgConfig, nil)
	pub1 := NewPublisher(identityFolder, cacheFolder, domain1ID, publisher1ID, testMessenger)
	pub1.signingMethod = SigningMethodJWS
	pub1.Nodes.UpdateNode(node1)
	tmpNode := pub1.Nodes.GetNodeByAddress(node1Addr)
	if !(assert.NotNil(t, tmpNode, "Failed getting discovered node") &&
		assert.Equal(t, node1Addr, tmpNode.Address, "Retrieved node not equal to expected node")) {
		return
	}
	assert.Equalf(t, node1Addr, tmpNode.Address, "Node address doesn't match")

	pub1.Inputs.UpdateInput(node1Input1)
	tmpIn := pub1.Inputs.GetInput(node1Addr, "switch", "0")
	if !(assert.NotNil(t, tmpIn, "Failed getting discovered input") &&
		assert.Equal(t, node1Input1.Address, tmpIn.Address, "Retrieved input 1 not equal to discovered input 1") &&
		assert.Equal(t, node1InputAddr, tmpIn.Address, "Input address incorrect")) {
		return
	}
	assert.Equalf(t, node1InputAddr, tmpIn.Address, "Input address doesn't match")
	assert.Equalf(t, "switch", tmpIn.InputType, "Input Type doesn't match")
	assert.Equalf(t, "0", tmpIn.Instance, "Input Instance doesn't match")

	pub1.Outputs.UpdateOutput(node1Output1)
	tmpOut := pub1.Outputs.GetOutput(node1Addr, "switch", "0")
	if !(assert.NotNil(t, tmpOut, "Failed getting discovered output") &&
		assert.Equal(t, node1Output1.Address, tmpOut.Address, "Retrieved output 1 not equal to discovered output 1")) {
		return
	}
	assert.Equalf(t, node1Output1Addr, tmpOut.Address, "Output address doesn't match")
}

// TestNodePublication tests if discoveries are published and properly signed
func TestNodePublication(t *testing.T) {
	var testMessenger = messenger.NewDummyMessenger(msgConfig, nil)
	pub1 := NewPublisher(identityFolder, cacheFolder, domain1ID, publisher1ID, testMessenger)
	pub1.signingMethod = SigningMethodJWS

	// Start synchroneous publications to verify publications in order
	pub1.Start()                            // publisher is first publication [0]
	pub1.Nodes.UpdateNode(node1)            // 2nd [1]
	pub1.Inputs.UpdateInput(node1Input1)    // 3rd [2]
	pub1.Outputs.UpdateOutput(node1Output1) // 4th [3]
	pub1.Stop()

	nrPublications := testMessenger.NrPublications()
	if !assert.Equal(t, 4, nrPublications, "Missing publication") {
		return
	}
	p0 := testMessenger.FindLastPublication(node1Addr)
	payload, err := messenger.VerifyJWSMessage(p0, &pub1.privateKeySigning.PublicKey)
	assert.NotNilf(t, p0, "Publication for publisher %s not found", pubAddr)
	assert.NoError(t, err, "Publication signing not valid %s", pubAddr)

	// assert.NotEmpty(t, p0.Signature, "Missing signature in publication")
	var p0Node iotc.NodeDiscoveryMessage
	err = json.Unmarshal([]byte(payload), &p0Node)
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
	var testMessenger = messenger.NewDummyMessenger(msgConfig, nil)
	pub1 := NewPublisher(identityFolder, cacheFolder, domain1ID, publisher1ID, testMessenger)
	pub1.signingMethod = SigningMethodJWS

	// update the node alias and see if its output is published with alias' as node id
	pub1.Start()
	pub1.Nodes.UpdateNode(node1)
	pub1.PublishUpdatedDiscoveries()
	// time.Sleep(1)

	// Stress concurrency, run test with -race
	for i := 1; i < 10; i++ {
		go pub1.Nodes.SetNodeConfigValues(node1Addr, map[iotc.NodeAttr]string{"alias": node1AliasID})
		time.Sleep(130 * time.Millisecond)
		node := pub1.Nodes.GetNodeByAddress(node1Addr)
		_, _ = json.Marshal(node)
	}

	// pub1.Nodes.UpdateNodeConfigValues(node1Addr, map[string]string{"alias": node1AliasID})
	pub1.Outputs.UpdateOutput(node1Output1) // expect an output discovery publication with the alias
	pub1.Stop()

	p3 := testMessenger.FindLastPublication(node1AliasOutput1Addr) // the output discovery publication
	if !assert.NotNil(t, p3, "output discovery should use alias with address %s but no publication was found", node1AliasOutput1Addr) {
		return
	}
	payload, err := messenger.VerifyJWSMessage(p3, &pub1.privateKeySigning.PublicKey)
	if !assert.NoError(t, err, "Failed to verify published message") {
		return
	}

	var out iotc.OutputDiscoveryMessage
	err = json.Unmarshal([]byte(payload), &out)
	if !assert.NoError(t, err, "Failed to unmarshal published message") {
		return
	}
	assert.Equal(t, node1Output1Addr, out.Address, "published output has unexpected address")
	assert.Equal(t, iotc.OutputTypeOnOffSwitch, out.OutputType, "published output has unexpected iotype")
}

// TestConfigure tests if the node configuration is handled
func TestConfigure(t *testing.T) {
	var testMessenger = messenger.NewDummyMessenger(msgConfig, nil)
	pub1 := NewPublisher(identityFolder, cacheFolder, domain1ID, publisher1ID, testMessenger)
	pub1.signingMethod = SigningMethodJWS

	// update the node alias and see if its output is published with alias' as node id
	pub1.Start() // call start to subscribe to node updates
	pub1.Nodes.UpdateNode(node1)
	config := nodes.NewNodeConfig("name", iotc.DataTypeString, "Friendly Name", "")
	pub1.Nodes.UpdateNodeConfig(node1Addr, config)

	// time.Sleep(time.Second * 1) // receive publications

	// publish a configuration update for the name -> NewName
	var payload = fmt.Sprintf(`{"address":"%s", "sender": "%s", "timestamp": "%s", "attr": {"name":"NewName"} }`,
		node1ConfigureAddr, pubAddr, time.Now().Format(iotc.TimeFormat))
	// signatureBase64 := messenger.CreateEcdsaSignature(m, pub1.privateKeySigning)
	// payload := fmt.Sprintf(`{"signature": "%s", "message": %s }`, signatureBase64, message)
	message, _ := messenger.CreateJWSSignature(payload, pub1.privateKeySigning)
	testMessenger.OnReceive(node1ConfigureAddr, message)

	// config := map[string]string{"alias": "myalias"}
	// publisher.UpdateNodeConfig(node1Addr, config) // p2
	// publisher.Outputs.UpdateOutput(node1Output1)        // p3
	pub1.Stop()
	node1 := pub1.Nodes.GetNodeByAddress(node1Addr)
	c := node1.Attr["name"]
	if !assert.NotNil(t, c, "Can't find configuration for name") {
		return
	}
	assert.Equal(t, "NewName", c, "Configuration wasn't applied")
}

// TestOutputValue tests publication of output values
func TestOutputValue(t *testing.T) {
	var testMessenger = messenger.NewDummyMessenger(msgConfig, nil)
	pub1 := NewPublisher(identityFolder, cacheFolder, domain1ID, publisher1ID, testMessenger)
	pub1.signingMethod = SigningMethodJWS

	// assert.Nilf(t, node1.Config["alias"], "Alias set for node 1, unexpected")
	node1 = nodes.NewNode(domain1ID, publisher1ID, node1ID, iotc.NodeTypeUnknown)

	// update the node alias and see if its output is published with alias' as node id
	pub1.Start()
	pub1.Nodes.UpdateNode(node1)
	pub1.Outputs.UpdateOutput(node1Output1)
	pub1.OutputValues.UpdateOutputValue(node1Output1Addr, "true")

	pub1.PublishUpdatedDiscoveries()
	pub1.PublishUpdatedOutputValues()
	// time.Sleep(time.Second * 1) // receive publications
	pub1.Stop()

	// test $raw publication
	p1 := testMessenger.FindLastPublication(node1valueAddr)
	if !assert.NotEmptyf(t, p1, "Unable to find published value on address %s", node1valueAddr) {
		return
	}
	payload, _ := messenger.VerifyJWSMessage(p1, &pub1.privateKeySigning.PublicKey)
	p1Str := string(payload)
	assert.Equal(t, "true", p1Str, "Published $raw value differs")

	// test $latest publication
	p2 := testMessenger.FindLastPublication(node1latestAddr)
	var latest iotc.OutputLatestMessage
	if !assert.NotNil(t, p2) {
		return
	}
	payload, _ = messenger.VerifyJWSMessage(p2, &pub1.privateKeySigning.PublicKey)
	json.Unmarshal([]byte(payload), &latest)
	assert.Equal(t, "true", latest.Value, "Published $latest differs")

	// test $history publication
	p3 := testMessenger.FindLastPublication(node1historyAddr)
	payload, _ = messenger.VerifyJWSMessage(p3, &pub1.privateKeySigning.PublicKey)
	var history iotc.OutputHistoryMessage
	json.Unmarshal([]byte(payload), &history)
	assert.Len(t, history.History, 1, "History length differs")

	// test int, float, string list publication
	intList := []int{1, 2, 3}
	pub1.OutputValues.UpdateOutputIntList(node1Output1Addr, intList)
	floatList := []float32{1.3, 2.5, 3.09}
	pub1.OutputValues.UpdateOutputFloatList(node1Output1Addr, floatList)
	stringList := []string{"hello", "world"}
	pub1.OutputValues.UpdateOutputStringList(node1Output1Addr, stringList)
}

// TestReceiveInput tests receiving input control commands
func TestReceiveInput(t *testing.T) {
	var testMessenger = messenger.NewDummyMessenger(msgConfig, nil)
	pub1 := NewPublisher(identityFolder, cacheFolder, domain1ID, publisher1ID, testMessenger)
	pub1.signingMethod = SigningMethodJWS

	// update the node alias and see if its output is published with alias' as node id
	pub1.SetNodeInputHandler(func(input *iotc.InputDiscoveryMessage, message *iotc.SetInputMessage) {
		pub1.logger.Infof("Received message: '%s'", message.Value)
		pub1.OutputValues.UpdateOutputValue(node1Output1Addr, message.Value)
	})
	pub1.Start()
	pub1.Nodes.UpdateNode(node1) // p1
	pub1.Inputs.UpdateInput(node1Input1)
	pub1.Outputs.UpdateOutput(node1Output1)
	pub1.PublishUpdatedDiscoveries()
	// process background messages
	// time.Sleep(time.Second * 1) // receive publications

	// Pass a set input message to the onreceive handler
	var payload = fmt.Sprintf(`{"address":"%s", "sender": "%s", "timestamp": "%s", "value": "true" }`,
		node1InputSetAddr, pubAddr, time.Now().Format(iotc.TimeFormat))
	message, err := messenger.CreateJWSSignature(payload, pub1.privateKeySigning)

	// encrypter, err := jose.NewEncrypter()

	assert.NoErrorf(t, err, "signing node1 failed")

	testMessenger.OnReceive(node1InputSetAddr, message)

	val := pub1.OutputValues.GetOutputValueByAddress(node1Output1Addr)
	if !assert.NotNilf(t, val, "Unable to find output value for output %s", node1Output1Addr) {
		return
	}
	assert.Equal(t, "true", val.Value, "Input value didn't update the output")

	pub1.Stop()
}

// TestSetInput tests the publishing and handling of a set command
func TestSetInput(t *testing.T) {
	var testMessenger = messenger.NewDummyMessenger(msgConfig, nil)
	var receivedInputValue = ""
	pub1 := NewPublisher(identityFolder, cacheFolder, domain1ID, publisher1ID, testMessenger)
	// pub1.signingMethod = SigningMethodJWS

	pub1.SetNodeInputHandler(func(input *iotc.InputDiscoveryMessage, message *iotc.SetInputMessage) {
		receivedInputValue = message.Value
	})
	pub1.Inputs.UpdateInput(node1Input1)

	pub1.Start()
	pub1.PublishSetInput(node1InputSetAddr, "true")
	pub1.Stop()
	assert.Equal(t, "true", receivedInputValue, "Expected input value not received")
}

// TestDiscoverPublishers tests receiving other publishers
func TestDiscoverPublishers(t *testing.T) {
	var testMessenger = messenger.NewDummyMessenger(msgConfig, nil)
	pub1 := NewPublisher(identityFolder, cacheFolder, domain1ID, publisher1ID, testMessenger)
	pub1.signingMethod = SigningMethodJWS

	// update the node alias and see if its output is published with alias' as node id
	pub1.Start()

	// Use the dummy messenger for multiple publishers
	pub2 := NewPublisher(identityFolder, cacheFolder, domain1ID, publisher2ID, testMessenger)
	pub2.signingMethod = SigningMethodJWS
	pub2.Start()
	// wait for incoming messages to be processed

	time.Sleep(time.Second * 1) // receive publications

	pub2.Stop()
	pub1.Stop()

	// publisher 1 and 2 should have been discovered
	nrPub := len(pub1.domainPublishers.GetAllPublishers())
	assert.Equal(t, 2, nrPub, "Expected discovery of 2 publishers")
}
