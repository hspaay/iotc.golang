package inputs_test

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/iotdomain/iotdomain-go/inputs"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/stretchr/testify/assert"
)

func TestReceiveFromFile(t *testing.T) {
	const deviceID = "node1"
	const inputType = types.InputTypeImage
	const instance = types.DefaultInputInstance
	const instance2 = "secondinstance"
	const instance3 = "notafile"
	const testFile = "../test/testImage.jpg"
	const test2File = "../test/testImage.jpg"
	const test3File = "~/test/doesntexist.jpg"
	var fileTouched = ""

	handler := func(input *types.InputDiscoveryMessage, sender string, file string) {
		fileTouched = file
	}
	regInputs := inputs.NewRegisteredInputs(domain, publisher1ID)

	iff := inputs.NewReceiveFromFiles(regInputs)
	iff.Start()

	addr1 := iff.CreateInput(deviceID, inputType, instance, testFile, handler)
	assert.NotEmpty(t, addr1, "No input address")

	// adding twice should return the same address
	addr1b := iff.CreateInput(deviceID, inputType, instance, testFile, handler)
	assert.Equal(t, addr1, addr1b, "Different address second add")

	// second input on the same file
	addr2 := iff.CreateInput(deviceID, inputType, instance2, test2File, handler)
	assert.NotEmpty(t, addr2, "Failed with two inputs on same file")

	// invalid file
	addr3 := iff.CreateInput(deviceID, inputType, instance3, test3File, handler)
	assert.Empty(t, addr3, "No error when watching non existing file")

	// trigger handler on change
	err := ioutil.WriteFile(testFile, []byte("Hello World"), 0644)
	time.Sleep(time.Second)
	assert.NoError(t, err, "Unexpected problem touching test file")
	assert.NotEmpty(t, fileTouched, "Handler not called when touching file")

	// no more trigger after deleting input
	iff.DeleteInput(deviceID, inputType, instance)
	fileTouched = ""
	ioutil.WriteFile(testFile, []byte("Hello World again"), 0644)
	time.Sleep(time.Second)
	assert.Empty(t, fileTouched, "Handler not called when touching file")
	input := regInputs.GetInputByDevice(deviceID, inputType, instance)
	assert.Nil(t, input, "Deleted input is still there")

	// error cases
	// - delete input with empty source
	input = regInputs.GetInputByDevice(deviceID, inputType, instance2)
	input.Source = ""
	iff.DeleteInput(deviceID, inputType, instance2)
	input = regInputs.GetInputByDevice(deviceID, inputType, instance2)
	assert.Nil(t, input, "Deleted input2 is still there")
	// - delete non existing input (it was just deleted)
	iff.DeleteInput(deviceID, inputType, instance2)

	// delete last input
	iff.DeleteInput(deviceID, inputType, instance3)

	iff.Stop()

}
