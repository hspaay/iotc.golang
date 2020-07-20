package inputs_test

import (
	"crypto/ecdsa"
	"testing"
	"time"

	"github.com/iotdomain/iotdomain-go/inputs"
	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/square/go-jose/json"
	"github.com/stretchr/testify/assert"
)

func TestCreateInputFromOutputs(t *testing.T) {
	const domain = "test"
	const publisher1ID = "pub1"
	const node1ID = "node1"
	const inputType = types.InputTypeImage
	const instance = types.DefaultInputInstance
	const outputAddrRaw = "test/pub1/node1/image/0/" + types.MessageTypeRaw       // pemberton, bc
	const outputAddrLatest = "test/pub1/node1/image/0/" + types.MessageTypeLatest // pemberton, bc
	var inputReceived = ""
	var privKey = messaging.CreateAsymKeys()
	var signatureVerificationKey = &privKey.PublicKey

	handler := func(addr string, sender string, value string) {
		inputReceived = value
	}
	msgr := messaging.NewDummyMessenger(nil)
	signer := messaging.NewMessageSigner(true, func(addr string) *ecdsa.PublicKey {
		return signatureVerificationKey
	}, msgr, privKey)
	regInputs := inputs.NewRegisteredInputs(domain, publisher1ID)

	i := inputs.NewInputFromOutputs(signer, regInputs)

	addr1 := i.CreateInput(node1ID, inputType, instance, outputAddrRaw, handler)
	assert.NotEmpty(t, addr1, "No input address")
	// todo: add input that fails and an input that responds with error

	// receive raw output
	inputList := regInputs.GetAllInputs()
	msgr.Publish(outputAddrRaw, false, "Hello")
	assert.NotEmpty(t, inputReceived, "No input received")

	// receive latest output
	i.CreateInput(node1ID, inputType, "latest", outputAddrLatest, handler)
	latest := types.OutputLatestMessage{
		Address:   outputAddrLatest,
		Value:     "World",
		Timestamp: time.Now().Format(types.TimeFormat),
	}
	payload, _ := json.Marshal(&latest)
	signer.PublishSigned(outputAddrLatest, false, string(payload))
	time.Sleep(2 * time.Second)
	assert.Equal(t, "World", inputReceived, "No input received")

	// older timestamp
	latest.Timestamp = (time.Now().Add(-time.Hour)).Format(types.TimeFormat)
	latest.Value = "Older"
	payload, _ = json.Marshal(&latest)
	signer.PublishSigned(outputAddrLatest, false, string(payload))
	time.Sleep(2 * time.Second)
	assert.Equal(t, "World", inputReceived, "No input received")

	// fail signing
	signatureVerificationKey = &messaging.CreateAsymKeys().PublicKey
	signer.PublishSigned(outputAddrLatest, false, string(payload))
	time.Sleep(2 * time.Second)
	// assert.Equal(t, "World", inputReceived, "No input received")

	// delete
	i.DeleteInput(node1ID, inputType, instance)
	inputList = regInputs.GetAllInputs()
	assert.Equal(t, 1, len(inputList), "Deleting input doesn't seem to work")

	// delete non existing input should not fail
	i.DeleteInput(node1ID, inputType, instance)

}
