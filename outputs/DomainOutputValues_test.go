package outputs_test

import (
	"crypto/ecdsa"
	"fmt"
	"testing"

	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/iotdomain/iotdomain-go/outputs"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/stretchr/testify/assert"
)

func TestCreateDomainOutputValues(t *testing.T) {
	const domain = "test"
	const publisherID = "pub1"
	const node1ID = "node1"
	const out1Type = types.OutputTypeSwitch
	var node1Base = fmt.Sprintf("%s/%s/%s", domain, publisherID, node1ID)
	var out1Addr = fmt.Sprintf("%s/%s/%s", node1Base, out1Type, types.DefaultOutputInstance)
	// var out1Addr = outputs.MakeOutputDiscoveryAddress(
	// 	domain, publisherID, node1ID, out1Type, types.DefaultOutputInstance)
	// }
	config := messaging.MessengerConfig{}
	messenger := messaging.NewDummyMessenger(&config)
	privKey := messaging.CreateAsymKeys()
	getPubKey := func(address string) *ecdsa.PublicKey {
		return &privKey.PublicKey
	}
	signer := messaging.NewMessageSigner(true, getPubKey, messenger, privKey)

	collection := outputs.NewDomainOutputValues(signer)
	assert.NotNil(t, collection)

	collection.GetRaw(out1Addr)
	collection.GetLatest(out1Addr)
	collection.UpdateEvent(&types.OutputEventMessage{})
	collection.UpdateHistory(&types.OutputHistoryMessage{})
	collection.UpdateLatest(&types.OutputLatestMessage{})
	collection.UpdateRaw(out1Addr, "raw")
}
