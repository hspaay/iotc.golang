// Package publisher with message publication functions
package publisher

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/asn1"
	"encoding/base64"
	"encoding/json"
	"iotzone/messenger"
	"iotzone/nodes"
	"math/big"
	"strings"
	"time"
)

// ECDSASignature ...
type ECDSASignature struct {
	R, S *big.Int
}

// ecdsaSign the message and return the base64 encoded signature
// This requires the signing private key to be set
func (publisher *ThisPublisherState) ecdsaSign(message []byte) string {
	if publisher.signPrivateKey == nil {
		return ""
	}
	hashed := sha256.Sum256(message)
	r, s, err := ecdsa.Sign(rand.Reader, publisher.signPrivateKey, hashed[:])
	if err != nil {
		return ""
	}
	sig, err := asn1.Marshal(ECDSASignature{r, s})
	return base64.StdEncoding.EncodeToString(sig)
}

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
	eventMessage := &EventMessage{
		Address:   addr,
		Event:     event,
		Sender:    publisher.publisherNode.Address,
		Timestamp: timeStampStr,
	}
	publisher.publishMessage(addr, eventMessage)
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
	latestMessage := &LatestMessage{
		Address:   addr,
		Sender:    publisher.publisherNode.Address,
		Timestamp: latest.TimeStamp,
		Unit:      output.Unit,
		Value:     latest.Value,
	}
	publisher.publishMessage(addr, latestMessage)
}

// publish the $history output values
func (publisher *ThisPublisherState) publishHistoryCommand(aliasAddress string, output *nodes.InOutput) {
	aliasSegments := strings.Split(aliasAddress, "/")
	aliasSegments[3] = HistoryCommand
	addr := strings.Join(aliasSegments, "/")
	timeStampStr := time.Now().Format("2006-01-02T15:04:05.000-0700")

	historyMessage := &HistoryMessage{
		Address:   addr,
		Duration:  0, // tbd
		Sender:    publisher.publisherNode.Address,
		Timestamp: timeStampStr,
		Unit:      output.Unit,
		History:   nodes.GetHistory(output),
	}
	publisher.publishMessage(addr, historyMessage)
}

// publishMessage encapsulates the message object in a payload, signs, and sends it
func (publisher *ThisPublisherState) publishMessage(address string, message interface{}) {
	buffer, err := json.MarshalIndent(message, " ", " ")
	if err != nil {
		publisher.Logger.Errorf("Error marshalling message for address %s: %s", address, err)
		return
	}
	signature := publisher.ecdsaSign(buffer)

	publication := &messenger.Publication{
		Message:   string(buffer),
		Signature: signature,
	}
	publisher.messenger.Publish(address, publication)
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
	s := latest.Value
	if len(s) > 30 {
		s = s[:30]
	}
	publisher.Logger.Infof("publish output value '%s' on %s", s, aliasAddress)

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
			publisher.publishMessage(addr, node)
		}
		publisher.updatedNodes = nil
	}

	// publish updated inputs or outputs
	if publisher.updatedInOutputs != nil {
		for addr, inoutput := range publisher.updatedInOutputs {
			aliasAddress := publisher.getAliasAddress(addr)
			publisher.Logger.Infof("publish in/output discovery: %s", aliasAddress)
			publisher.publishMessage(aliasAddress, inoutput)
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
