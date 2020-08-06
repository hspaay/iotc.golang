// Package identities with management of discovered publisher identities
package identities

import (
	"crypto/ecdsa"
	"encoding/json"
	"io/ioutil"
	"reflect"
	"strings"

	"github.com/iotdomain/iotdomain-go/lib"
	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/sirupsen/logrus"
)

// const DSSAddress = ""

// DomainPublisherIdentities with discovered and verified identities of publishers
type DomainPublisherIdentities struct {
	c              lib.DomainCollection //
	publicKeyCache map[string]*ecdsa.PublicKey
}

// AddIdentity adds a new public identity and generate its public key in the cache
// If the identity already exists, it will be replaced
func (pubIdentities *DomainPublisherIdentities) AddIdentity(identity *types.PublisherIdentityMessage) {
	pubIdentities.c.Add(identity.Address, identity)
	pubKey := messaging.PublicKeyFromPem(identity.PublicKey)
	pubIdentities.publicKeyCache[identity.Address] = pubKey
}

// GetAllPublishers returns a list of discovered publishers
func (pubIdentities *DomainPublisherIdentities) GetAllPublishers() []*types.PublisherIdentityMessage {
	var identList = make([]*types.PublisherIdentityMessage, 0)
	pubIdentities.c.GetAll(&identList)
	return identList
}

// GetDSSIdentity returns the Domain Security Service publisher identity
// Returns nil if no DSS was received
func (pubIdentities *DomainPublisherIdentities) GetDSSIdentity(domain string) *types.PublisherIdentityMessage {
	addr := MakePublisherIdentityAddress(domain, types.DSSPublisherID)
	dssMessage := pubIdentities.GetPublisherByAddress(addr)
	if dssMessage == nil {
		// DSS for the domain wasn't received
		return nil
	}
	return dssMessage
}

// GetPublisherByAddress returns a publisher Identity by its identity discovery address
// Returns nil if address has no known node
func (pubIdentities *DomainPublisherIdentities) GetPublisherByAddress(address string) *types.PublisherIdentityMessage {
	var domainIdentity = pubIdentities.c.GetByAddress(address)
	if domainIdentity == nil {
		return nil
	}
	return domainIdentity.(*types.PublisherIdentityMessage)
}

// GetPublisherKey returns the public key of a publisher for signature verification or encryption
// publisherAddress must start with domain/publisherId
// returns public key or nil if publisher public key is not found
func (pubIdentities *DomainPublisherIdentities) GetPublisherKey(publisherAddress string) *ecdsa.PublicKey {
	// cleanup the address
	segments := strings.Split(publisherAddress, "/")
	if len(segments) < 2 {
		// missing publisherId
		return nil
	}
	identityAddress := MakePublisherIdentityAddress(segments[0], segments[1])
	// first try using the public key cache
	pubKey := pubIdentities.publicKeyCache[identityAddress]
	// if pubKey == nil {
	// 	// if the public key isn't cached yet, try generating it from identity PEM record
	// 	idMsg := domainIdentities.c.GetByAddress(identityAddress)
	// 	if idMsg != nil {
	// 		domainIdentities.AddIdentity(idMsg.(*types.PublisherIdentityMessage))
	// 	}
	// }
	return pubKey
}

// LoadIdentities loads previously save identities from file
// Existing identities are retained but replaced if contained in the file
func (pubIdentities *DomainPublisherIdentities) LoadIdentities(filename string) error {
	identList := make([]*types.PublisherIdentityMessage, 0)

	jsonNodes, err := ioutil.ReadFile(filename)
	if err != nil {
		return lib.MakeErrorf("LoadIdentities: Unable to open file %s: %s", filename, err)
	}
	err = json.Unmarshal(jsonNodes, &identList)
	if err != nil {
		return lib.MakeErrorf("LoadIdentities: Error parsing JSON node file %s: %v", filename, err)
	}
	logrus.Infof("LoadIdentities: Identities loaded successfully from %s", filename)
	for _, ident := range identList {
		pubIdentities.AddIdentity(ident)
	}
	return nil

}

