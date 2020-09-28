package inputs_test

import (
	"testing"
	"time"

	"github.com/iotdomain/iotdomain-go/inputs"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/stretchr/testify/assert"
)

func TestCreateInputFromHttp(t *testing.T) {
	const domain = "test"
	const publisher1ID = "pub1"
	const node1ID = "node1"
	const inputType = types.InputTypeImage
	const instance1 = "1"
	const instance2 = "2"
	const imageUrl = "https://images.drivebc.ca/bchighwaycam/pub/cameras/598.jpg" // pemberton, bc
	const badUrl = "https://localhost/test"
	const login = "test"
	const password = ""
	const interval = 10
	var inputReceived = ""

	handler := func(input *types.InputDiscoveryMessage, sender string, value string) {
		inputReceived = value
	}
	regInputs := inputs.NewRegisteredInputs(domain, publisher1ID)

	i := inputs.NewReceiveFromHTTP(regInputs)
	i.Start()

	input1 := i.CreateHTTPInput(node1ID, inputType, instance1, imageUrl, login, password, interval, handler)
	assert.NotNil(t, input1, "Input1 not created")

	input2 := i.CreateHTTPInput(node1ID, inputType, instance2, badUrl, "", "", interval, handler)
	assert.NotEmpty(t, input2, "Input2 not created")

	inputList := regInputs.GetAllInputs()
	assert.Equal(t, 2, len(inputList), "Deleting http input doesn't seem to work")

	// wait 2 seconds for the poll loop to query the url
	time.Sleep(2 * time.Second)
	assert.NotEmpty(t, inputReceived, "No input received")

	// deleting input
	i.DeleteInput(input1.InputID)
	inputList = regInputs.GetAllInputs()
	assert.Equal(t, 1, len(inputList), "Deleting http input doesn't seem to work")

	// delete an already deleted input should not fail nor blow up
	i.DeleteInput(input1.InputID)

	i.Stop()
}
