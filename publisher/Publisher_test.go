package publisher_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/iotdomain/iotdomain-go/inputs"
	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/iotdomain/iotdomain-go/nodes"
	"github.com/iotdomain/iotdomain-go/outputs"
	"github.com/iotdomain/iotdomain-go/publisher"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const domain = "test"

const node1ID = "node1"
const node1AliasID = "alias1"
const publisher1ID = "publisher1"
const publisher2ID = "publisher2"
const identityFolder = "../test"
const cacheFolder = "../test/cache"

var node1Base = fmt.Sprintf("%s/%s/%s", domain, publisher1ID, node1ID)
var node1Alias = fmt.Sprintf("%s/%s/%s", domain, publisher1ID, node1AliasID)
var node1Addr = node1Base + "/$node"
var node1 = nodes.NewNode(domain, publisher1ID, node1ID, types.NodeTypeUnknown)
var node1ConfigureAddr = node1Base + "/$configure"
var node1InputAddr = node1Base + "/switch/0/$input"
var node1InputSetAddr = node1Base + "/switch/0/$set"

var node1Output1Addr = node1Base + "/switch/0/$output"
var node1Output1Type = types.OutputTypeSwitch // "switch"
var node1Output1Instance = "0"

var node1AliasOutput1Addr = node1Alias + "/switch/0/$output"
var node1valueAddr = node1Base + "/switch/0/$raw"
var node1latestAddr = node1Base + "/switch/0/$latest"
var node1historyAddr = node1Base + "/switch/0/$history"

var node1Input1 = inputs.NewInput(domain, publisher1ID, node1ID, "switch", "0")
var node1Output1 = outputs.NewOutput(domain, publisher1ID, node1ID, "switch", "0")
var pubAddr = fmt.Sprintf("%s/%s/$identity", domain, publisher1ID)

var pub2Addr = fmt.Sprintf("%s/%s/$identity", domain, publisher2ID)

var msgConfig *messaging.MessengerConfig = &messaging.MessengerConfig{Domain: domain}
var signMessages = true

// TestNew publisher instance
func TestNewPublisher(t *testing.T) {
	var testMessenger = messaging.NewDummyMessenger(msgConfig)

	pub1 := publisher.NewPublisher(identityFolder, cacheFolder, domain, publisher1ID, signMessages, testMessenger)
	require.NotNil(t, pub1, "Failed creating publisher")

	assert.NotNil(t, pub1.GetIdentity, "Missing publisher identity")
	assert.NotEmpty(t, pub1.Address(), "Missing publisher address")
}

// TestRegister tests if registered nodes, input and output are propery accessible via the publisher
func TestRegister(t *testing.T) {
	var testMessenger = messaging.NewDummyMessenger(msgConfig)
	pub1 := publisher.NewPublisher(identityFolder, cacheFolder, domain, publisher1ID, signMessages, testMessenger)
	pub1.UpdateNode(node1)
	tmpNode := pub1.GetNodeByID(node1ID)
	if !(assert.NotNil(t, tmpNode, "Failed getting registered node") &&
		assert.Equal(t, node1Addr, tmpNode.Address, "Retrieved node not equal to expected node")) {
		return
	}
	assert.Equalf(t, node1Addr, tmpNode.Address, "Node address doesn't match")

	pub1.UpdateInput(node1Input1)
	tmpIn := pub1.GetInput(node1ID, "switch", "0")
	if !(assert.NotNil(t, tmpIn, "Failed getting registered input") &&
		assert.Equal(t, node1Input1.Address, tmpIn.Address, "Retrieved input 1 not equal to registered input 1") &&
		assert.Equal(t, node1InputAddr, tmpIn.Address, "Input address incorrect")) {
		return
	}
	assert.Equalf(t, node1InputAddr, tmpIn.Address, "Input address doesn't match")
	assert.Equalf(t, types.InputTypeSwitch, tmpIn.InputType, "Input Type doesn't match")
	assert.Equalf(t, "0", tmpIn.Instance, "Input Instance doesn't match")

	pub1.UpdateOutput(node1Output1)
	tmpOut := pub1.GetOutput(node1ID, "switch", "0")
	require.NotNil(t, tmpOut, "Failed getting registered output")
	assert.Equal(t, node1Output1.Address, tmpOut.Address, "Retrieved output 1 not equal to registered output 1")
	assert.Equalf(t, node1Output1Addr, tmpOut.Address, "Output address doesn't match")
}

