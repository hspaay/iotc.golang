// Package publisher with updating and publishing of node outputs
package publisher

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/hspaay/iotconnect.golang/messenger"
	"github.com/hspaay/iotconnect.golang/standard"
)

// GetForecast returns the output's forecast list
// Returns nil if the type or instance is unknown or no forecast is available
func (publisher *PublisherState) GetForecast(inOutput *standard.InOutput) standard.HistoryList {
	publisher.updateMutex.Lock()
	var forecastList = publisher.outputForecast[inOutput.Address]
	publisher.updateMutex.Unlock()
	return forecastList
}

// PublishUpdates publishes updated nodes, inputs and outputs
func (publisher *PublisherState) PublishUpdates() {
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
func (publisher *PublisherState) UpdateForecast(node *standard.Node, outputType string, outputInstance string, forecast standard.HistoryList) {
	addr := standard.MakeOutputDiscoveryAddress(node.Zone, node.PublisherID, node.ID, outputType, outputInstance)

	publisher.updateMutex.Lock()
	publisher.outputForecast[addr] = forecast
	output := publisher.Outputs.GetOutputByAddress(addr)
	aliasAddress := publisher.getOutputAliasAddress(addr)
	publisher.publishForecastCommand(aliasAddress, output)
	publisher.updateMutex.Unlock()
}

// UpdateOutputStringList adds a list of strings as the output value in the format: "[value1, value2, ...]"
func (publisher *PublisherState) UpdateOutputStringList(node *standard.Node, outputType string, outputInstance string, values []string) {
	valuesAsString, _ := json.Marshal(values)
	publisher.OutputHistory.UpdateOutputValue(node, outputType, outputInstance, string(valuesAsString))
}

// UpdateOutputFloatList adds a list of floats as the output value in the format: "[value1, value2, ...]"
func (publisher *PublisherState) UpdateOutputFloatList(node *standard.Node, outputType string, outputInstance string, values []float32) {
	valuesAsString, _ := json.Marshal(values)
	publisher.OutputHistory.UpdateOutputValue(node, outputType, outputInstance, string(valuesAsString))
}

// UpdateOutputIntList adds a list of integers as the output value in the format: "[value1, value2, ...]"
func (publisher *PublisherState) UpdateOutputIntList(node *standard.Node, outputType string, outputInstance string, values []int) {
	valuesAsString, _ := json.Marshal(values)
	if publisher.OutputHistory.UpdateOutputValue(node, outputType, outputInstance, string(valuesAsString)) {
		publisher.publishOutputValues()
	}
}

// Replace the address with the node's alias instead the node ID, if available
// return the address if the node doesn't have an alias
// This method is not thread safe and should only be used in a locked section
func (publisher *PublisherState) getOutputAliasAddress(address string) string {
	node := publisher.Nodes.GetNodeByAddress(address)
	if node == nil {
		return address
	}
	aliasConfig, configExists := node.Config[standard.AttrNameAlias]
	if !configExists || (aliasConfig.Value == "") {
		return address
	}
	parts := strings.Split(address, "/")
	parts[2] = aliasConfig.Value
	aliasAddr := strings.Join(parts, "/")
	return aliasAddr
}

// publish all node output values in the $event command
// zone/publisher/node/$event
// TODO: decide when to invoke this
func (publisher *PublisherState) publishEventCommand(aliasAddress string, node *standard.Node) {
	aliasSegments := strings.Split(aliasAddress, "/")
	aliasSegments[3] = standard.CommandEvent
	addr := strings.Join(aliasSegments[:4], "/")
	publisher.Logger.Infof("publish node event: %s", addr)

	outputs := publisher.Outputs.GetNodeOutputs(node)
	event := make(map[string]string)
	timeStampStr := time.Now().Format("2006-01-02T15:04:05.000-0700")
	for _, output := range outputs {
		latest := publisher.OutputHistory.GetOutputValueByAddress(output.Address)
		attrID := output.IOType + "/" + output.Instance
		event[attrID] = latest.Value
	}
	eventMessage := &standard.EventMessage{
		Address:   addr,
		Event:     event,
		Sender:    publisher.PublisherNode.Address,
		Timestamp: timeStampStr,
	}
	publisher.publishMessage(addr, true, eventMessage)
}

// publish the $latest output value
// not thread-safe, using within a locked section
func (publisher *PublisherState) publishLatestCommand(aliasAddress string, output *standard.InOutput) {
	aliasSegments := strings.Split(aliasAddress, "/")
	aliasSegments[3] = standard.CommandLatest
	addr := strings.Join(aliasSegments, "/")

	// zone/publisher/node/$latest/iotype/instance
	latest := publisher.OutputHistory.GetOutputValueByAddress(output.Address)
	if latest == nil {
		publisher.Logger.Warningf("publishLatest, no latest value. This is unexpected")
		return
	}
	publisher.Logger.Infof("publish output latest: %s", addr)
	latestMessage := &standard.LatestMessage{
		Address:   addr,
		Sender:    publisher.PublisherNode.Address,
		Timestamp: latest.Timestamp.Format("2006-01-02T15:04:05.000-0700"),
		// Timestamp: latest.TimeStamp,
		Unit:  output.Unit,
		Value: latest.Value,
	}
	publisher.publishMessage(addr, true, latestMessage)
}

// publish the $forecast output values retained=true
// not thread-safe, using within a locked section
func (publisher *PublisherState) publishForecastCommand(aliasAddress string, output *standard.InOutput) {
	aliasSegments := strings.Split(aliasAddress, "/")
	aliasSegments[3] = standard.CommandForecast
	addr := strings.Join(aliasSegments, "/")
	timeStampStr := time.Now().Format("2006-01-02T15:04:05.000-0700")

	forecastMessage := &standard.ForecastMessage{
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
func (publisher *PublisherState) publishHistoryCommand(aliasAddress string, output *standard.InOutput) {
	aliasSegments := strings.Split(aliasAddress, "/")
	aliasSegments[3] = standard.CommandHistory
	addr := strings.Join(aliasSegments, "/")
	timeStampStr := time.Now().Format("2006-01-02T15:04:05.000-0700")

	historyMessage := &standard.HistoryMessage{
		Address:   addr,
		Duration:  0, // tbd
		Sender:    publisher.PublisherNode.Address,
		Timestamp: timeStampStr,
		Unit:      output.Unit,
		History:   publisher.OutputHistory.GetHistory(output.Address),
	}
	publisher.publishMessage(addr, true, historyMessage)
}

// publishMessage encapsulates the message object in a payload, signs, and sends it
// not thread-safe, using within a locked section
// address of the publication
// object to publish. This will be marshalled to JSON and signed by this publisher
func (publisher *PublisherState) publishMessage(address string, retained bool, object interface{}) {
	buffer, err := json.MarshalIndent(object, " ", " ")
	if err != nil {
		publisher.Logger.Errorf("Error marshalling message for address %s: %s", address, err)
		return
	}
	signature := standard.CreateEcdsaSignature(buffer, publisher.signPrivateKey)

	publication := &messenger.Publication{
		Message:   buffer,
		Signature: signature,
	}
	publisher.messenger.Publish(address, retained, publication)
}

// publish the raw output $value (retained)
// not thread-safe, using within a locked section
func (publisher *PublisherState) publishValueCommand(aliasAddress string, output *standard.InOutput) {
	aliasSegments := strings.Split(aliasAddress, "/")

	// publish raw value with the $value command
	// zone/publisher/node/$value/iotype/instance
	latest := publisher.OutputHistory.GetOutputValueByAddress(output.Address)
	if latest == nil {
		publisher.Logger.Warningf("publishValue, no latest value. This is unexpected")
		return
	}
	aliasSegments[3] = standard.CommandValue
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
func (publisher *PublisherState) publishOutputValues() {
	// publish updated output values using alias address if configured
	addressesOfUpdatedOutputs := publisher.OutputHistory.GetUpdatedOutputs(true)
	for _, addr := range addressesOfUpdatedOutputs {
		aliasAddress := publisher.getOutputAliasAddress(addr)
		output := publisher.Outputs.GetOutputByAddress(addr)
		publisher.publishValueCommand(aliasAddress, output)
		publisher.publishLatestCommand(aliasAddress, output)
		publisher.publishHistoryCommand(aliasAddress, output)
	}
}
