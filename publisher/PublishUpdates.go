// Package publisher with publication of updates of registered entities
package publisher

import (
	"time"

	"github.com/iotdomain/iotdomain-go/inputs"
	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/iotdomain/iotdomain-go/nodes"
	"github.com/iotdomain/iotdomain-go/outputs"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/sirupsen/logrus"
)

// PublishUpdates publishes changes to registered nodes, inputs, outputs, values and this publisher identity
func (publisher *Publisher) PublishUpdates() {

	updatedNodes := publisher.registeredNodes.GetUpdatedNodes(true)
	nodes.PublishRegisteredNodes(updatedNodes, publisher.messageSigner)

	updatedInputs := publisher.registeredInputs.GetUpdatedInputs(true)
	inputs.PublishRegisteredInputs(updatedInputs, publisher.messageSigner)

	updatedOutputs := publisher.registeredOutputs.GetUpdatedOutputs(true)
	outputs.PublishRegisteredOutputs(updatedOutputs, publisher.messageSigner)

	updatedValues := publisher.registeredOutputValues.GetUpdatedOutputValues(true)
	publisher.PublishUpdatedOutputValues(updatedValues, publisher.messageSigner)
}

// PublishUpdatedOutputValues publishes updated outputs discovery and values of registered outputs
// This uses the node config to determine which output publications to use: eg raw, latest, history
func (publisher *Publisher) PublishUpdatedOutputValues(
	updatedOutputValueAddresses []string,
	messageSigner *messaging.MessageSigner) {
	regOutputValues := publisher.registeredOutputValues

	for _, outputAddress := range updatedOutputValueAddresses {
		output := publisher.registeredOutputs.GetOutputByAddress(outputAddress)
		node := publisher.registeredNodes.GetNodeByAddress(outputAddress)

		latestValue := regOutputValues.GetOutputValueByAddress(outputAddress)
		if latestValue == nil {
			logrus.Warningf("PublishOutputValues: no latest value for %s. This is unexpected", output.Address)
		} else {
			pubRaw, _ := publisher.registeredNodes.GetNodeConfigBool(node.Address, types.NodeAttrPublishRaw, true)
			if pubRaw {
				outputs.PublishOutputRaw(output, latestValue.Value, messageSigner)
			}
			pubLatest, _ := publisher.registeredNodes.GetNodeConfigBool(node.Address, types.NodeAttrPublishLatest, true)
			if pubLatest {
				outputs.PublishOutputLatest(output, latestValue, messageSigner)
			}
			pubHistory, _ := publisher.registeredNodes.GetNodeConfigBool(node.Address, types.NodeAttrPublishHistory, true)
			if pubHistory {
				history := regOutputValues.GetHistory(outputAddress)
				outputs.PublishOutputHistory(output, history, messageSigner)
			}
			pubEvent, _ := publisher.registeredNodes.GetNodeConfigBool(node.Address, types.NodeAttrPublishEvent, false)
			if pubEvent {
				PublishOutputEvent(node, publisher.registeredOutputs, publisher.registeredOutputValues, messageSigner)
			}
		}
	}
}

// PublishOutputEvent publishes all node output values in the $event command
// zone/publisher/nodealias/$event
// TODO: decide when to invoke this
func PublishOutputEvent(
	node *types.NodeDiscoveryMessage,
	registeredOutputs *outputs.RegisteredOutputs,
	outputValues *outputs.RegisteredOutputValues,
	messageSigner *messaging.MessageSigner,
) {
	// output values are published using their alias address, if any
	aliasAddress := outputs.ReplaceMessageType(node.Address, types.MessageTypeEvent)
	logrus.Infof("Publisher.publishEvent: %s", aliasAddress)

	nodeOutputs := registeredOutputs.GetOutputsByDeviceID(node.Address)
	event := make(map[string]string)
	timeStampStr := time.Now().Format("2006-01-02T15:04:05.000-0700")
	for _, output := range nodeOutputs {
		latest := outputValues.GetOutputValueByAddress(output.Address)
		attrID := string(output.OutputType) + "/" + output.Instance
		event[attrID] = latest.Value
	}
	eventMessage := &types.OutputEventMessage{
		Address:   aliasAddress,
		Event:     event,
		Timestamp: timeStampStr,
	}
	messageSigner.PublishObject(aliasAddress, true, eventMessage, nil)
}
