// Package publisher with handling of publisher discovery
package publisher

import (
	"encoding/json"

	"github.com/hspaay/iotc.golang/iotc"
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

// handlePublisherDiscovery collects and saves remote publishers
// Intended for discovery of available publishers and for verification of signatures of
// configuration and input messages received from these publishers.
// address contains the publisher's identity address: zone/publisher/$identity
// message contains the publisher identity
func (publisher *Publisher) handlePublisherDiscovery(address string, message []byte) {
	var identity iotc.PublisherIdentityMessage

	// FIXME: verify with the signature with the domain security's public key
	payload, err := publisher.signer.Verify(message)
	if err != nil {
		publisher.logger.Warningf("handlePublisherDiscovery Invalid message: %s", err)
		// return
	}

	// Decode the message into a NodeDiscoveryMessage type
	err = json.Unmarshal(payload, &identity)
	if err != nil {
		publisher.logger.Warningf("Unable to unmarshal Publisher Identity in %s: %s", address, err)
		return
	}
	// Verify that the message comes from the publisher using the publisher's own address and public key
	// sender := publisher.identity.Address
	// pubKeyStr := publisher.identity.PublicKeySigning

	// var pubKey *ecdsa.PublicKey = messenger.DecodePublicKey(pubKeyStr)
	// isValid := messenger.VerifyEcdsaSignature(message, publication.Signature, pubKey)
	// if !isValid {
	// 	publisher.Logger.Warningf("Incoming node publication verification failed for sender: %s", sender)
	// 	return
	// }

	// TODO: if the publisher is in a secure zone its identity must have a valid signature from the ZCAS service
	// assume the publisher has a valid identity
	publisher.updateMutex.Lock()
	defer publisher.updateMutex.Unlock()

	// TODO: Verify that the publisher is valid...
	publisher.domainPublishers.UpdatePublisher(&identity)
	publisher.logger.Infof("Discovered publisher %s", address)
}
