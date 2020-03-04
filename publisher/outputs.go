// Package publisher with updating and publishing of node outputs
package publisher

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hspaay/iotconnect.golang/messenger"
	"github.com/hspaay/iotconnect.golang/standard"
)

// GetHistory returns the history list
// Returns nil if the type or instance is unknown
func (publisher *PublisherState) GetHistory(inOutput *standard.InOutput) standard.HistoryList {
	publisher.updateMutex.Lock()
	var historyList = publisher.outputHistory[inOutput.Address]
	publisher.updateMutex.Unlock()
	return historyList
}

// GetOutput returns the output of one of this publisher's nodes
// Returns nil if the type or instance is unknown
func (publisher *PublisherState) GetOutput(
	node *standard.Node, outputType string, instance string) *standard.InOutput {

	outputAddr := fmt.Sprintf("%s/%s/%s/%s/%s/%s", node.Zone, node.PublisherID, node.ID,
		standard.CommandOutputDiscovery, outputType, instance)
	publisher.updateMutex.Lock()
	var output = publisher.outputs[outputAddr]
	publisher.updateMutex.Unlock()
	return output
}

// GetOutputValue returns the current output value
func (publisher *PublisherState) GetOutputValue(node *standard.Node, outputType string, instance string) string {
	output := publisher.GetOutput(node, outputType, instance)
	publisher.updateMutex.Lock()
	latest := publisher.getLatestOutput(output)
	publisher.updateMutex.Unlock()
	if latest == nil {
		return ""
	}
	return latest.Value
}

// UpdateOutputValue adds the new output value to the front of the history
// If the output has a repeatDelay configured, then the value is only added if
//  it has changed or the previous update was older than the repeatDelay.
// The history retains a max of 24 hours
func (publisher *PublisherState) UpdateOutputValue(node *standard.Node, outputType string, outputInstance string, newValue string) {
	var previous *standard.HistoryValue
	var repeatDelay = 3600 // default repeat delay
	var ageSeconds = -1

	addr := standard.MakeOutputDiscoveryAddress(node.Zone, node.PublisherID, node.ID, outputType, outputInstance)

	publisher.updateMutex.Lock()
	// auto create the output if it hasn't been discovered yet
	output := publisher.outputs[addr]
	history := publisher.outputHistory[addr]

	// only update output if value changes or delay has passed
	if node.RepeatDelay != 0 {
		repeatDelay = node.RepeatDelay
	}
	if len(history) > 0 {
		previous = history[0]
		age := time.Now().Sub(previous.Time)
		ageSeconds = int(age.Seconds())
	}

	doUpdate := ageSeconds < 0 || ageSeconds > repeatDelay || newValue != previous.Value
	if doUpdate {
		newHistory := updateHistory(history, newValue, node.HistorySize)

		publisher.outputHistory[addr] = newHistory
		if publisher.updatedOutputValues == nil {
			publisher.updatedOutputValues = make(map[string]*standard.InOutput)
		}
		publisher.updatedOutputValues[output.Address] = output

		if publisher.synchroneous {
			publisher.publishOutputValues()
		}
	}
	publisher.updateMutex.Unlock()
}

// Get the latest output historyvalue
// This is not thread-safe. Use within a locked area
func (publisher *PublisherState) getLatestOutput(inoutput *standard.InOutput) *standard.HistoryValue {
	history := publisher.outputHistory[inoutput.Address]
	if history == nil || len(history) == 0 {
		return nil
	}
	return history[0]
}

// Replace the address with the node's alias instead the node ID, if available
// return the address if the node doesn't have an alias
// This method is not thread safe and should only be used in a locked section
func (publisher *PublisherState) getOutputAliasAddress(address string) string {
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
func (publisher *PublisherState) publishEventCommand(aliasAddress string, node *standard.Node) {
	aliasSegments := strings.Split(aliasAddress, "/")
	aliasSegments[3] = standard.CommandEvent
	addr := strings.Join(aliasSegments[:4], "/")
	publisher.Logger.Infof("publish node event: %s", addr)

	outputs := publisher.getNodeOutputs(node)
	event := make(map[string]string)
	timeStampStr := time.Now().Format("2006-01-02T15:04:05.000-0700")
	for _, output := range outputs {
		history := publisher.outputHistory[output.Address]
		attrID := output.IOType + "/" + output.Instance
		event[attrID] = history[0].Value
	}
	eventMessage := &standard.EventMessage{
		Address:   addr,
		Event:     event,
		Sender:    publisher.publisherNode.Address,
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
	latest := publisher.getLatestOutput(output)
	if latest == nil {
		publisher.Logger.Warningf("publishLatest, no latest value. This is unexpected")
		return
	}
	publisher.Logger.Infof("publish output latest: %s", addr)
	latestMessage := &standard.LatestMessage{
		Address:   addr,
		Sender:    publisher.publisherNode.Address,
		Timestamp: latest.TimeStamp,
		Unit:      output.Unit,
		Value:     latest.Value,
	}
	publisher.publishMessage(addr, true, latestMessage)
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
		Sender:    publisher.publisherNode.Address,
		Timestamp: timeStampStr,
		Unit:      output.Unit,
		History:   publisher.outputHistory[output.Address],
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
	latest := publisher.getLatestOutput(output)
	if latest == nil {
		publisher.Logger.Warningf("publishValue, no latest value. This is unexpected")
		return
	}
	aliasSegments[3] = standard.CommandValue
	alias := strings.Join(aliasSegments, "/")
	s := latest.Value
	if len(s) > 30 {
		s = s[:30]
	}
	publisher.Logger.Infof("publish output value '%s' on %s", s, aliasAddress)

	publisher.messenger.PublishRaw(alias, true, []byte(latest.Value)) // raw
}

// publishOutputValues publishes pending updates to output values
// not thread-safe, using within a locked section
func (publisher *PublisherState) publishOutputValues() {
	// publish updated output values using alias address if configured
	if publisher.updatedOutputValues != nil {
		for addr, output := range publisher.updatedOutputValues {
			aliasAddress := publisher.getOutputAliasAddress(addr)
			publisher.publishValueCommand(aliasAddress, output)
			publisher.publishLatestCommand(aliasAddress, output)
			publisher.publishHistoryCommand(aliasAddress, output)
		}
		publisher.updatedOutputValues = nil
	}
}

// updateHistory inserts a new value at the front of the history
// The resulting list contains a max of historySize entries limited to 24 hours
// This function is not thread-safe and should only be used from within a locked section
// history is optional and used to insert the value in the front. If nil then a new history is returned
// maxHistorySize is optional and limits the size in addition to the 24 hour limit
// returns the history list with the new value at the front of the list
func updateHistory(history standard.HistoryList, newValue string, maxHistorySize int) standard.HistoryList {

	timeStamp := time.Now()
	timeStampStr := timeStamp.Format("2006-01-02T15:04:05.000-0700")

	latest := standard.HistoryValue{
		Time:      timeStamp,
		TimeStamp: timeStampStr,
		Value:     newValue,
	}
	if history == nil {
		history = make(standard.HistoryList, 1)
	} else {
		copy(history[1:], history[0:])
	}
	history[0] = &latest

	// remove old entries, determine the max
	if maxHistorySize == 0 || len(history) < maxHistorySize {
		maxHistorySize = len(history)
	}
	// cap at 24 hours
	for ; maxHistorySize > 1; maxHistorySize-- {
		entry := history[maxHistorySize-1]
		if timeStamp.Sub(entry.Time) <= time.Hour*24 {
			break
		}
	}
	history = history[0:maxHistorySize]
	return history
}
