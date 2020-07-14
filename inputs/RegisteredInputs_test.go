package inputs_test

import (
	"fmt"
	"testing"

	"github.com/iotdomain/iotdomain-go/inputs"
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

func TestNewInput(t *testing.T) {
	collection := inputs.NewRegisteredInputs(domain, publisher1ID)
	require.NotNil(t, collection, "Failed creating registered input collection")

	input := collection.NewInput(node1ID, types.InputTypeSwitch, types.DefaultInputInstance)
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
	updated = collection.GetUpdatedInputs(false)
	require.Equal(t, 0, len(updated), "Expected no more updated inputs")
}

func TestUpdateInput(t *testing.T) {
	collection := inputs.NewRegisteredInputs(domain, publisher1ID)
	collection.NewInput(node1ID, types.InputTypeSwitch, types.DefaultInputInstance)
	updated := collection.GetUpdatedInputs(true)
	require.Equal(t, 1, len(updated), "Expected 1 updated input")

	input2 := inputs.NewInput(domain, publisher1ID, node2ID, types.InputTypeChannel, types.DefaultInputInstance)
	collection.UpdateInput(input2)
	updated = collection.GetUpdatedInputs(false)
	require.Len(t, updated, 1, "Missing new input")

	input2b := updated[0]
	assert.Equal(t, node2Input1Address, input2b.Address, "Node ID mismatch")
	assert.Equal(t, types.InputTypeChannel, input2b.InputType, "Input Type mismatch")
	assert.Equal(t, types.DefaultInputInstance, input2b.Instance, "Input instance mismatch")

}
