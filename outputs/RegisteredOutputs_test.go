package outputs_test

import (
	"fmt"
	"testing"

	"github.com/iotdomain/iotdomain-go/outputs"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const domain = "test"

const node1ID = "node1"
const node1AliasID = "alias1"
const publisher1ID = "publisher1"
const publisher2ID = "publisher2"

var node1Base = fmt.Sprintf("%s/%s/%s", domain, publisher1ID, node1ID)

// var node1Alias = fmt.Sprintf("%s/%s/%s", domain1ID, publisher1ID, node1AliasID)
var node1Addr = node1Base + "/$node"

var node1Output1Addr = node1Base + "/switch/0/$output"
var node1Output1Type = types.OutputTypeSwitch
var node1Output1Instance = "0"

func TestCreateOutputs(t *testing.T) {
	collection := outputs.NewRegisteredOutputs(domain, publisher1ID)
	output := collection.CreateOutput(node1ID, types.OutputTypeSwitch, types.DefaultOutputInstance)

	require.NotNil(t, output, "Failed creating output")

	output2 := collection.GetOutput(node1ID, node1Output1Type, types.DefaultOutputInstance)
	require.NotNil(t, output2, "Failed getting created output")

	output2 = collection.GetOutputByAddress(node1Output1Addr)
	require.NotNil(t, output2, "Failed getting created output")

	// expect an updated output
	updated := collection.GetUpdatedOutputs(true)
	require.Equal(t, 1, len(updated), "Expected 1 updated output")

	//
	updated = collection.GetUpdatedOutputs(false)
	require.Equal(t, 0, len(updated), "Expected no more updated outputs")

	outs := collection.GetAllOutputs()
	assert.Equal(t, 1, len(outs), "Expected 1 output")

	nodeOuts := collection.GetNodeOutputs(node1Addr)
	assert.Equal(t, 1, len(nodeOuts), "Expected 1 output")

	nodeOuts = collection.GetOutputsByNode(node1ID)
	assert.Equal(t, 1, len(nodeOuts), "Expected 1 output")
}

func TestUpdateOutputs(t *testing.T) {
	collection := outputs.NewRegisteredOutputs(domain, publisher1ID)
	output1 := collection.CreateOutput(node1ID, types.OutputTypeSwitch, types.DefaultOutputInstance)
	if !assert.NotNil(t, output1, "Failed creating output") {
		return
	}
	// expect 1 updated output
	updated := collection.GetUpdatedOutputs(true)
	if !(assert.Equal(t, 1, len(updated), "Expected 1 updated output")) {
		return
	}
	// update
	collection.UpdateOutput(output1)
	updated = collection.GetUpdatedOutputs(false)
	if !(assert.Equal(t, 1, len(updated), "Expected 1 updated output")) {
		return
	}
}
