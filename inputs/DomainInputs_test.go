package inputs_test

import (
	"crypto/ecdsa"
	"encoding/json"
	"testing"

	"github.com/iotdomain/iotdomain-go/inputs"
	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var dummyConfig = &messaging.MessengerConfig{}

func TestNewDomainInput(t *testing.T) {
	const Source1ID = "source1"
	const domain = "test"
	const publisherID = "pub2"
	const nodeID = "node1"
	const node1Addr = domain + "/" + publisherID + "/" + nodeID
	const inputType = types.InputTypeSwitch
	privKey := messaging.CreateAsymKeys()
	getPubKey := func(address string) *ecdsa.PublicKey {
		return &privKey.PublicKey
	}
	messenger := messaging.NewDummyMessenger(dummyConfig)
	signer := messaging.NewMessageSigner(true, getPubKey, messenger, privKey)

	collection := inputs.NewDomainInputs(signer)
	require.NotNil(t, collection, "Failed creating registered input collection")

	input := inputs.NewInput(domain, publisherID, nodeID, inputType, types.DefaultInputInstance)
	collection.AddInput(input)

	// must be able to get the newly created input
	addr := inputs.MakeInputDiscoveryAddress(domain, publisherID, nodeID, inputType, types.DefaultInputInstance)
	input2 := collection.GetInputByAddress(addr)
	require.NotNil(t, input2, "Failed getting created input")
	input3 := collection.GetInputByAddress("fake/input/address/test/test")
	require.Nil(t, input3, "Unexpected seeing an input3 here")

	// input should be included in list of node inputs
	nodeInputs := collection.GetNodeInputs(node1Addr)
	allInputs := collection.GetAllInputs()
	require.NotNil(t, nodeInputs, "Failed getting node inputs")
	assert.Equal(t, 1, len(nodeInputs), "Expected 1 input for node")
	assert.Equal(t, 1, len(allInputs), "Expected 1 input in GetAllInputs")

	// remove the input
	collection.RemoveInput(input2.Address)
	allInputs = collection.GetAllInputs()
	require.NotNil(t, allInputs, "Failed getting all inputs")
	assert.Equal(t, 0, len(allInputs), "Expected no inputs in GetAllInputs")
}

func TestDiscoverDomainInputs(t *testing.T) {
	const Source1ID = "source1"
	const domain = "test"
	const publisherID = "pub2"
	const nodeID = "node1"
	const node1Addr = domain + "/" + publisherID + "/" + nodeID
	const inputType = types.InputTypeSwitch
	privKey := messaging.CreateAsymKeys()
	getPubKey := func(address string) *ecdsa.PublicKey {
		return &privKey.PublicKey
	}
	messenger := messaging.NewDummyMessenger(dummyConfig)
	signer := messaging.NewMessageSigner(true, getPubKey, messenger, privKey)

	collection := inputs.NewDomainInputs(signer)
	require.NotNil(t, collection, "Failed creating registered input collection")
	collection.Start()

	input := inputs.NewInput("domain2", "publisher2", "node55", types.InputTypeSwitch, types.DefaultInputInstance)
	inputAsBytes, err := json.Marshal(input)
	require.NoErrorf(t, err, "Failed serializing input discovery message")
	messenger.Publish(input.Address, false, string(inputAsBytes))

	inList := collection.GetAllInputs()
	assert.Equal(t, 1, len(inList), "Expected 1 discovered input. Got %d", len(inList))
	collection.Stop()
}
