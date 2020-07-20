package inputs_test

import (
	"crypto/ecdsa"
	"fmt"
	"testing"
	"time"

	"github.com/iotdomain/iotdomain-go/inputs"
	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var privKey = messaging.CreateAsymKeys()

// get publisher key for signature verification
func getPublisherKey(addr string) *ecdsa.PublicKey {
	return &privKey.PublicKey
}

func TestCreateSetInput(t *testing.T) {
	const input1Type = types.InputTypeTemperature

	msgr := messaging.NewDummyMessenger(nil)
	signer := messaging.NewMessageSigner(true, getPublisherKey, msgr, privKey)
	registeredInputs := inputs.NewRegisteredInputs(domain, publisher1ID)

	receiver := inputs.NewInputFromSetCommands(domain, publisher1ID,
		signer, registeredInputs, privKey)

	receiver.CreateInput(node1ID, input1Type, types.DefaultInputInstance, nil)
	receiver.DeleteInput(node1ID, input1Type, types.DefaultInputInstance)

}

func TestPublishSetInput(t *testing.T) {
	const domain = "test"
	const publisher1ID = "publisher1"
	const node1ID = "node1"
	const input1Type = types.InputTypeTemperature
	var setInput1Addr = inputs.MakeSetInputAddress(domain, publisher1ID, node1ID, input1Type, types.DefaultInputInstance)
	var input1Addr = inputs.MakeInputDiscoveryAddress(domain, publisher1ID, node1ID, input1Type, types.DefaultInputInstance)
	var senderAddr = fmt.Sprintf("%s/publisher1/node2/$node", domain)
	var receivedInputs map[string]string = make(map[string]string)
	var signatureVerificationKey = &privKey.PublicKey

	msgr := messaging.NewDummyMessenger(nil)
	signer := messaging.NewMessageSigner(true, func(addr string) *ecdsa.PublicKey {
		return signatureVerificationKey
	}, msgr, privKey)

	inputHandler := func(address string, sender string, value string) {
		logrus.Printf("Received set message for input %s from %s", address, sender)
		receivedInputs[address] = value
	}

	registeredInputs := inputs.NewRegisteredInputs(domain, publisher1ID)
	// the receiver registers the inputs and listens for set commands
	receiver := inputs.NewInputFromSetCommands(domain, publisher1ID, signer, registeredInputs, privKey)

	// publish the encrypted set input message for node 1 temperature
	receiver.CreateInput(node1ID, input1Type, types.DefaultInputInstance, inputHandler)

	setMsgOld := types.SetInputMessage{
		Address: input1Addr, Value: "content old", Sender: senderAddr,
		Timestamp: time.Now().Format(types.TimeFormat),
	}
	time.Sleep(time.Millisecond)
	setMsg := types.SetInputMessage{
		Address: input1Addr, Value: "content1", Sender: senderAddr,
		Timestamp: time.Now().Format(types.TimeFormat),
	}
	// without encryption the message is rejected
	signer.PublishObject(setInput1Addr, false, setMsg, nil)
	rxMsg := receivedInputs[input1Addr]
	assert.NotEqual(t, "content1", rxMsg, "non encrypted message should not be accepted")

	// with the wrong private encryption key the message is rejected
	wrongKey := &messaging.CreateAsymKeys().PublicKey
	signer.PublishObject(setInput1Addr, false, setMsg, wrongKey)
	rxMsg = receivedInputs[input1Addr]
	assert.NotEqual(t, "content1", rxMsg, "non encrypted message should not be accepted")

	// with the wrong signature encryption key the message is rejected
	signatureVerificationKey = &messaging.CreateAsymKeys().PublicKey
	signer.PublishObject(setInput1Addr, false, setMsg, &privKey.PublicKey)
	rxMsg = receivedInputs[input1Addr]
	assert.NotEqual(t, "content1", rxMsg, "message with failed signature verification should not be accepted")
	signatureVerificationKey = &privKey.PublicKey

	// unsigned message should be reject
	signer.SetSignMessages(false)
	signer.PublishObject(setInput1Addr, false, setMsg, &privKey.PublicKey)
	rxMsg = receivedInputs[input1Addr]
	signer.SetSignMessages(true)
	assert.NotEqual(t, "content1", rxMsg, "message with failed signature verification should not be accepted")

	// with encryption expect to have received input
	inputs.PublishSetInput(setInput1Addr, setMsg.Value, setMsg.Sender, signer, &privKey.PublicKey)
	// signer.PublishObject(setInput1Addr, false, setMsg, &privKey.PublicKey)
	rxMsg = receivedInputs[input1Addr]
	assert.Equal(t, "content1", rxMsg, "Set message content doesnt match")

	// older message should be rejected - protect against replay attack
	signer.PublishObject(setInput1Addr, false, setMsgOld, &privKey.PublicKey)
	rxMsg = receivedInputs[input1Addr]
	assert.NotEqual(t, "content old", rxMsg, "Older message should not be accepted")

}
