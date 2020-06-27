package publisher

import (
	"errors"
	"fmt"
	"time"

	"github.com/iotdomain/iotdomain-go/types"
)

// PublishUpdatedOutputValues publishes pending updates to output values
func (publisher *Publisher) PublishUpdatedOutputValues() {
	// publish updated output values using alias address if configured
	addressesOfUpdatedOutputs := publisher.OutputValues.GetUpdatedOutputs(true)

	for _, outputAddress := range addressesOfUpdatedOutputs {
		output := publisher.Outputs.GetOutputByAddress(outputAddress)
		unit := output.Unit
		publisher.publishRawValue(outputAddress)
		publisher.publishLatest(outputAddress, unit)
		publisher.publishHistory(outputAddress, unit)
	}
	addressesOfUpdatedForecasts := publisher.OutputForecasts.GetUpdatedForecasts(true)
	for _, outputAddress := range addressesOfUpdatedForecasts {
		output := publisher.Outputs.GetOutputByAddress(outputAddress)
		unit := output.Unit
		publisher.publishForecast(outputAddress, unit)
	}
}

// publish all node output values in the $event command
// zone/publisher/nodealias/$event
// TODO: decide when to invoke this
func (publisher *Publisher) publishEvent(node *types.NodeDiscoveryMessage) {
	// output values are published using their alias address, if any
	aliasAddress := publisher.getOutputAliasAddress(node.Address, types.MessageTypeEvent)
	publisher.logger.Infof("Publisher.publishEvent: %s", aliasAddress)

	outputs := publisher.Outputs.GetNodeOutputs(node.Address)
	event := make(map[string]string)
	timeStampStr := time.Now().Format("2006-01-02T15:04:05.000-0700")
	for _, output := range outputs {
		latest := publisher.OutputValues.GetOutputValueByAddress(output.Address)
		attrID := string(output.OutputType) + "/" + output.Instance
		event[attrID] = latest.Value
	}
	eventMessage := &types.OutputEventMessage{
		Address:   aliasAddress,
		Event:     event,
		Timestamp: timeStampStr,
	}
	publisher.publishObject(aliasAddress, true, eventMessage, nil)
}

// publish the $forecast output values retained=true
// not thread-safe, using within a locked section
func (publisher *Publisher) publishForecast(outputAddress string, unit types.Unit) {
	// output values are published using their alias address, if any
	aliasAddress := publisher.getOutputAliasAddress(outputAddress, types.MessageTypeForecast)
	timeStampStr := time.Now().Format("2006-01-02T15:04:05.000-0700")
	forecast := publisher.OutputForecasts.GetForecast(outputAddress)

	forecastMessage := &types.OutputForecastMessage{
		Address:   aliasAddress,
		Duration:  0, // tbd
		Timestamp: timeStampStr,
		Unit:      unit,
		Forecast:  forecast,
	}
	publisher.logger.Debugf("Publisher.publishForecast: %d entries on %s", len(forecastMessage.Forecast), aliasAddress)
	publisher.publishObject(aliasAddress, true, forecastMessage, nil)
}

// publish the $history output values retained=true
// not thread-safe, using within a locked section
func (publisher *Publisher) publishHistory(outputAddress string, unit types.Unit) {
	// output values are published using their alias address, if any
	aliasAddress := publisher.getOutputAliasAddress(outputAddress, types.MessageTypeHistory)
	timeStampStr := time.Now().Format("2006-01-02T15:04:05.000-0700")
	history := publisher.OutputValues.GetHistory(outputAddress)

	historyMessage := &types.OutputHistoryMessage{
		Address:   aliasAddress,
		Duration:  0, // tbd
		Timestamp: timeStampStr,
		Unit:      unit,
		History:   history,
	}
	publisher.logger.Debugf("Publisher.publishHistory: %d entries on %s", len(historyMessage.History), aliasAddress)
	publisher.publishObject(aliasAddress, true, historyMessage, nil)
}

// publish the $latest output value
// not thread-safe, using within a locked section
func (publisher *Publisher) publishLatest(outputAddress string, unit types.Unit) {
	// output values are published using their alias address, if any
	aliasAddress := publisher.getOutputAliasAddress(outputAddress, types.MessageTypeLatest)

	// zone/publisher/node/iotype/instance/$latest
	latest := publisher.OutputValues.GetOutputValueByAddress(outputAddress)
	if latest == nil {
		publisher.logger.Warningf("Publisher.publishLatest: no latest value. This is unexpected")
		return
	}
	publisher.logger.Infof("Publisher.publishLatest: %s", aliasAddress)
	latestMessage := &types.OutputLatestMessage{
		Address:   aliasAddress,
		Timestamp: latest.Timestamp,
		// Timestamp: latest.TimeStamp,
		Unit:  unit,
		Value: latest.Value,
	}
	publisher.publishObject(aliasAddress, true, latestMessage, nil)
}

// publishRawValue to the raw output $raw (retained)
// not thread-safe, using within a locked section
func (publisher *Publisher) publishRawValue(outputAddress string) error {

	// output values are published using their alias address, if any
	aliasAddress := publisher.getOutputAliasAddress(outputAddress, types.MessageTypeRaw)

	// publish raw value with the $raw command
	latest := publisher.OutputValues.GetOutputValueByAddress(outputAddress)
	if latest == nil {
		errMsg := fmt.Sprintf("Publisher.publishRawValue:, no latest value for %s. This is unexpected", outputAddress)
		publisher.logger.Error(errMsg)
		return errors.New(errMsg)
	}
	s := latest.Value
	// don't send full images ???
	// if len(s) > 30 {
	// 	s = s[:30]
	// }
	publisher.logger.Infof("Publisher.publishRawValue: output value '%s' on %s", s, aliasAddress)

	err := publisher.publishSigned(aliasAddress, true, s)
	return err
}
