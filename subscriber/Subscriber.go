// Package subscriber with discovery and receiving of remote nodes
package subscriber

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/hspaay/iotconnect.golang/messaging"
	"github.com/hspaay/iotconnect.golang/messenger"
	"github.com/hspaay/iotconnect.golang/nodes"
	log "github.com/sirupsen/logrus"
)

// Subscriber carries the operating state of the subscriber
// Start() will subscribe to discover all publishers.
// To discover nodes, subscribe to the publisher
type Subscriber struct {
	Logger        *log.Logger          //
	messenger     messenger.IMessenger // Message bus messenger to use
	zoneID        string               // The zone in which we live
	isRunning     bool                 // publisher was started and is running
	subscriptions nodes.NodeList       // publishers to which we subscribe to receive their nodes

	// handle updates in the background
	updateMutex *sync.Mutex       // mutex for async updating and publishing
	publishers  *nodes.NodeList   // publishers on the network
	nodes       *nodes.NodeList   // nodes by discovery address
	inputList   *nodes.InputList  // inputs by discovery address
	outputList  *nodes.OutputList // outputs by discovery address
}

// Start listen for publisher nodes
func (subscriber *Subscriber) Start() {
	if !subscriber.isRunning {
		subscriber.Logger.Warningf("Starting subscriber")
		subscriber.updateMutex.Lock()
		subscriber.isRunning = true
		subscriber.updateMutex.Unlock()

		// TODO: support LWT
		subscriber.messenger.Connect("", "")

		// subscribe to receive any publisher node
		pubAddr := fmt.Sprintf("+/+/%s/%s", messaging.PublisherNodeID, messaging.MessageTypeNodeDiscovery)
		subscriber.messenger.Subscribe(pubAddr, subscriber.handlePublisherDiscovery)

		subscriber.Logger.Warningf("Subscriber started")
	}
}

// Stop listen for messages
func (subscriber *Subscriber) Stop() {
	if subscriber.isRunning {
		subscriber.Logger.Warningf("Stopping subscriber")
		subscriber.updateMutex.Lock()
		subscriber.isRunning = false
		subscriber.updateMutex.Unlock()
		subscriber.Logger.Warningf("Subscriber stopped")
	}
}

// handlePublisherDiscovery stores discovered publishers in the zone for their public key
// Used to verify signatures of incoming configuration and input messages
// address contains the publisher's discovery address: zone/publisher/$publisher/$node
// publication contains a message with the publisher node info
func (subscriber *Subscriber) handlePublisherDiscovery(address string, publication *messaging.Publication) {
	var pubNode nodes.Node
	err := json.Unmarshal([]byte(publication.Message), &pubNode)
	if err != nil {
		subscriber.Logger.Warningf("Unable to unmarshal Publisher Node in %s", address)
		return
	} else if pubNode.Address != address {
		subscriber.Logger.Warningf("Received publisher Node with address %s on a different address %s", pubNode.Address)
		return
	}
	subscriber.publishers.UpdateNode(&pubNode)
	subscriber.Logger.Infof("Discovered publisher %s", address)
}

// NewSubscriber creates a subscriber instance for discoverying publishers, nodes, inputs and
// outputs and receive output values
// zoneID for the zone this subscriber lives in
// messenger for subscribing to the message bus
func NewSubscriber(zoneID string, messenger messenger.IMessenger) *Subscriber {

	var subscriber = &Subscriber{
		inputList:   nodes.NewInputList(),
		Logger:      log.New(),
		messenger:   messenger,
		nodes:       nodes.NewNodeList(),
		outputList:  nodes.NewOutputList(),
		publishers:  nodes.NewNodeList(),
		updateMutex: &sync.Mutex{},
		zoneID:      zoneID,
	}
	return subscriber
}
