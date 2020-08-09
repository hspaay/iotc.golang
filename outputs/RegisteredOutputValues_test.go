package outputs_test

import (
	"crypto/ecdsa"
	"testing"

	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/iotdomain/iotdomain-go/outputs"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateOutputValues(t *testing.T) {
	const domain = "test"
	const publisher1ID = "publisher1"
	const node1ID = "node1"
	value1 := "Hello"
	collection := outputs.NewRegisteredOutputValues(domain, publisher1ID)

	// output := collection.NewOutput(node1ID, types.OutputTypeSwitch, types.DefaultOutputInstance)
	// add a value and get the result
	outputID := outputs.MakeOutputID(node1ID, types.OutputTypeSwitch, types.DefaultOutputInstance)

	res := collection.UpdateOutputValue(outputID, value1)
	assert.True(t, true, res, "Output update not returning 'updated' status")
	val1 := collection.GetOutputValueByID(outputID)
	assert.NotNil(t, val1)

	val1 = collection.GetOutputValueByType(node1ID, types.OutputTypeSwitch, types.DefaultInputInstance)
	require.NotNil(t, val1, "Failed getting output value")

	updates := collection.GetUpdatedOutputValues(true)
	assert.Equal(t, 1, len(updates))

	// no output should return nil
	val2 := collection.GetOutputValueByID("not an output")
	assert.Nil(t, val2)

	// Add another a value
	collection.UpdateOutputValue(outputID, "World")
	history := collection.GetHistory(outputID)
	assert.Equal(t, 2, len(history), "expected 2 output values in history")

	// Add a list of ints, floats and strings
	out3ID := outputs.MakeOutputID(node1ID, types.OutputTypeSwitch, "floatList")
	floatList := []float32{1.1, 2, 3}
	collection.UpdateOutputFloatList(out3ID, floatList)
	val3 := collection.GetOutputValueByID(out3ID)
	assert.Equal(t, val3.Value, "[1.1,2,3]")
	intList := []int{1, 2, 3}
	collection.UpdateOutputIntList(out3ID, intList)
	val3 = collection.GetOutputValueByID(out3ID)
	assert.Equal(t, val3.Value, "[1,2,3]")
	stringList := []string{"a", "b", "c"}
	collection.UpdateOutputStringList(out3ID, stringList)
	val3 = collection.GetOutputValueByID(out3ID)
	assert.Equal(t, val3.Value, "[\"a\",\"b\",\"c\"]")
}

func TestPublishOutputValues(t *testing.T) {
	const domain = "test"
	const publisher1ID = "publisher1"
	const node1ID = "node1"
	output1Type := types.OutputTypeSwitch
	output1 := outputs.NewOutput(domain, publisher1ID, node1ID, output1Type, types.DefaultOutputInstance)
	collection := outputs.NewRegisteredOutputValues(domain, publisher1ID)

	config := messaging.MessengerConfig{}
	messenger := messaging.NewDummyMessenger(&config)
	privKey := messaging.CreateAsymKeys()
	getPubKey := func(address string) *ecdsa.PublicKey {
		return &privKey.PublicKey
	}
	signer := messaging.NewMessageSigner(messenger, privKey, getPubKey)

	collection.UpdateOutputValue(output1.OutputID, "World.")

	history := collection.GetHistory(output1.OutputID)
	outputs.PublishOutputHistory(output1, history, signer)

	latest := collection.GetOutputValueByID(output1.OutputID)
	outputs.PublishOutputLatest(output1, latest, signer)

	outputs.PublishOutputRaw(output1, "The terms anno Domini (AD) and"+
		" before Christ (BC)[note 1] are used to label or number years in the Julian "+
		"and Gregorian calendars.", signer)

}
