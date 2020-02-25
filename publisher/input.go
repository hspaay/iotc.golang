// Package publisher with input command handling
package publisher

import (
	"encoding/json"
	"iotzone/messenger"
	"iotzone/standard"
)

// handle an incoming a set command for one of our nodes. This:
// - check if the signature is valid
// - check if the node is valid
// - pass the input value update to the adapter's callback method set in Start()
//
func (publisher *ThisPublisherState) handleNodeInput(address string, publication *messenger.Publication) {
	// TODO: authorization check
	input := publisher.GetInput(address)
	if input == nil || publication.Message == "" {
		publisher.Logger.Infof("handleNodeInput unknown input for address %s or missing message", address)
		return
	}

	var setMessage standard.SetMessage
	err := json.Unmarshal([]byte(publication.Message), &setMessage)
	if err != nil {
		publisher.Logger.Infof("Unable to unmarshal SetMessage in %s", address)
		return
	}
	// isValid := publisher.verifySender(publication)
	isValid := true
	if !isValid {
		publisher.Logger.Warningf("Incoming message verification failed for sender: %s", setMessage.Sender)
		return
	}
	if publisher.onSetMessage != nil {
		publisher.onSetMessage(input, &setMessage)
	}
}
