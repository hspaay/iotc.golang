// Package publisher with handling of publisher discovery
package publisher

import (
	"crypto/ecdsa"
	"encoding/json"

	"github.com/hspaay/iotc.golang/iotc"
	"github.com/hspaay/iotc.golang/messenger"
)

// // handleNodeDiscovery collects and saves any discovered node
// func (publisher *Publisher) handleNodeDiscovery(address string, publication *iotc.Publication) {
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
// These publishers will be added to the zonePublishers for verification of signatures of
// input or configuration commands.
// address contains the publisher's discovery address: zone/publisher/$publisher/$node
// publication contains a message with the publisher node info.
func (publisher *Publisher) handlePublisherDiscovery(address string, publication *iotc.Publication) {
	var pubNode iotc.NodeDiscoveryMessage

	// Decode the message into a NodeDiscoveryMessage type
	err := json.Unmarshal([]byte(publication.Message), &pubNode)
	if err != nil {
		publisher.Logger.Warningf("Unable to unmarshal Publisher Node in %s: %s", address, err)
		return
	}
	// Verify that the message comes from the publisher using the publisher's own address and public key
	sender := pubNode.Identity.Address
	pubKeyStr := pubNode.Identity.PublicKeySigning

	var pubKey *ecdsa.PublicKey = messenger.DecodePublicKey(pubKeyStr)
	isValid := messenger.VerifyEcdsaSignature(publication.Message, publication.Signature, pubKey)
	if !isValid {
		publisher.Logger.Warningf("Incoming node publication verification failed for sender: %s", sender)
		return
	}

	// TODO: if the publisher is in a secure zone its identity must have a valid signature from the ZCAS service
	// assume the publisher has a valid identity
	publisher.updateMutex.Lock()
	defer publisher.updateMutex.Unlock()

	// TODO: Verify that the publisher is valid...
	publisher.zonePublishers[address] = &pubNode
	publisher.Logger.Infof("Discovered publisher %s", address)
}
