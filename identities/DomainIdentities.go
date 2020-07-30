// Package identities with management of discovered publisher identities
package identities

import (
	"crypto/ecdsa"
	"reflect"
	"strings"

	"github.com/iotdomain/iotdomain-go/lib"
	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/iotdomain/iotdomain-go/types"
)

// const DSSAddress = ""

// DomainIdentities with discovered and verified identities of publishers
type DomainIdentities struct {
	c              lib.DomainCollection //
	publicKeyCache map[string]*ecdsa.PublicKey
}

// // AddPublisher adds or replace a publisher identity
// func (domainIdentities *DomainIdentities) AddPublisher(pub *types.PublisherIdentityMessage) {
// 	domainIdentities.c.Add(pub.Address, pub)
// 	domainIdentities.c.UpdateMutex.Lock()
// 	defer domainIdentities.c.UpdateMutex.Unlock()
// 	// cache the publisher's identity public key
// 	pubKey := messaging.PublicKeyFromPem(pub.PublicKey)
// 	domainIdentities.publicKeyCache[pub.Address] = pubKey
// }

// GetAllPublishers returns a list of discovered publishers
func (domainIdentities *DomainIdentities) GetAllPublishers() []*types.PublisherIdentityMessage {
	var identList = make([]*types.PublisherIdentityMessage, 0)
	domainIdentities.c.GetAll(&identList)
	return identList
}

// GetDSSIdentity returns the Domain Security Service publisher identity
// Returns nil if no DSS was received
func (domainIdentities *DomainIdentities) GetDSSIdentity(domain string) *types.PublisherIdentityMessage {
	addr := MakePublisherIdentityAddress(domain, types.DSSPublisherID)
	dssMessage := domainIdentities.GetPublisherByAddress(addr)
	if dssMessage == nil {
		// DSS for the domain wasn't received
		return nil
	}
	return dssMessage
}

// GetPublisherByAddress returns a publisher Identity by its identity discovery address
// Returns nil if address has no known node
func (domainIdentities *DomainIdentities) GetPublisherByAddress(address string) *types.PublisherIdentityMessage {
	var domainIdentity = domainIdentities.c.GetByAddress(address)
	if domainIdentity == nil {
		return nil
	}
	return domainIdentity.(*types.PublisherIdentityMessage)
}

// GetPublisherKey returns the public key of a publisher for signature verification or encryption
// publisherAddress must start with domain/publisherId
// returns public key or nil if publisher public key is not found
func (domainIdentities *DomainIdentities) GetPublisherKey(publisherAddress string) *ecdsa.PublicKey {
	segments := strings.Split(publisherAddress, "/")
	if len(segments) < 2 {
		// missing publisherId
		return nil
	}
	identityAddress := MakePublisherIdentityAddress(segments[0], segments[1])

	pubKey := domainIdentities.publicKeyCache[identityAddress]
	if pubKey == nil {
		// if the public key isn't cached yet, then try getting it now
		identity := domainIdentities.c.GetByAddress(identityAddress).(*types.PublisherIdentityMessage)
		if identity != nil {
			pubKey = messaging.PublicKeyFromPem(identity.PublicKey)
			domainIdentities.publicKeyCache[identityAddress] = pubKey
		}
	}
	return pubKey
}

// Start subscribing to publisher identity discovery
func (domainIdentities *DomainIdentities) Start() {
	// subscription address for all inputs domain/publisher/node/type/instance/$input
	addr := MakePublisherIdentityAddress("+", "+")
	domainIdentities.c.MessageSigner.Subscribe(addr, domainIdentities.handleDiscoverIdentity)
}

// Stop polling for publishers
func (domainIdentities *DomainIdentities) Stop() {
	addr := MakePublisherIdentityAddress("+", "+")
	domainIdentities.c.MessageSigner.Unsubscribe(addr, domainIdentities.handleDiscoverIdentity)
}

// // SignIdentity returns a base64URL encoded signature of the given identity
// // used to sign the identity.
// func (pubList *PublisherList) SignIdentity(ident *types.PublisherIdentity, privKey *jose.SigningKey) string {
// 	signingKey := jose.SigningKey{Algorithm: jose.ES256, Key: privKey}
// 	signer, _ := jose.NewSigner(signingKey, nil)
// 	payload, _ := json.Marshal(ident)
// 	jws, _ := signer.Sign(payload)
// 	sig := jws.Signatures[0].Signature
// 	sigStr := base64.URLEncoding.EncodeToString(sig)
// 	return sigStr
// }

// handleDiscoverIdentity adds discovered publisher identities to the collection
func (domainIdentities *DomainIdentities) handleDiscoverIdentity(address string, message string) error {
	var discoMsg types.PublisherIdentityMessage

	err := domainIdentities.c.HandleDiscovery(address, message, &discoMsg)
	return err
}

// NewDomainIdentities creates a new list of discovered publishers
func NewDomainIdentities(messageSigner *messaging.MessageSigner) *DomainIdentities {
	domainIdentities := &DomainIdentities{
		c:              lib.NewDomainCollection(messageSigner, reflect.TypeOf(&types.InputDiscoveryMessage{})),
		publicKeyCache: make(map[string]*ecdsa.PublicKey),
	}
	return domainIdentities
}
