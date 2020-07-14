package inputs_test

import (
	"crypto/ecdsa"
	"fmt"
	"testing"

	"github.com/iotdomain/iotdomain-go/inputs"
	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var privKey = messaging.CreateAsymKeys()

func getPublisherKey(addr string) *ecdsa.PublicKey {
	return &privKey.PublicKey
}

func TestPublishSetInput(t *testing.T) {
	const domain = "test"
	const publisher1ID = "publisher1"
	const node1ID = "node1"
	const input1Type = types.InputTypeTemperature
	const message1Content = "55"
	var setInput1Addr = inputs.MakeSetInputAddress(domain, publisher1ID, node1ID, input1Type, types.DefaultInputInstance)
	var input1Addr = inputs.MakeInputDiscoveryAddress(domain, publisher1ID, node1ID, input1Type, types.DefaultInputInstance)
	var senderAddr = fmt.Sprintf("%s/publisher1/node2/$node", domain)
	var receivedInputs map[string]*types.SetInputMessage = make(map[string]*types.SetInputMessage)

	msgr := messaging.NewDummyMessenger(nil)
	signer := messaging.NewMessageSigner(true, getPublisherKey, msgr, privKey)

	inputHandler := func(address string, msg *types.SetInputMessage) {
		logrus.Printf("Received set message for input %s", address)
		receivedInputs[address] = msg
	}

	registeredInputs := inputs.NewRegisteredInputs(domain, publisher1ID)
	receiver := inputs.NewReceiveSetInput(domain, publisher1ID, inputHandler,
		signer, registeredInputs, privKey, getPublisherKey)

	// publish the encrypted set input message for node 1 temperature
	receiver.Start()
	setMsg := types.SetInputMessage{
		Address: input1Addr, Value: message1Content, Sender: senderAddr,
	}
	signer.PublishObject(setInput1Addr, false, setMsg, &privKey.PublicKey)
	receiver.Stop()

	// expect to have received input
	rxMsg := receivedInputs[input1Addr]
	require.NotNil(t, rxMsg, "Did not receive the set input message on %s", input1Addr)
	assert.Equal(t, message1Content, rxMsg.Value, "Set message content doesnt match")
}