// SaveIdentities saves previously added identities to file
func (pubIdentities *DomainPublisherIdentities) SaveIdentities(filename string) error {
	collection := pubIdentities.GetAllPublishers()
	jsonText, err := json.MarshalIndent(collection, "", "  ")
	if err != nil {
		return lib.MakeErrorf("SaveIdentities: Error Marshalling JSON collection '%s': %v", filename, err)
	}
	err = ioutil.WriteFile(filename, jsonText, 0664)
	if err != nil {
		return lib.MakeErrorf("SaveIdentities: Error saving collection to JSON file %s: %v", filename, err)
	}
	logrus.Infof("SaveIdentities: Collection saved successfully to JSON file %s", filename)
	return nil
}

// VerifyPublisherIdentity verifies the integrity of the given identity record
// Verification fails when:
//  - domain/publisherID doesn't match the received address
//  - the issuer ID or its public key is missing
//  - the issuer is either the DSS or the publisher itself (self-signed)
//  - identity is expired
//  - a newer identity is already received
//  - the identity signature doesn't verify against the signing key (if provided)
//
// The signing key is only available if the identity of the issuer is known;
//  When a secured domain is joined, the issuer is the DSS whose identity must be received first.
//  When no secured domain is joined, the identity is self signed. Protection is
//   based on message bus ACLs. Only publishers can self sign their own identity.
//  When the issuer is a CA, the CA public key must be known
func VerifyPublisherIdentity(rxAddress string, ident *types.PublisherIdentityMessage,
	dssSigningKey *ecdsa.PublicKey) error {

	var signingKey *ecdsa.PublicKey

	// identity must contain public key, issuer and signature
	if ident.PublicKey == "" ||
		ident.Domain == "" ||
		ident.IssuerID == "" ||
		ident.IdentitySignature == "" {
		err := lib.MakeErrorf("VerifyIdentity: identity is incomplete for '%s'", rxAddress)
		return err
	}
	// domain/publisherID must match the address
	segments := strings.Split(rxAddress, "/")
	if len(segments) < 2 ||
		ident.Address != rxAddress ||
		ident.Domain != segments[0] ||
		ident.PublisherID != segments[1] {
		err := lib.MakeErrorf("VerifyPublisherIdentity: invalid domain/publisher '%s/%s', or address '%s'",
			ident.Domain, ident.PublisherID, ident.Address)
		return err
	}
	// only DSS or publisher itself are allowed to issue identity
	// TODO: support CA
	if ident.IssuerID != ident.PublisherID &&
		ident.IssuerID != types.DSSPublisherID {

		err := lib.MakeErrorf("VerifyPublisherIdentity: identity issuer %s of domain/publisher %s/%s must "+
			"be the DSS or self-signed", ident.IssuerID, ident.Domain, ident.PublisherID)
		return err
	}

	// identity must not be expired
	expired := IsIdentityExpired(ident)
	if expired {
		err := lib.MakeErrorf("VerifyIdentity: Identity '%s' is expired", rxAddress)
		return err
	}
	if ident.IssuerID == types.DSSPublisherID {
		signingKey = dssSigningKey
	} else {
		signingKey = messaging.PublicKeyFromPem(ident.PublicKey)
	}

	// Self signed or DSS signed identity
	err := messaging.VerifyIdentitySignature(ident, signingKey)
	if err != nil {
		return lib.MakeErrorf("VerifyIdentity: Verification of %s message failed. "+
			"The identity signature doesn't match", rxAddress)
	}

	return nil
}

// NewDomainPublisherIdentities creates a new list of discovered publishers
func NewDomainPublisherIdentities() *DomainPublisherIdentities {
	domainIdentities := &DomainPublisherIdentities{
		c:              lib.NewDomainCollection(reflect.TypeOf(&types.InputDiscoveryMessage{}), nil),
		publicKeyCache: make(map[string]*ecdsa.PublicKey),
	}
	domainIdentities.c.GetPublicKey = domainIdentities.GetPublisherKey
	return domainIdentities
}
