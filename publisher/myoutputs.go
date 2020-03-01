// Package publisher with publication of my node outputs
package publisher

import (
	"encoding/json"
	"iotconnect/messenger"
	"iotconnect/standard"
	"strings"
	"time"
)

// Replace the address with the node's alias instead the node ID, if available
// return the address if the node doesn't have an alias
// This method is not thread safe and should only be used in a locked section
func (publisher *ThisPublisherState) getAliasAddress(address string) string {
	node := publisher.getNode(address)
	if node == nil {
		return address
	}
	aliasConfig := node.Config[standard.AttrNameAlias]
	if (aliasConfig == nil) || (aliasConfig.Value == "") {
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
func (publisher *ThisPublisherState) publishEventCommand(aliasAddress string, node *standard.Node) {
	aliasSegments := strings.Split(aliasAddress, "/")
	aliasSegments[3] = standard.CommandEvent
	addr := strings.Join(aliasSegments[:4], "/")
	publisher.Logger.Infof("publish node event: %s", addr)

	outputs := publisher.getNodeOutputs(node)
	event := make(map[string]string)
	timeStampStr := time.Now().Format("2006-01-02T15:04:05.000-0700")
	for _, output := range outputs {
		history := standard.GetHistory(output)
		attrID := output.IOType + "/" + output.Instance
		event[attrID] = history[0].Value
	}
	eventMessage := &standard.EventMessage{
		Address:   addr,
		Event:     event,
		Sender:    publisher.publisherNode.Address,
		Timestamp: timeStampStr,
	}
	publisher.publishMessage(addr, eventMessage)
}

// publish the $latest output value
func (publisher *ThisPublisherState) publishLatestCommand(aliasAddress string, output *standard.InOutput) {
	aliasSegments := strings.Split(aliasAddress, "/")
	aliasSegments[3] = standard.CommandLatest
	addr := strings.Join(aliasSegments, "/")

	// zone/publisher/node/$latest/iotype/instance
	history := standard.GetHistory(output)
	latest := history[0]
	publisher.Logger.Infof("publish output latest: %s", addr)
	latestMessage := &standard.LatestMessage{
		Address:   addr,
		Sender:    publisher.publisherNode.Address,
		Timestamp: latest.TimeStamp,
		Unit:      output.Unit,
		Value:     latest.Value,
	}
	publisher.publishMessage(addr, latestMessage)
}

// publish the $history output values
func (publisher *ThisPublisherState) publishHistoryCommand(aliasAddress string, output *standard.InOutput) {
	aliasSegments := strings.Split(aliasAddress, "/")
	aliasSegments[3] = standard.CommandHistory
	addr := strings.Join(aliasSegments, "/")
	timeStampStr := time.Now().Format("2006-01-02T15:04:05.000-0700")

	historyMessage := &standard.HistoryMessage{
		Address:   addr,
		Duration:  0, // tbd
		Sender:    publisher.publisherNode.Address,
		Timestamp: timeStampStr,
		Unit:      output.Unit,
		History:   standard.GetHistory(output),
	}
	publisher.publishMessage(addr, historyMessage)
}

// publishMessage encapsulates the message object in a payload, signs, and sends it
// address of the publication
// object to publish. This will be marshalled to JSON and signed by this publisher
func (publisher *ThisPublisherState) publishMessage(address string, object interface{}) {
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
	publisher.messenger.Publish(address, publication)
}

// publish the raw output $value
func (publisher *ThisPublisherState) publishValueCommand(aliasAddress string, output *standard.InOutput) {
	aliasSegments := strings.Split(aliasAddress, "/")

	// publish raw value with the $value command
	// zone/publisher/node/$value/iotype/instance
	history := standard.GetHistory(output)
	latest := history[0]
	aliasSegments[3] = standard.CommandValue
	alias := strings.Join(aliasSegments, "/")
	s := latest.Value
	if len(s) > 30 {
		s = s[:30]
	}
	publisher.Logger.Infof("publish output value '%s' on %s", s, aliasAddress)

	publisher.messenger.PublishRaw(alias, []byte(latest.Value)) // raw
}

// publishOutputValues publishes pending updates to output values
func (publisher *ThisPublisherState) publishOutputValues() {
	// publish updated output values using alias address if configured
	if publisher.updatedOutputValues != nil {
		for addr, output := range publisher.updatedOutputValues {
			aliasAddress := publisher.getAliasAddress(addr)
			publisher.publishValueCommand(aliasAddress, output)
			publisher.publishLatestCommand(aliasAddress, output)
			publisher.publishHistoryCommand(aliasAddress, output)
		}
		publisher.updatedOutputValues = nil
	}
}
