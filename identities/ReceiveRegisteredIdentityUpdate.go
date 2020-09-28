// Package identities with handling of registered identity update command
package identities

import (
	"fmt"

	"github.com/iotdomain/iotdomain-go/lib"
	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/iotdomain/iotdomain-go/types"
)

// ReceiveRegisteredIdentityUpdate listens for the identity update command from the DSS
// This decrypts and verifies the signature of the command using the DSS public key when available
type ReceiveRegisteredIdentityUpdate struct {
	domain             string                   // the domain of this publisher
	publisherID        string                   // the registered publisher for the inputs
	messageSigner      *messaging.MessageSigner // subscription to command
	registeredIdentity *RegisteredIdentity      // the identity to update
}

// Start listening for updates to the registered identity
// Intended to receive new keys from the DSS
func (rxIdentity *ReceiveRegisteredIdentityUpdate) Start() {
	addr := fmt.Sprintf("%s/%s/%s", rxIdentity.domain, "+", types.MessageTypeSetInput)
	rxIdentity.messageSigner.Subscribe(addr, rxIdentity.ReceiveIdentityUpdate)
}

// Stop listening
func (rxIdentity *ReceiveRegisteredIdentityUpdate) Stop() {
	addr := fmt.Sprintf("%s/%s/%s", rxIdentity.domain, rxIdentity.publisherID, types.MessageTypeSetInput)
	rxIdentity.messageSigner.Unsubscribe(addr, rxIdentity.ReceiveIdentityUpdate)
}

// ReceiveIdentityUpdate handles an incoming a identity update command. This:
// - checks if the rawMessage is encrypted
// - checks the sender is the DSS
// - verifies if the sender (dss) signature is valid
// - passes the update to the adapter's callback set in Start()
func (rxIdentity *ReceiveRegisteredIdentityUpdate) ReceiveIdentityUpdate(address string, rawMessage string) error {
	var newIdentity types.PublisherFullIdentity

	isEncrypted, isSigned, err := rxIdentity.messageSigner.DecodeMessage(
		rawMessage, &newIdentity)

	if err != nil {
		return lib.MakeErrorf("HandleIdentityUpdate: Message to %s. Error %s'. Message discarded.", address, err)
	} else if !isEncrypted {
		return lib.MakeErrorf("HandleIdentityUpdate: Identity update '%s' is not encrypted. Message discarded.", address)
	} else if !isSigned {
		return lib.MakeErrorf("HandleIdentityUpdate: Identity update '%s' is not signed. Message discarded.", address)
	}
	// the sender must be the DSS (domain security service) of this domain
	dssAddress := MakePublisherIdentityAddress(rxIdentity.domain, types.DSSPublisherID)
	if newIdentity.Sender != dssAddress {
		return lib.MakeErrorf("HandleIdentityUpdate: Sender is %s instead of the DSS %s. Identity update discarded.",
			newIdentity.Sender, dssAddress)
	}
	if rxIdentity.registeredIdentity != nil {
		rxIdentity.registeredIdentity.UpdateIdentity(&newIdentity)
		rxIdentity.registeredIdentity.SaveIdentity()
	}
	return err
}

// NewReceiveRegisteredIdentityUpdate listens for updates to the identity as provided by
// the domain security service (DSS). Run Start() to start listening.
// RegisteredIdentity or DomainIdentites can be nil for testing
func NewReceiveRegisteredIdentityUpdate(
	regIdent *RegisteredIdentity,
	messageSigner *messaging.MessageSigner) *ReceiveRegisteredIdentityUpdate {

	rxIdent := &ReceiveRegisteredIdentityUpdate{
		domain:             regIdent.domain,
		messageSigner:      messageSigner,
		publisherID:        regIdent.publisherID,
		registeredIdentity: regIdent,
	}
	return rxIdent
}
