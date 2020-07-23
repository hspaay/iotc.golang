package inputs_test

import (
	"fmt"
	"testing"

	"github.com/iotdomain/iotdomain-go/inputs"
	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const domain = "test"

const node1ID = "node1"
const node2ID = "node2"
const node1AliasID = "alias1"
const publisher1ID = "publisher1"

var node1Base = fmt.Sprintf("%s/%s/%s", domain, publisher1ID, node1ID)
var node2Base = fmt.Sprintf("%s/%s/%s", domain, publisher1ID, node2ID)

// var node1Alias = fmt.Sprintf("%s/%s/%s", domain1ID, publisher1ID, node1AliasID)
// var node1Addr = node1Base + "/$node"

var node1InputAddr = node1Base + "/switch/0/$input"
var node1InputSetAddr = node1Base + "/switch/0/$set"
var node1Input1Type = types.InputTypeSwitch
var node2Input1Address = node2Base + "/" + string(types.InputTypeChannel) + "/0/$input"

func TestNewRegisteredInput(t *testing.T) {
	const Source1ID = "source1"
	collection := inputs.NewRegisteredInputs(domain, publisher1ID)
	require.NotNil(t, collection, "Failed creating registered input collection")

	input := collection.CreateInput(node1ID, types.InputTypeSwitch, types.DefaultInputInstance, nil)
	require.NotNil(t, input, "Failed creating input")

	// must be able to get the newly created input
	input2 := collection.GetInput(node1ID, node1Input1Type, types.DefaultInputInstance)
	require.NotNil(t, input2, "Failed getting created input")
	input2 = collection.GetInputByAddress(node1InputAddr)
	require.NotNil(t, input2, "Failed getting created input")

	// the new input must show in the list of updated inputs
	updated := collection.GetUpdatedInputs(true)
	require.Equal(t, 1, len(updated), "Expected 1 updated input")
	// after the list is clear it should no longer be returned
	updated = collection.GetUpdatedInputs(true)
	require.Equal(t, 0, len(updated), "Expected no more updated inputs")

	// delete input
	collection.DeleteInput(node1ID, node1Input1Type, types.DefaultInputInstance)

	// input with source
	collection.CreateInputWithSource(node1ID, types.InputTypeSwitch, types.DefaultInputInstance, Source1ID, nil)
	require.NotNil(t, input, "Failed creating input with source")
	inputsFromSource := collection.GetInputsWithSource(Source1ID)
	assert.Equal(t, 1, len(inputsFromSource), "Not received the input by its source")

}

func TestUpdateInput(t *testing.T) {
	collection := inputs.NewRegisteredInputs(domain, publisher1ID)

	collection.CreateInput(node1ID, types.InputTypeSwitch, types.DefaultInputInstance, nil)
	updated := collection.GetUpdatedInputs(true)
	require.Equal(t, 1, len(updated), "Expected 1 updated input")
	updatedInput := updated[0]
	assert.Equal(t, node1InputAddr, updatedInput.Address, "Node ID mismatch")
	assert.Equal(t, types.InputTypeSwitch, updatedInput.InputType, "Input Type mismatch")
	assert.Equal(t, types.DefaultInputInstance, updatedInput.Instance, "Input instance mismatch")

	// update non existing input should fail
	input2 := inputs.NewInput(domain, publisher1ID, node2ID, types.InputTypeChannel, types.DefaultInputInstance)
	collection.UpdateInput(input2)
	updated = collection.GetUpdatedInputs(false)
	require.Len(t, updated, 0, "Non existing input should not be updated")

	allInputs := collection.GetAllInputs()
	assert.Equalf(t, 1, len(allInputs), "Expected 1 input but got %d", len(allInputs))

	// update existing Input should succeed
	input1b := inputs.NewInput(domain, publisher1ID, node1ID, types.InputTypeSwitch, types.DefaultInputInstance)
	input1b.Source = "hello"
	collection.UpdateInput(input1b)
	updated = collection.GetUpdatedInputs(false)
	require.Len(t, updated, 1, "Existing input should be updated")
	assert.Equal(t, "hello", input1b.Source, "Updating input not successful")
}

func TestAlias(t *testing.T) {
	const Alias1 = "bob"
	collection := inputs.NewRegisteredInputs(domain, publisher1ID)
	collection.CreateInput(node1ID, types.InputTypeSwitch, types.DefaultInputInstance, nil)
	collection.SetAlias(node1ID, Alias1)

	input1b := collection.GetInput(Alias1, types.InputTypeSwitch, types.DefaultInputInstance)
	require.NotNilf(t, input1b, "Input not retrievable using alias nodeID")
	assert.Equal(t, Alias1, input1b.NodeID, "Input doesn't have the alias NodeID")
}

func TestPublish(t *testing.T) {

	var privKey = messaging.CreateAsymKeys()

	msgr := messaging.NewDummyMessenger(nil)
	signer := messaging.NewMessageSigner(true, getPublisherKey, msgr, privKey)

	collection := inputs.NewRegisteredInputs(domain, publisher1ID)
	input := collection.CreateInput(node1ID, types.InputTypeSwitch, types.DefaultInputInstance, nil)
	require.NotNil(t, input, "Failed creating input")

	allInputs := collection.GetAllInputs()

	inputs.PublishRegisteredInputs(allInputs, signer)
}
