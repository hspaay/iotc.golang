package outputs_test

import (
	"crypto/ecdsa"
	"testing"

	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/iotdomain/iotdomain-go/outputs"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/stretchr/testify/assert"
)

func TestCreateForecastValues(t *testing.T) {
	const domain = "test"
	const publisher1ID = "publisher1"
	const node1ID = "node1"
	const output1Type = types.OutputTypeTemperature
	// out1addr := outputs.MakeOutputDiscoveryAddress(domain, publisher1ID, node1ID, types.OutputTypeSwitch, types.DefaultOutputInstance)

	collection := outputs.NewRegisteredForecastValues(domain, publisher1ID)
	assert.NotNil(t, collection)

	output1ID := outputs.MakeOutputID(node1ID, output1Type, types.DefaultOutputInstance)

	fc := collection.GetForecast(output1ID)
	assert.Nil(t, fc)
	fcList := collection.GetUpdatedForecasts(false)
	assert.NotNil(t, fcList)

	forecast := make([]types.OutputValue, 0)
	collection.UpdateForecast(output1ID, forecast)

	fcList = collection.GetUpdatedForecasts(true)
	assert.Equal(t, 1, len(fcList), "Expect one updated forecast")

}

func TestPublishForecast(t *testing.T) {
	const domain = "test"
	const publisher1ID = "publisher1"
	const node1ID = "node1"
	const output1Type = types.OutputTypeTemperature

	var privKey = messaging.CreateAsymKeys()

	// get publisher key for signature verification
	var getPublisherKey = func(addr string) *ecdsa.PublicKey {
		return &privKey.PublicKey
	}

	msgr := messaging.NewDummyMessenger(nil)
	signer := messaging.NewMessageSigner(msgr, privKey, getPublisherKey)
	assert.NotNil(t, signer)

	collection := outputs.NewRegisteredForecastValues(domain, publisher1ID)
	regOutputs := outputs.NewRegisteredOutputs(domain, publisher1ID)
	assert.NotNil(t, collection)
	// output := collection.CreateOutput(node1ID, types.OutputTypeSwitch, types.DefaultOutputInstance)

	// no forecasts
	outputs.PublishUpdatedForecasts(collection, regOutputs, signer)

	// one forecast
	output1 := regOutputs.CreateOutput(node1ID, output1Type, types.DefaultOutputInstance)
	forecast := make([]types.OutputValue, 0)
	collection.UpdateForecast(output1.OutputID, forecast)

	outputs.PublishUpdatedForecasts(collection, regOutputs, signer)

	// TODO: verify the forecast was published
}
