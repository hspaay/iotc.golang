// Package nodes with management of discovered publishers
package nodes

import (
	"crypto/ecdsa"
	"fmt"
	"strings"
	"sync"

	"github.com/hspaay/iotc.golang/iotc"
	"github.com/hspaay/iotc.golang/messenger"
)

// const DSSAddress = ""

// PublisherList with discovered and verified publishers
type PublisherList struct {
	// don't access directly. This is only accessible for serialization
	publisherMap  map[string]*iotc.PublisherIdentityMessage
	publisherKeys map[string]*ecdsa.PublicKey
	updateMutex   *sync.Mutex // mutex for async updating of nodes
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

// GetDSSIdentity returns the Domain Security Service publisher identity
// Returns nil if no DSS was received
func (pubList *PublisherList) GetDSSIdentity(domain string) *iotc.PublisherPublicIdentity {
	addr := MakePublisherIdentityAddress(domain, iotc.DSSPublisherID)
	dssMessage := pubList.GetPublisherByAddress(addr)
	if dssMessage == nil {
		// DSS for the domain wasn't received
		return nil
	}
	return &dssMessage.Public
}

// GetPublisherByAddress returns a publisher Identity by its identity discovery address
// Returns nil if address has no known node
func (pubList *PublisherList) GetPublisherByAddress(address string) *iotc.PublisherIdentityMessage {
	pubList.updateMutex.Lock()
	defer pubList.updateMutex.Unlock()

	var identity = pubList.publisherMap[address]
	return identity
}

// GetPublisherKey returns the public key of a publisher for signature verification or encryption
// publisherAddress starts with domain/publisherId
// returns public key or nil if publisher public key is not found
func (pubList *PublisherList) GetPublisherKey(publisherAddress string) *ecdsa.PublicKey {
	segments := strings.Split(publisherAddress, "/")
	if len(segments) < 2 {
		// missing publisherId
		return nil
	}
	identityAddress := fmt.Sprintf("%s/%s/%s", segments[0], segments[1], iotc.MessageTypeIdentity)

	// Use cached key instead of regenerating them each time
	pubKey := pubList.publisherKeys[identityAddress]
	if pubKey == nil {
		pub := pubList.GetPublisherByAddress(identityAddress)
		if pub == nil || pub.Public.PublicKey == "" {
			// unknown publisher
			return nil
		}
		pubKey = messenger.PublicKeyFromPem(pub.Public.PublicKey)
		pubList.publisherKeys[identityAddress] = pubKey
	}
	return pubKey
}

// // SignIdentity returns a base64URL encoded signature of the given identity
// // used to sign the identity.
// func (pubList *PublisherList) SignIdentity(ident *iotc.PublisherIdentity, privKey *jose.SigningKey) string {
// 	signingKey := jose.SigningKey{Algorithm: jose.ES256, Key: privKey}
// 	signer, _ := jose.NewSigner(signingKey, nil)
// 	payload, _ := json.Marshal(ident)
// 	jws, _ := signer.Sign(payload)
// 	sig := jws.Signatures[0].Signature
// 	sigStr := base64.URLEncoding.EncodeToString(sig)
// 	return sigStr
// }

// UpdatePublisher replaces a publisher identity
// Intended for use within a locked section
func (pubList *PublisherList) UpdatePublisher(pub *iotc.PublisherIdentityMessage) {
	pubList.updateMutex.Lock()
	defer pubList.updateMutex.Unlock()

	pubList.publisherMap[pub.Address] = pub
	pubList.publisherKeys[pub.Address] = nil // public key will be generated on next use
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
		publisherMap:  make(map[string]*iotc.PublisherIdentityMessage),
		publisherKeys: make(map[string]*ecdsa.PublicKey),
		updateMutex:   &sync.Mutex{},
	}
	return pubList
}
