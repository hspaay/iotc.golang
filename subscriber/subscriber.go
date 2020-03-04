// Package subscriber with discovery and receiving of remote nodes
package subscriber

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/hspaay/iotconnect.golang/messenger"
	"github.com/hspaay/iotconnect.golang/standard"
	log "github.com/sirupsen/logrus"
)

// ThisSubscriberState carries the operating state of the subscriber
// Start() will subscribe to discover all publishers.
// To discover nodes, subscribe to the publisher
type ThisSubscriberState struct {
	Logger        *log.Logger               //
	messenger     messenger.IMessenger      // Message bus messenger to use
	zoneID        string                    // The zone in which we live
	isRunning     bool                      // publisher was started and is running
	subscriptions map[string]*standard.Node // publishers to which we subscribe to receive their nodes

	// handle updates in the background
	updateMutex *sync.Mutex                   // mutex for async updating and publishing
	publishers  map[string]*standard.Node     // publishers on the network
	nodes       map[string]*standard.Node     // nodes by discovery address
	inputs      map[string]*standard.InOutput // inputs by discovery address
	outputs     map[string]*standard.InOutput // outputs by discovery address
}

// Start listen for publisher nodes
func (subscriber *ThisSubscriberState) Start() {
	if !subscriber.isRunning {
		subscriber.Logger.Warningf("Starting subscriber")
		subscriber.updateMutex.Lock()
		subscriber.isRunning = true
		subscriber.updateMutex.Unlock()

		// TODO: support LWT
		subscriber.messenger.Connect("", "")

		// subscribe to receive any publisher node
		pubAddr := fmt.Sprintf("+/+/%s/%s", standard.PublisherNodeID, standard.CommandNodeDiscovery)
		subscriber.messenger.Subscribe(pubAddr, subscriber.handlePublisherDiscovery)

		subscriber.Logger.Warningf("Subscriber started")
	}
}

// Stop listen for messages
func (subscriber *ThisSubscriberState) Stop() {
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
func (subscriber *ThisSubscriberState) handlePublisherDiscovery(address string, publication *messenger.Publication) {
	var pubNode standard.Node
	err := json.Unmarshal([]byte(publication.Message), &pubNode)
	if err != nil {
		subscriber.Logger.Infof("Unable to unmarshal Publisher Node in %s", address)
		return
	}
	subscriber.publishers[address] = &pubNode
	subscriber.Logger.Infof("Discovered publisher %s", address)
}

// NewSubscriber creates a subscriber instance for discoverying nodes, inputs and
// outputs and receive output values
// zoneID for the zone this subscriber lives in
// messenger for subscribing to the message bus
func NewSubscriber(zoneID string, messenger messenger.IMessenger) *ThisSubscriberState {

	var subscriber = &ThisSubscriberState{
		inputs:      make(map[string]*standard.InOutput, 0),
		Logger:      log.New(),
		messenger:   messenger,
		nodes:       make(map[string]*standard.Node),
		outputs:     make(map[string]*standard.InOutput),
		publishers:  make(map[string]*standard.Node),
		updateMutex: &sync.Mutex{},
		zoneID:      zoneID,
	}
	return subscriber
}
