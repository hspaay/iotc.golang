// Package publisher with handling of publisher discovery
package publisher

import (
	"encoding/json"

	"github.com/hspaay/iotconnect.golang/messaging"
	"github.com/hspaay/iotconnect.golang/nodes"
)

// handlePublisherDiscovery collects remote publishers of the publisher's zone for their public key
// Used to verify signatures of incoming configuration and input messages
// address contains the publisher's discovery address: zone/publisher/$publisher/$node
// publication contains a message with the publisher node info
func (publisher *Publisher) handlePublisherDiscovery(address string, publication *messaging.Publication) {
	var pubNode nodes.Node
	err := json.Unmarshal(publication.Message, &pubNode)
	if err != nil {
		publisher.Logger.Warningf("Unable to unmarshal Publisher Node in %s: %s", address, err)
		return
	}
	publisher.updateMutex.Lock()
	publisher.zonePublishers[address] = &pubNode
	publisher.updateMutex.Unlock()
	publisher.Logger.Infof("Discovered publisher %s", address)
}
