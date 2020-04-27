// Package publisher with updating and publishing of node outputs
package publisher

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/hspaay/iotconnect.golang/messaging"
	"github.com/hspaay/iotconnect.golang/messenger"
	"github.com/hspaay/iotconnect.golang/nodes"
)

// GetForecast returns the output's forecast list
// Returns nil if the type or instance is unknown or no forecast is available
func (publisher *Publisher) GetForecast(output *nodes.Output) messaging.OutputHistoryList {
	publisher.updateMutex.Lock()
	var forecastList = publisher.outputForecast[output.Address]
	publisher.updateMutex.Unlock()
	return forecastList
}

// PublishUpdates publishes updated nodes, inputs and outputs
func (publisher *Publisher) PublishUpdates() {
	if publisher.messenger == nil {
		publisher.Logger.Error("PublishUpdates: No messenger")
		return // can't do anything here, just go home
	}
	publisher.updateMutex.Lock()
	nodeList := publisher.Nodes.GetUpdatedNodes(true)
	inputList := publisher.Inputs.GetUpdatedInputs(true)
	outputList := publisher.Outputs.GetUpdatedOutputs(true)
	publisher.updateMutex.Unlock()

	// publish updated nodes
	for _, node := range nodeList {
		publisher.Logger.Infof("publish node discovery: %s", node.Address)
		publisher.publishMessage(node.Address, true, node)
	}
	// publish updated input or output discovery
	for _, input := range inputList {
		aliasAddress := publisher.getOutputAliasAddress(input.Address)
		publisher.Logger.Infof("publish input discovery: %s", aliasAddress)
		publisher.publishMessage(aliasAddress, true, input)
	}
	for _, output := range outputList {
		aliasAddress := publisher.getOutputAliasAddress(output.Address)
		publisher.Logger.Infof("publish output discovery: %s", aliasAddress)
		publisher.publishMessage(aliasAddress, true, output)
	}
}

// UpdateForecast publishes the output forecast list of values"
func (publisher *Publisher) UpdateForecast(node *nodes.Node, outputType string, outputInstance string, forecast messaging.OutputHistoryList) {
	addr := nodes.MakeOutputDiscoveryAddress(node.Zone, node.PublisherID, node.ID, outputType, outputInstance)

	publisher.updateMutex.Lock()
	publisher.outputForecast[addr] = forecast
	output := publisher.Outputs.GetOutputByAddress(addr)
	aliasAddress := publisher.getOutputAliasAddress(addr)
	publisher.publishForecast(aliasAddress, output)
	publisher.updateMutex.Unlock()
}

// Replace the address with the node's alias instead the node ID, if available
// return the address if the node doesn't have an alias
// This method is not thread safe and should only be used in a locked section
func (publisher *Publisher) getOutputAliasAddress(address string) string {
	node := publisher.Nodes.GetNodeByAddress(address)
	if node == nil {
		return address
	}
	alias, hasAlias := node.GetAlias()
	if !hasAlias {
		return address
	}
	parts := strings.Split(address, "/")
	parts[2] = alias
	aliasAddr := strings.Join(parts, "/")
	return aliasAddr
}

// publish all node output values in the $event command
// zone/publisher/node/$event
// TODO: decide when to invoke this
func (publisher *Publisher) publishEvent(aliasAddress string, node *nodes.Node) {
	aliasSegments := strings.Split(aliasAddress, "/")
	aliasSegments[3] = messaging.MessageTypeEvent
	addr := strings.Join(aliasSegments[:4], "/")
	publisher.Logger.Infof("publish node event: %s", addr)

	outputs := publisher.Outputs.GetNodeOutputs(node)
	event := make(map[string]string)
	timeStampStr := time.Now().Format("2006-01-02T15:04:05.000-0700")
	for _, output := range outputs {
		latest := publisher.OutputValues.GetOutputValueByAddress(output.Address)
		attrID := output.OutputType + "/" + output.Instance
		event[attrID] = latest.Value
	}
	eventMessage := &messaging.OutputEventMessage{
		Address:   addr,
		Event:     event,
		Sender:    publisher.PublisherNode.Address,
		Timestamp: timeStampStr,
	}
	publisher.publishMessage(addr, true, eventMessage)
}

