// Package nodes with publisher management
package nodes

import (
	"fmt"
	"sync"

	"github.com/hspaay/iotc.golang/iotc"
)

// PublisherList with discovered publishers
type PublisherList struct {
	// don't access directly. This is only accessible for serialization
	publisherMap map[string]*iotc.PublisherIdentityMessage
	updateMutex  *sync.Mutex // mutex for async updating of nodes
}

// GetAllPublishers returns a list of discovered publishers
func (pubList *PublisherList) GetAllPublishers() []*iotc.PublisherIdentityMessage {
	pubList.updateMutex.Lock()
	defer pubList.updateMutex.Unlock()

	var identList = make([]*iotc.PublisherIdentityMessage, 0)
	for _, identity := range pubList.publisherMap {
		identList = append(identList, identity)
	}
	return identList
}

// GetPublisherByAddress returns a publisher Identity by its identity discovery address
// Returns nil if address has no known node
func (pubList *PublisherList) GetPublisherByAddress(address string) *iotc.PublisherIdentityMessage {
	pubList.updateMutex.Lock()
	defer pubList.updateMutex.Unlock()

	var identity = pubList.publisherMap[address]
	return identity
}

// UpdatePublisher replaces a publisher identity
// Intended for use within a locked section
func (pubList *PublisherList) UpdatePublisher(pub *iotc.PublisherIdentityMessage) {
	pubList.updateMutex.Lock()
	defer pubList.updateMutex.Unlock()

	pubList.publisherMap[pub.Address] = pub
}

// MakePublisherIdentityAddress generates the address of a publisher:
//   domain/publisherID/$identity
// Intended for lookup of nodes in the node list.
// domain of the domain the node lives in.
// publisherID of the publisher for this node, unique for the domain
func MakePublisherIdentityAddress(domain string, publisherID string) string {
	address := fmt.Sprintf("%s/%s/%s", domain, publisherID, iotc.MessageTypeIdentity)
	return address
}

// NewPublisherList creates a new list of discovered publishers
func NewPublisherList() *PublisherList {
	pubList := &PublisherList{
		publisherMap: make(map[string]*iotc.PublisherIdentityMessage),
		updateMutex:  &sync.Mutex{},
	}
	return pubList
}
