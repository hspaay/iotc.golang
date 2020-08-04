package outputs

import (
	"time"

	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/sirupsen/logrus"
)

// PublishForecast publishes the $forecast output values retained=true
// not thread-safe, using within a locked section
func PublishForecast(
	output *types.OutputDiscoveryMessage,
	forecast OutputForecast,
	messageSigner *messaging.MessageSigner,
) {

	aliasAddress := ReplaceMessageType(output.Address, types.MessageTypeForecast)
	timeStampStr := time.Now().Format("2006-01-02T15:04:05.000-0700")

	forecastMessage := &types.OutputForecastMessage{
		Address:   aliasAddress,
		Duration:  0, // tbd
		Timestamp: timeStampStr,
		Unit:      output.Unit,
		Forecast:  forecast,
	}
	logrus.Debugf("Publisher.publishForecast: %d entries on %s", len(forecastMessage.Forecast), aliasAddress)
	messageSigner.PublishObject(aliasAddress, true, forecastMessage, nil)
}

// PublishUpdatedForecasts publishes the output forecasts
// While every output has a history, forecasts are only available for outputs that are able to
// provide a prediction. For example a weather forecast. This is therefore a separate collection
func PublishUpdatedForecasts(
	regFCValues *RegisteredForecastValues,
	regOutputs *RegisteredOutputs,
	messageSigner *messaging.MessageSigner) {

	for _, outputID := range regFCValues.GetUpdatedForecasts(true) {
		output := regOutputs.GetOutputByID(outputID)
		forecast := regFCValues.GetForecast(outputID)

		PublishForecast(output, forecast, messageSigner)
	}
}
