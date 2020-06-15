// Package publisher with updating and publishing of node outputs
package publisher

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hspaay/iotc.golang/iotc"
	"github.com/hspaay/iotc.golang/messenger"
	"github.com/hspaay/iotc.golang/persist"
)

// PublishIdentity publishes this publisher's identity on startup or update
func (publisher *Publisher) PublishIdentity() {
	identity := publisher.identity
	publisher.logger.Infof("Publisher.PublishIdentity: publish identity: %s", publisher.identity.Address)
	publisher.publishObject(identity.Address, true, identity)
}

// PublishUpdatedDiscoveries publishes updated nodes, inputs and outputs discovery messages
// If updates are available then nodes are saved
func (publisher *Publisher) PublishUpdatedDiscoveries() {
	if publisher.messenger == nil {
		publisher.logger.Error("Publisher.PublishUpdates: No messenger")
		return // can't do anything here, just go home
	}
	publisher.updateMutex.Lock()
	nodeList := publisher.Nodes.GetUpdatedNodes(true)
	inputList := publisher.Inputs.GetUpdatedInputs(true)
	outputList := publisher.Outputs.GetUpdatedOutputs(true)
	publisher.updateMutex.Unlock()

	// publish updated nodes
	for _, node := range nodeList {
		publisher.logger.Infof("Publisher.PublishUpdates: publish node discovery: %s", node.Address)
		publisher.publishObject(node.Address, true, node)
	}
	if len(nodeList) > 0 && publisher.autosaveFolder != "" {
		allNodes := publisher.Nodes.GetAllNodes()
		persist.SaveNodes(publisher.autosaveFolder, publisher.publisherID, allNodes)
	}

	// publish updated input discovery
	for _, input := range inputList {
		aliasAddress := publisher.getOutputAliasAddress(input.Address, "")
		publisher.logger.Infof("Publisher.PublishUpdates: publish input discovery: %s", aliasAddress)
		publisher.publishObject(aliasAddress, true, input)
	}
	if len(inputList) > 0 && publisher.autosaveFolder != "" {
		allInputs := publisher.Inputs.GetAllInputs()
		persist.SaveInputs(publisher.autosaveFolder, publisher.publisherID, allInputs)
	}

	// publish updated output discovery
	for _, output := range outputList {
		aliasAddress := publisher.getOutputAliasAddress(output.Address, "")
		publisher.logger.Infof("Publisher.PublishUpdates: publish output discovery: %s", aliasAddress)
		publisher.publishObject(aliasAddress, true, output)
	}
	if len(outputList) > 0 && publisher.autosaveFolder != "" {
		allOutputs := publisher.Outputs.GetAllOutputs()
		persist.SaveOutputs(publisher.autosaveFolder, publisher.publisherID, allOutputs)
	}
}

// PublishUpdatedOutputValues publishes pending updates to output values
// not thread-safe, using within a locked section
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

// Replace the address with the node's alias instead the node ID, and the message type with the given
//  message type for publication.
// If the node doesn't have an alias then its nodeId will be kept.
// messageType to substitute in the address. Use "" to keep the original message type (usually discovery message)
func (publisher *Publisher) getOutputAliasAddress(address string, messageType iotc.MessageType) string {
	node := publisher.Nodes.GetNodeByAddress(address)
	if node == nil {
		return address
	}
	alias, hasAlias := publisher.Nodes.GetNodeConfigValue(address, iotc.NodeAttrAlias)
	// alias, hasAlias := nodes.GetNodeAlias(node)
	// zone/pub/node/outtype/instance/messagetype
	parts := strings.Split(address, "/")
	if !hasAlias || alias == "" {
		alias = parts[2]
	}
	parts[2] = alias
	if messageType != "" {
		parts[5] = string(messageType)
	}
	aliasAddr := strings.Join(parts, "/")
	return aliasAddr
}

// publish all node output values in the $event command
// zone/publisher/nodealias/$event
// TODO: decide when to invoke this
func (publisher *Publisher) publishEvent(node *iotc.NodeDiscoveryMessage) {
	// output values are published using their alias address, if any
	aliasAddress := publisher.getOutputAliasAddress(node.Address, iotc.MessageTypeEvent)
	publisher.logger.Infof("Publisher.publishEvent: %s", aliasAddress)

	outputs := publisher.Outputs.GetNodeOutputs(node)
	event := make(map[string]string)
	timeStampStr := time.Now().Format("2006-01-02T15:04:05.000-0700")
	for _, output := range outputs {
		latest := publisher.OutputValues.GetOutputValueByAddress(output.Address)
		attrID := output.OutputType + "/" + output.Instance
		event[attrID] = latest.Value
	}
	eventMessage := &iotc.OutputEventMessage{
		Address:   aliasAddress,
		Event:     event,
		Timestamp: timeStampStr,
	}
	publisher.publishObject(aliasAddress, true, eventMessage)
}