// publish the $latest output value
// not thread-safe, using within a locked section
func (publisher *Publisher) publishLatest(aliasAddress string, output *nodes.Output) {
	aliasSegments := strings.Split(aliasAddress, "/")
	aliasSegments[3] = messaging.MessageTypeLatest
	addr := strings.Join(aliasSegments, "/")

	// zone/publisher/node/$latest/iotype/instance
	latest := publisher.OutputValues.GetOutputValueByAddress(output.Address)
	if latest == nil {
		publisher.Logger.Warningf("publishLatest, no latest value. This is unexpected")
		return
	}
	publisher.Logger.Infof("publish output latest: %s", addr)
	latestMessage := &messaging.OutputLatestMessage{
		Address:   addr,
		Sender:    publisher.PublisherNode.Address,
		Timestamp: latest.Timestamp,
		// Timestamp: latest.TimeStamp,
		Unit:  output.Unit,
		Value: latest.Value,
	}
	publisher.publishMessage(addr, true, latestMessage)
}

// publish the $forecast output values retained=true
// not thread-safe, using within a locked section
func (publisher *Publisher) publishForecast(aliasAddress string, output *nodes.Output) {
	aliasSegments := strings.Split(aliasAddress, "/")
	aliasSegments[3] = messaging.MessageTypeForecast
	addr := strings.Join(aliasSegments, "/")
	timeStampStr := time.Now().Format("2006-01-02T15:04:05.000-0700")

	forecastMessage := &messaging.OutputForecastMessage{
		Address:   addr,
		Duration:  0, // tbd
		Sender:    publisher.PublisherNode.Address,
		Timestamp: timeStampStr,
		Unit:      output.Unit,
		Forecast:  publisher.outputForecast[output.Address],
	}
	publisher.publishMessage(addr, true, forecastMessage)
}

// publish the $history output values retained=true
// not thread-safe, using within a locked section
func (publisher *Publisher) publishHistory(aliasAddress string, output *nodes.Output) {
	aliasSegments := strings.Split(aliasAddress, "/")
	aliasSegments[3] = messaging.MessageTypeHistory
	addr := strings.Join(aliasSegments, "/")
	timeStampStr := time.Now().Format("2006-01-02T15:04:05.000-0700")

	historyMessage := &messaging.OutputHistoryMessage{
		Address:   addr,
		Duration:  0, // tbd
		Sender:    publisher.PublisherNode.Address,
		Timestamp: timeStampStr,
		Unit:      output.Unit,
		History:   publisher.OutputValues.GetHistory(output.Address),
	}
	publisher.publishMessage(addr, true, historyMessage)
}

// publishMessage encapsulates the message object in a payload, signs, and sends it
// not thread-safe, using within a locked section
// address of the publication
// object to publish. This will be marshalled to JSON and signed by this publisher
func (publisher *Publisher) publishMessage(address string, retained bool, object interface{}) {
	buffer, err := json.MarshalIndent(object, " ", " ")
	if err != nil {
		publisher.Logger.Errorf("Error marshalling message for address %s: %s", address, err)
		return
	}
	signature := messenger.CreateEcdsaSignature(buffer, publisher.signPrivateKey)

	publication := &messaging.Publication{
		Message:   buffer,
		Signature: signature,
	}
	publisher.messenger.Publish(address, retained, publication)
}

// publish the raw output $value (retained)
// not thread-safe, using within a locked section
func (publisher *Publisher) publishValueCommand(aliasAddress string, output *nodes.Output) {
	aliasSegments := strings.Split(aliasAddress, "/")

	// publish raw value with the $value command
	// zone/publisher/node/$value/iotype/instance
	latest := publisher.OutputValues.GetOutputValueByAddress(output.Address)
	if latest == nil {
		publisher.Logger.Warningf("publishValue, no latest value. This is unexpected")
		return
	}
	aliasSegments[3] = messaging.MessageTypeValue
	addr := strings.Join(aliasSegments, "/")
	s := latest.Value
	if len(s) > 30 {
		s = s[:30]
	}
	publisher.Logger.Infof("publish output value '%s' on %s", s, addr)

	publisher.messenger.PublishRaw(addr, true, []byte(latest.Value)) // raw
}

// publishOutputValues publishes pending updates to output values
// not thread-safe, using within a locked section
func (publisher *Publisher) publishOutputValues() {
	// publish updated output values using alias address if configured
	addressesOfUpdatedOutputs := publisher.OutputValues.GetUpdatedOutputs(true)
	for _, addr := range addressesOfUpdatedOutputs {
		aliasAddress := publisher.getOutputAliasAddress(addr)
		output := publisher.Outputs.GetOutputByAddress(addr)
		publisher.publishValueCommand(aliasAddress, output)
		publisher.publishLatest(aliasAddress, output)
		publisher.publishHistory(aliasAddress, output)
	}
}
