// Package publisher with handling of publisher discovery
package publisher

import (
	"encoding/json"

	"github.com/hspaay/iotc.golang/messaging"
	"github.com/hspaay/iotc.golang/nodes"
)

// // handleNodeDiscovery collects and saves any discovered node
// func (publisher *Publisher) handleNodeDiscovery(address string, publication *messaging.Publication) {
// 	var pubNode nodes.Node
// 	err := json.Unmarshal(publication.Message, &pubNode)
// 	if err != nil {
// 		publisher.Logger.Warningf("Unable to unmarshal Node in %s: %s", address, err)
// 		return
// 	}
// 	// TODO. Do we need to verify the node identity?
// 	publisher.Nodes.UpdateNode(&pubNode)

// 	// save the new node
// 	if publisher.persistFolder != "" {
// 		persist.SaveNodes(publisher.persistFolder, publisher.publisherID, publisher.Nodes)
// 	}

// 	publisher.Logger.Infof("Discovered node %s", address)
// }

// handlePublisherDiscovery collects and saves remote publishers of the publisher's zone for their public key
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
	// TODO. Do we need to verify the identity?
	publisher.Nodes.UpdateNode(&pubNode)

	publisher.updateMutex.Lock()
	defer publisher.updateMutex.Unlock()

	publisher.zonePublishers[address] = &pubNode
	// save the new node

	// if publisher.persistFolder != "" {
	// 	persist.SaveNodes(publisher.persistFolder, publisher.publisherID, publisher.Nodes)
	// }

	publisher.Logger.Infof("Discovered publisher %s", address)
}
