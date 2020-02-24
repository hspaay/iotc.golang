// Package publisher with message publication functions
package publisher

import (
	"iotzone/nodes"
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
	aliasConfig := node.Config[nodes.AttrNameAlias]
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
func (publisher *ThisPublisherState) publishEventCommand(aliasAddress string, node *nodes.Node) {
	aliasSegments := strings.Split(aliasAddress, "/")
	aliasSegments[3] = EventCommand
	addr := strings.Join(aliasSegments[:4], "/")
	publisher.Logger.Infof("publish node event: %s", addr)

	outputs := publisher.getNodeOutputs(node)
	event := make(map[string]string)
	timeStampStr := time.Now().Format("2006-01-02T15:04:05.000-0700")
	for _, output := range outputs {
		history := nodes.GetHistory(output)
		attrID := output.IOType + "/" + output.Instance
		event[attrID] = history[0].Value
	}
	eventMessage := &nodes.EventMessage{
		Address:   addr,
		Event:     event,
		Sender:    publisher.publisherNode.Address,
		Timestamp: timeStampStr,
	}
	publisher.messenger.Publish(addr, eventMessage)
}

// publish the $latest output value
func (publisher *ThisPublisherState) publishLatestCommand(aliasAddress string, output *nodes.InOutput) {
	aliasSegments := strings.Split(aliasAddress, "/")
	aliasSegments[3] = LatestCommand
	addr := strings.Join(aliasSegments, "/")

	// zone/publisher/node/$latest/iotype/instance
	history := nodes.GetHistory(output)
	latest := history[0]
	publisher.Logger.Infof("publish output latest: %s", addr)
	latestMessage := &nodes.LatestMessage{
		Address:   addr,
		Sender:    publisher.publisherNode.Address,
		Timestamp: latest.TimeStamp,
		Unit:      output.Unit,
		Value:     latest.Value,
	}
	publisher.messenger.Publish(addr, latestMessage)
}

// publish the $history output values
func (publisher *ThisPublisherState) publishHistoryCommand(aliasAddress string, output *nodes.InOutput) {
	aliasSegments := strings.Split(aliasAddress, "/")
	aliasSegments[3] = HistoryCommand
	addr := strings.Join(aliasSegments, "/")
	timeStampStr := time.Now().Format("2006-01-02T15:04:05.000-0700")

	historyMessage := &nodes.HistoryMessage{
		Address:   addr,
		Duration:  0, // tbd
		Sender:    publisher.publisherNode.Address,
		Timestamp: timeStampStr,
		Unit:      output.Unit,
		History:   nodes.GetHistory(output),
	}
	publisher.messenger.Publish(addr, historyMessage)
}

// publish the raw output $value
func (publisher *ThisPublisherState) publishValueCommand(aliasAddress string, output *nodes.InOutput) {
	aliasSegments := strings.Split(aliasAddress, "/")

	// publish raw value with the $value command
	// zone/publisher/node/$value/iotype/instance
	history := nodes.GetHistory(output)
	latest := history[0]
	aliasSegments[3] = ValueCommand
	alias := strings.Join(aliasSegments, "/")
	publisher.Logger.Infof("publish output value: %s", aliasAddress)

	publisher.messenger.PublishRaw(alias, latest.Value) // raw
}

// Publish discovery and value updates onto the message bus
// The order is nodes first, followed by in/outputs, followed by values.
//   the sequence within nodes, in/outputs, and values does not follow the discovery sequence
//   as the map used to record updates is unordered.
// This method is not thread safe and should only be used in a locked section
func (publisher *ThisPublisherState) publishUpdates() {
	// publish changes to nodes
	if publisher.messenger == nil {
		return // can't do anything here, just go home
	}
	// publish updated nodes
	if publisher.updatedNodes != nil {
		for addr, node := range publisher.updatedNodes {
			publisher.Logger.Infof("publish node discovery: %s", addr)
			publisher.messenger.Publish(addr, node)
		}
		publisher.updatedNodes = nil
	}

	// publish updated inputs or outputs
	if publisher.updatedInOutputs != nil {
		for addr, inoutput := range publisher.updatedInOutputs {
			aliasAddress := publisher.getAliasAddress(addr)
			publisher.Logger.Infof("publish in/output discovery: %s", aliasAddress)
			publisher.messenger.Publish(aliasAddress, inoutput)
		}
		publisher.updatedInOutputs = nil
	}
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
