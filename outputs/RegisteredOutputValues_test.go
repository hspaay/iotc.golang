package outputs_test

import (
	"testing"

	"github.com/iotdomain/iotdomain-go/outputs"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateOutputValues(t *testing.T) {
	value1 := "Hello"
	collection := outputs.NewRegisteredOutputValues(domain, publisher1ID)

	// output := collection.NewOutput(node1ID, types.OutputTypeSwitch, types.DefaultOutputInstance)
	out1addr := outputs.MakeOutputDiscoveryAddress(domain, publisher1ID, node1ID, types.OutputTypeSwitch, types.DefaultOutputInstance)
	res := collection.UpdateOutputValue(out1addr, value1)
	assert.True(t, true, res, "Output update not returning 'updated' status")

	val1 := collection.GetOutputValueByType(node1ID, types.OutputTypeSwitch, types.DefaultInputInstance)
	require.NotNil(t, val1, "Failed getting output value")

	updates := collection.GetUpdatedOutputValues(true)
	assert.Equal(t, 1, len(updates))

	val2 := collection.GetOutputValueByAddress(out1addr)
	assert.NotNil(t, val2)

	collection.UpdateOutputValue(out1addr, "World")

	history := collection.GetHistory(out1addr)
	assert.NotNil(t, history)
}
