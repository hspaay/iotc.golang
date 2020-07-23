package outputs_test

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/iotdomain/iotdomain-go/outputs"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateDomainOutputs(t *testing.T) {
	const domain = "test"
	const publisherID = "pub1"
	const node1ID = "node1"
	const out1Type = types.OutputTypeSwitch
	var node1Base = fmt.Sprintf("%s/%s/%s", domain, publisherID, node1ID)
	var out1Addr = outputs.MakeOutputDiscoveryAddress(
		domain, publisherID, node1ID, out1Type, types.DefaultOutputInstance)

	config := messaging.MessengerConfig{}
	messenger := messaging.NewDummyMessenger(&config)
	privKey := messaging.CreateAsymKeys()
	getPubKey := func(address string) *ecdsa.PublicKey {
		return &privKey.PublicKey
	}
	signer := messaging.NewMessageSigner(true, getPubKey, messenger, privKey)
	collection := outputs.NewDomainOutputs(signer)

	output1 := outputs.NewOutput(domain, publisherID, node1ID, out1Type, types.DefaultOutputInstance)
	collection.AddOutput(output1)

	// output should be included in list of node outputs
	outList := collection.GetAllOutputs()
	assert.NotNil(t, outList)
	assert.Equal(t, 1, len(outList), "expected 1 output")
	nodeOutputs := collection.GetNodeOutputs(node1Base)
	assert.Equal(t, 1, len(nodeOutputs), "Expected 1 output for node")

	// test getting the output
	out1 := collection.GetOutput(node1Base, out1Type, types.DefaultOutputInstance)
	assert.NotNil(t, out1, "Output not found")
	out1 = collection.GetOutputByAddress(out1Addr)
	assert.NotNilf(t, out1, "GetOutputByAddress not found on %s", out1Addr)

	// invalid node address
	out1 = collection.GetOutput(node1ID, types.OutputTypeSwitch, types.DefaultOutputInstance)
	assert.Nil(t, out1)
	out1 = collection.GetOutputByAddress("not/an address")
	assert.Nil(t, out1)

	// remove the output
	collection.RemoveOutput(output1.Address)
	outList = collection.GetAllOutputs()
	require.NotNil(t, outList, "Failed getting all outputs")
	assert.Equal(t, 0, len(outList), "Expected no outputs in GetAlloutputs")

	assert.NotNil(t, collection)
}

func TestDiscoverDomainOutputs(t *testing.T) {
	var dummyConfig = &messaging.MessengerConfig{}
	const Source1ID = "source1"
	const domain = "test"
	const publisherID = "pub2"
	const nodeID = "node1"
	const node1Addr = domain + "/" + publisherID + "/" + nodeID
	const outputType = types.OutputTypeSwitch
	privKey := messaging.CreateAsymKeys()
	getPubKey := func(address string) *ecdsa.PublicKey {
		return &privKey.PublicKey
	}
	messenger := messaging.NewDummyMessenger(dummyConfig)
	signer := messaging.NewMessageSigner(true, getPubKey, messenger, privKey)

	collection := outputs.NewDomainOutputs(signer)
	require.NotNil(t, collection, "Failed creating registered output collection")
	collection.Start()

	output := outputs.NewOutput("domain2", "publisher2", "node55", types.OutputTypeSwitch, types.DefaultOutputInstance)
	outputAsBytes, err := json.Marshal(output)
	require.NoErrorf(t, err, "Failed serializing output discovery message")
	messenger.Publish(output.Address, false, string(outputAsBytes))

	inList := collection.GetAllOutputs()
	assert.Equal(t, 1, len(inList), "Expected 1 discovered output. Got %d", len(inList))
	collection.Stop()
}