// publish the $latest output value
// not thread-safe, using within a locked section
func (publisher *Publisher) publishLatest(outputAddress string, unit iotc.Unit) {
	// output values are published using their alias address, if any
	aliasAddress := publisher.getOutputAliasAddress(outputAddress, iotc.MessageTypeLatest)

	// zone/publisher/node/iotype/instance/$latest
	latest := publisher.OutputValues.GetOutputValueByAddress(outputAddress)
	if latest == nil {
		publisher.logger.Warningf("Publisher.publishLatest: no latest value. This is unexpected")
		return
	}
	publisher.logger.Infof("Publisher.publishLatest: %s", aliasAddress)
	latestMessage := &iotc.OutputLatestMessage{
		Address:   aliasAddress,
		Timestamp: latest.Timestamp,
		// Timestamp: latest.TimeStamp,
		Unit:  unit,
		Value: latest.Value,
	}
	publisher.publishObject(aliasAddress, true, latestMessage)
}

// publish the $forecast output values retained=true
// not thread-safe, using within a locked section
func (publisher *Publisher) publishForecast(outputAddress string, unit iotc.Unit) {
	// output values are published using their alias address, if any
	aliasAddress := publisher.getOutputAliasAddress(outputAddress, iotc.MessageTypeForecast)
	timeStampStr := time.Now().Format("2006-01-02T15:04:05.000-0700")
	forecast := publisher.OutputForecasts.GetForecast(outputAddress)

	forecastMessage := &iotc.OutputForecastMessage{
		Address:   aliasAddress,
		Duration:  0, // tbd
		Timestamp: timeStampStr,
		Unit:      unit,
		Forecast:  forecast,
	}
	publisher.logger.Debugf("Publisher.publishForecast: %d entries on %s", len(forecastMessage.Forecast), aliasAddress)
	publisher.publishObject(aliasAddress, true, forecastMessage)
}

// publish the $history output values retained=true
// not thread-safe, using within a locked section
func (publisher *Publisher) publishHistory(outputAddress string, unit iotc.Unit) {
	// output values are published using their alias address, if any
	aliasAddress := publisher.getOutputAliasAddress(outputAddress, iotc.MessageTypeHistory)
	timeStampStr := time.Now().Format("2006-01-02T15:04:05.000-0700")
	history := publisher.OutputValues.GetHistory(outputAddress)

	historyMessage := &iotc.OutputHistoryMessage{
		Address:   aliasAddress,
		Duration:  0, // tbd
		Timestamp: timeStampStr,
		Unit:      unit,
		History:   history,
	}
	publisher.logger.Debugf("Publisher.publishHistory: %d entries on %s", len(historyMessage.History), aliasAddress)
	publisher.publishObject(aliasAddress, true, historyMessage)
}

// publishObject encapsulates the message object in a payload, signs the message, and sends it.
// address of the publication
// object to publish. This will be marshalled to JSON and signed by this publisher
func (publisher *Publisher) publishObject(address string, retained bool, object interface{}) error {
	payload, err := json.Marshal(object)
	// buffer, err := json.MarshalIndent(object, " ", " ")
	if err != nil {
		publisher.logger.Errorf("Publisher.publishMessage: Error marshalling message for address %s: %s", address, err)
		return err
	}
	err = publisher.publishPayload(address, retained, string(payload))
	return err
}

// publishRawValue to the raw output $raw (retained)
// not thread-safe, using within a locked section
func (publisher *Publisher) publishRawValue(outputAddress string) error {

	// output values are published using their alias address, if any
	aliasAddress := publisher.getOutputAliasAddress(outputAddress, iotc.MessageTypeRaw)

	// publish raw value with the $value command
	// zone/publisher/node/$value/iotype/instance
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

	err := publisher.publishPayload(aliasAddress, true, s)
	return err
}

// publishMessage sign the payload and publish the resulting message
// address of the publication
// payload to publish
func (publisher *Publisher) publishPayload(address string, retained bool, payload string) error {
	var err error

	// default is unsigned
	message := payload

	if publisher.signingMethod == SigningMethodJWS {
		message, err = messenger.CreateJWSSignature(string(payload), publisher.privateKeySigning)
		if err != nil {
			publisher.logger.Errorf("Publisher.publishMessage: Error signing message for address %s: %s", address, err)
		}
	}
	err = publisher.messenger.Publish(address, retained, message)
	return err
}