// TestNodePublication tests if registered nodes are published and properly signed
func TestNodePublication(t *testing.T) {
	var testMessenger = messaging.NewDummyMessenger(msgConfig)
	pub1 := publisher.NewPublisher(identityFolder, cacheFolder, domain, publisher1ID, signMessages, testMessenger)

	// Start synchroneous publications to verify publications in order
	pub1.Start()                    // publisher is first publication [0]
	pub1.UpdateNode(node1)          // 2nd [1]
	pub1.UpdateInput(node1Input1)   // 3rd [2]
	pub1.UpdateOutput(node1Output1) // 4th [3]
	pub1.Stop()

	nrPublications := testMessenger.NrPublications()
	if !assert.Equal(t, 4, nrPublications, "Missing publication") {
		return
	}
	p0 := testMessenger.FindLastPublication(node1Addr)
	payload, err := messaging.VerifyJWSMessage(p0, &pub1.GetIdentityKeys().PublicKey)
	assert.NotNilf(t, p0, "Publication for publisher %s not found", pubAddr)
	assert.NoError(t, err, "Publication signing not valid %s", pubAddr)

	// assert.NotEmpty(t, p0.Signature, "Missing signature in publication")
	var p0Node types.NodeDiscoveryMessage
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

// TestAlias tests the use of alias in the inout address publication
func TestAlias(t *testing.T) {
	var testMessenger = messaging.NewDummyMessenger(msgConfig)
	signMessages = true
	pub1 := publisher.NewPublisher(identityFolder, cacheFolder, domain, publisher1ID, signMessages, testMessenger)

	// update the node alias and see if its output is published with alias' as node id
	pub1.Start()
	pub1.UpdateNode(node1)
	// pub1.Nodes.UpdateNodeConfigValues(node1Addr, map[string]string{"alias": node1AliasID})
	node1Output1 := outputs.NewOutput(domain, publisher1ID, node1ID, "switch", "0")
	pub1.UpdateOutput(node1Output1) // expect an output discovery publication with the alias

	pub1.PublishUpdates()
	// time.Sleep(1)

	pub1.HandleAliasCommand(node1Addr, &types.NodeAliasMessage{Alias: node1AliasID})
	node := pub1.GetNodeByID(node1AliasID)
	assert.NotNil(t, node, "Node not found using alias")

	pub1.Stop()

	// p3 := testMessenger.FindLastPublication(node1AliasOutput1Addr) // the output discovery publication
	// require.NotEmpty(t, p3, "output should be published using alias with address %s but no publication was found", node1AliasOutput1Addr)

	// payload, err := messaging.VerifyJWSMessage(p3, &pub1.GetIdentityKeys().PublicKey)
	// require.NoError(t, err, "Failed to verify published message")

	out := pub1.GetDomainOutput(node1AliasOutput1Addr)
	require.NotNilf(t, out, "Output not found by its alias: %s", node1AliasOutput1Addr)
	// var out types.OutputDiscoveryMessage
	// err = json.Unmarshal([]byte(payload), &out)
	// if !assert.NoError(t, err, "Failed to unmarshal published message") {
	// 	return
	// }
	assert.Equal(t, node1AliasOutput1Addr, out.Address, "published output has unexpected address")
	assert.Equal(t, types.OutputTypeOnOffSwitch, out.OutputType, "published output has unexpected iotype")
}

// TestConfigure tests if the node configuration is handled
func TestConfigure(t *testing.T) {
	var testMessenger = messaging.NewDummyMessenger(msgConfig)
	pub1 := publisher.NewPublisher(identityFolder, cacheFolder, domain, publisher1ID, signMessages, testMessenger)

	// update the node alias and see if its output is published with alias' as node id
	pub1.Start() // call start to subscribe to node updates
	pub1.UpdateNode(node1)
	config := nodes.NewNodeConfig(types.DataTypeString, "Friendly Name", "")
	pub1.UpdateNodeConfig(node1ID, "name", config)

	// time.Sleep(time.Second * 1) // receive publications

	// publish a configuration update for the name -> NewName
	// var payload = fmt.Sprintf(`{"address":"%s", "sender": "%s", "timestamp": "%s", "attr": {"name":"NewName"} }`,
	// node1ConfigureAddr, pubAddr, time.Now().Format(types.TimeFormat))
	// // signatureBase64 := messaging.CreateEcdsaSignature(m, pub1.identityPrivateKey)
	// // payload := fmt.Sprintf(`{"signature": "%s", "message": %s }`, signatureBase64, message)
	// message, _ := messaging.CreateJWSSignature(payload, pub1.identityPrivateKey)
	// testMessenger.OnReceive(node1ConfigureAddr, message)
	destination := node1.Address
	privKey := pub1.GetIdentityKeys()
	attrMap := types.NodeAttrMap{"name": "NewName"}
	signer := messaging.NewMessageSigner(true, pub1.GetPublisherKey, testMessenger, privKey)
	nodes.PublishConfigureNode(destination, attrMap, pub1.Address(), signer, &privKey.PublicKey)

	// config := map[string]string{"alias": "myalias"}
	// publisher.UpdateNodeConfig(node1Addr, config) // p2
	// publisher.Outputs.UpdateOutput(node1Output1)        // p3
	pub1.Stop()
	node1 := pub1.GetNodeByID(node1ID)
	c := node1.Attr["name"]
	if !assert.NotNil(t, c, "Can't find configuration for name") {
		return
	}
	assert.Equal(t, "NewName", c, "Configuration wasn't applied")
}

// TestOutputValue tests publication of output values
func TestOutputValue(t *testing.T) {
	var testMessenger = messaging.NewDummyMessenger(msgConfig)
	pub1 := publisher.NewPublisher(identityFolder, cacheFolder, domain, publisher1ID, signMessages, testMessenger)

	// assert.Nilf(t, node1.Config["alias"], "Alias set for node 1, unexpected")
	node1 = nodes.NewNode(domain, publisher1ID, node1ID, types.NodeTypeUnknown)

	// update the node alias and see if its output is published with alias' as node id
	pub1.Start()
	pubKey := &pub1.GetIdentityKeys().PublicKey
	pub1.UpdateNode(node1)
	pub1.UpdateOutput(node1Output1)
	pub1.UpdateOutputValue(node1ID, node1Output1.OutputType, types.DefaultOutputInstance, "true")
	pub1.PublishUpdates()

	// time.Sleep(time.Second * 1) // receive publications
	pub1.Stop()

	// test $raw publication
	p1 := testMessenger.FindLastPublication(node1valueAddr)
	if !assert.NotEmptyf(t, p1, "Unable to find published value on address %s", node1valueAddr) {
		return
	}
	payload, _ := messaging.VerifyJWSMessage(p1, pubKey)
	p1Str := string(payload)
	assert.Equal(t, "true", p1Str, "Published $raw value differs")

	// test $latest publication
	p2 := testMessenger.FindLastPublication(node1latestAddr)
	var latest types.OutputLatestMessage
	if !assert.NotNil(t, p2) {
		return
	}
	payload, _ = messaging.VerifyJWSMessage(p2, pubKey)
	json.Unmarshal([]byte(payload), &latest)
	assert.Equal(t, "true", latest.Value, "Published $latest differs")

	// test $history publication
	p3 := testMessenger.FindLastPublication(node1historyAddr)
	payload, _ = messaging.VerifyJWSMessage(p3, pubKey)
	var history types.OutputHistoryMessage
	json.Unmarshal([]byte(payload), &history)
	assert.Len(t, history.History, 1, "History length differs")

	// test int, float, string list publication
	intList := []int{1, 2, 3}
	valuesAsString, _ := json.Marshal(intList)
	pub1.UpdateOutputValue(node1ID, node1Output1Type, types.DefaultOutputInstance, string(valuesAsString))
	floatList := []float32{1.3, 2.5, 3.09}
	valuesAsString, _ = json.Marshal(floatList)
	pub1.UpdateOutputValue(node1ID, node1Output1Type, types.DefaultOutputInstance, string(valuesAsString))
	stringList := []string{"hello", "world"}
	valuesAsString, _ = json.Marshal(stringList)
	pub1.UpdateOutputValue(node1ID, node1Output1Type, types.DefaultOutputInstance, string(valuesAsString))
}

// TestReceiveInput tests receiving input control commands
func TestReceiveInput(t *testing.T) {
	var testMessenger = messaging.NewDummyMessenger(msgConfig)
	// signMessages = false
	pub1 := publisher.NewPublisher(identityFolder, cacheFolder, domain, publisher1ID, signMessages, testMessenger)

	// update the node alias and see if its output is published with alias' as node id
	pub1.SetNodeInputHandler(func(address string, message *types.SetInputMessage) {
		logrus.Infof("Received message: '%s'", message.Value)
		pub1.UpdateOutputValue(node1ID, node1Output1Type, types.DefaultOutputInstance, message.Value)
	})
	pub1.Start()
	pub1.UpdateNode(node1) // p1
	pub1.UpdateInput(node1Input1)
	pub1.UpdateOutput(node1Output1)
	pub1.PublishUpdates()

	// process background messages
	// time.Sleep(time.Second * 1) // receive publications

	// Pass a set input command to the onreceive handler
	var payload = fmt.Sprintf(`{"address":"%s", "sender": "%s", "timestamp": "%s", "value": "true" }`,
		node1InputSetAddr, pubAddr, time.Now().Format(types.TimeFormat))

	// sign the command
	message, err := messaging.CreateJWSSignature(payload, pub1.GetIdentityKeys())
	assert.NoErrorf(t, err, "signing input message failed")

	// encrypt the command using the GetPublisherKey(..of myself..)
	pubKey := pub1.GetPublisherKey(pub1.Address())
	emessage, err := messaging.EncryptMessage(message, pubKey)

	assert.NoErrorf(t, err, "encrypting input message failed")

	testMessenger.OnReceive(node1InputSetAddr, emessage)

	in1 := pub1.GetInputByAddress(node1InputAddr)
	assert.NotNilf(t, in1, "Input 1 not found on address %s", node1InputAddr)

	val := pub1.GetOutputValue(node1ID, node1Output1Type, types.DefaultOutputInstance)
	if !assert.NotNilf(t, val, "Unable to find output value for output %s", node1Output1Addr) {
		return
	}
	assert.Equal(t, "true", val.Value, "Input value didn't update the output")

	pub1.Stop()
}

// TestSetInput tests the publishing and handling of a set command
func TestSetInput(t *testing.T) {
	var testMessenger = messaging.NewDummyMessenger(msgConfig)
	var receivedInputValue = ""
	pub1 := publisher.NewPublisher(identityFolder, cacheFolder, domain, publisher1ID, signMessages, testMessenger)

	pub1.SetNodeInputHandler(func(address string, message *types.SetInputMessage) {
		receivedInputValue = message.Value
	})
	pub1.UpdateInput(node1Input1)

	pub1.Start()
	// encrypt
	signer := messaging.NewMessageSigner(true, pub1.GetPublisherKey, testMessenger, pub1.GetIdentityKeys())
	pubKey := pub1.GetPublisherKey(node1InputSetAddr)
	inputs.PublishSetInput(node1InputSetAddr, "true", pub1.Address(), signer, pubKey)
	pub1.Stop()
	assert.Equal(t, "true", receivedInputValue, "Expected input value not received")
}

// TestDiscoverPublishers tests receiving other publishers
func TestDiscoverPublishers(t *testing.T) {
	var testMessenger = messaging.NewDummyMessenger(msgConfig)
	pub1 := publisher.NewPublisher(identityFolder, cacheFolder, domain, publisher1ID, false, testMessenger)

	// update the node alias and see if its output is published with alias' as node id
	pub1.Start()

	// Use the dummy messenger for multiple publishers
	pub2 := publisher.NewPublisher(identityFolder, cacheFolder, domain, publisher2ID, signMessages, testMessenger)
	pub2.SetSigningOnOff(true)
	pub2.Start()
	// wait for incoming messages to be processed

	time.Sleep(time.Second * 1) // receive publications

	pub2.Stop()
	pub1.Stop()

	// publisher 1 and 2 should have been discovered
	nrPub := len(pub1.GetDomainPublishers())
	assert.Equal(t, 2, nrPub, "Expected discovery of 2 publishers")
}

// run a bunch of facade commands with invalid arguments
func TestErrors(t *testing.T) {
	var testMessenger = messaging.NewDummyMessenger(msgConfig)
	pub1 := publisher.NewPublisher(identityFolder, cacheFolder, domain, publisher1ID, false, testMessenger)
	pub1.GetDomainInputs()
	pub1.GetDomainNodes()
	pub1.GetDomainOutputs()
	pub1.GetDomainPublishers()
	pub1.GetDomainInput("fakeaddr")
	pub1.GetDomainNode("fakeaddr")
	pub1.GetDomainOutput("fakeaddr")
	pub1.GetIdentity()
	pub1.GetIdentityKeys()
	pub1.GetInput("fakenode", "", "")
	pub1.GetInputs()
	pub1.GetNodeAttr("fakenode", "fakeattr")
	pub1.GetNodeByAddress("fakeaddr")
	pub1.GetNodeByID("fakenode")
	pub1.GetNodeConfigBool("fakeid", "fakeattr", false)
	pub1.GetNodeConfigFloat("fakeid", "fakeattr", 42.0)
	pub1.GetNodeConfigInt("fakeid", "fakeattr", 42)
	pub1.GetNodeConfigString("fakeid", "fakeattr", "fake")
	pub1.GetNodes()
	pub1.GetNodeStatus("fakeid", "fakeattr")
	pub1.GetOutput("fakenode", "", "")
	pub1.GetOutputs()
	pub1.GetOutputValue("fakeid", "faketype", "")
	pub1.MakeNodeDiscoveryAddress("fakeid")
	pub1.NewInput("fakeid", types.InputTypeColor, types.DefaultInputInstance)
	pub1.NewNode("fakeid", types.NodeTypeAlarm)
	pub1.NewOutput("fakeid", types.OutputTypeAlarm, types.DefaultOutputInstance)
	pub1.UpdateNodeErrorStatus("fakeid", types.NodeRunStateError, "fake status")
	pub1.UpdateNodeAttr("fakeid", types.NodeAttrMap{})
	pub1.UpdateNodeConfig("fakeid", types.NodeAttrName, nil)
	pub1.UpdateNodeConfigValues("fakeid", types.NodeAttrMap{})
	pub1.UpdateNodeStatus("fakeid", types.NodeStatusMap{})
}
