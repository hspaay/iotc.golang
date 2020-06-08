// Package subscriber with discovery and receiving of remote nodes
package subscriber

import (
	"encoding/json"
	"sync"

	"github.com/hspaay/iotc.golang/iotc"
	"github.com/hspaay/iotc.golang/messenger"
	"github.com/hspaay/iotc.golang/nodes"
	log "github.com/sirupsen/logrus"
)

// Subscriber carries the operating state of the consumer
// Start() will subscribe to discover all publishers.
// To discover nodes, subscribe to the publisher
type Subscriber struct {
	domain        string                   // The domain in which we live
	isRunning     bool                     // publisher was started and is running
	logger        *log.Logger              //
	inputList     *nodes.InputList         // inputs by discovery address
	messenger     messenger.IMessenger     // Message bus messenger to use
	nodes         *nodes.NodeList          // nodes by discovery address
	outputList    *nodes.OutputList        // outputs by discovery address
	publishers    *nodes.PublisherList     // publishers on the network
	signer        messenger.IMessageSigner // Signing and verification of messages
	subscriptions nodes.NodeList           // publishers to which we subscribe to receive their nodes
	updateMutex   *sync.Mutex              // mutex for async updating and publishing
}

// Start listen for publisher nodes
func (subscriber *Subscriber) Start() {
	if !subscriber.isRunning {
		subscriber.logger.Warningf("Starting subscriber")
		subscriber.updateMutex.Lock()
		subscriber.isRunning = true
		subscriber.updateMutex.Unlock()

		// TODO: support LWT
		subscriber.messenger.Connect("", "")

		// subscribe to receive all publisher identities
		pubAddr := nodes.MakePublisherIdentityAddress("+", "+")
		subscriber.messenger.Subscribe(pubAddr, subscriber.handlePublisherDiscovery)

		subscriber.logger.Warningf("Subscriber started")
	}
}

// Stop listen for messages
func (subscriber *Subscriber) Stop() {
	if subscriber.isRunning {
		subscriber.logger.Warningf("Stopping subscriber")
		subscriber.updateMutex.Lock()
		subscriber.isRunning = false
		subscriber.updateMutex.Unlock()
		subscriber.logger.Warningf("Subscriber stopped")
	}
}

// handlePublisherDiscovery stores discovered publishers for their public key
// Used to verify signatures of incoming configuration and input messages
// address contains the publisher's identity address: domain/publisherId/$identity
// message is the LWS signed message containing the publisher identity
func (subscriber *Subscriber) handlePublisherDiscovery(address string, message []byte) {
	var identity iotc.PublisherIdentityMessage

	// FIXME: verify with the signature with the domain security service (DSS) public key or certificate
	payload, err := subscriber.signer.Verify(message)
	if err != nil {
		subscriber.logger.Warningf("handlePublisherDiscovery Invalid message: %s", err)
		// return
	}

	// Decode the message into a NodeDiscoveryMessage type
	err = json.Unmarshal(payload, &identity)
	if err != nil {
		subscriber.logger.Warningf("Unable to unmarshal Publisher Identity in %s: %s", address, err)
		return
	}

	// TODO: if the publisher is in a secure zone its identity must have a valid signature from the ZCAS service
	// assume the publisher has a valid identity
	subscriber.updateMutex.Lock()
	defer subscriber.updateMutex.Unlock()

	// TODO: Verify that the publisher is valid...
	subscriber.publishers.UpdatePublisher(&identity)
	subscriber.logger.Infof("Discovered publisher %s", address)

}

// NewSubscriber creates a subscriber instance for discoverying publishers, nodes, inputs and
// outputs and receive output values
// domain for this subscriber lives in
// messenger for subscribing to the message bus
func NewSubscriber(domain string, messenger messenger.IMessenger) *Subscriber {

	var subscriber = &Subscriber{
		domain:      domain,
		inputList:   nodes.NewInputList(),
		logger:      log.New(),
		messenger:   messenger,
		nodes:       nodes.NewNodeList(),
		outputList:  nodes.NewOutputList(),
		publishers:  nodes.NewPublisherList(),
		updateMutex: &sync.Mutex{},
	}
	return subscriber
}
