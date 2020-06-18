// Package publisher with handling of publisher discovery
package publisher

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/hspaay/iotc.golang/iotc"
	"github.com/hspaay/iotc.golang/messenger"
	"gopkg.in/square/go-jose.v2"
)

// handleIdentityUpdate handles the set command for an update to this publisher identity.
// Intended to be set by the DSS to issue renewed keys and signature.
// This message must be encrypted and signed or it will be discarded.
func (publisher *Publisher) handleIdentityUpdate(address string, message string) {
	var fullIdentity iotc.PublisherFullIdentity

	// Expect the message to be encrypted
	isEncrypted, dmessage, err := messenger.DecryptMessage(message, publisher.identityPrivateKey)

	if !isEncrypted {
		publisher.logger.Warnf("handleIdentityUpdate: message to '%s' must be encrypted but isn't. Message discarded.", address)
		return
	} else if err != nil {
		publisher.logger.Warnf("handleIdentityUpdate: decryption failed of message to '%s'. Message discarded.", address)
		return
	}

	// Verify the message using the public key of the sender (should be the DSS)
	isSigned, err := messenger.VerifySender(dmessage, &fullIdentity, publisher.domainPublishers.GetPublisherKey)
	if !isSigned {
		// commands must use signed messages
		publisher.logger.Warnf("handleIdentityUpdate: Identity update '%s' is not signed. Message discarded.", address)
		return
	} else if err != nil {
		// signing failed, discard the message
		publisher.logger.Warnf("handleIdentityUpdate: Signature verification failed for  %s. Message discarded.", address)
		return
	}
	privKey := messenger.PrivateKeyFromPem(fullIdentity.PrivateKey)
	publisher.identityPrivateKey = privKey
	publisher.identity = &fullIdentity
}

// handleDSSDiscovery discoveres the identity of the domain security service
// The DSS publish signing key is used to verify the identity of all publishers
// Without a DSS, all publishers are unverified.
func (publisher *Publisher) handleDSSDiscovery(dssIdentityMsg *iotc.PublisherIdentityMessage) {
	var dssIdentity *iotc.PublisherIdentityMessage
	// Verify the identity of the DSS
	// TODO: CA support. For now assume address protection is used so this is trusted.

	// dssSigningPem := dssIdentity.Identity.PublicKeySigning
	// dssSigningKey := messenger.PublicKeyFromPem(dssSigningPem)
	// publisher.dssSigningKey = dssSigningKey
	publisher.domainPublishers.UpdatePublisher(dssIdentity)
}

// handlePublisherDiscovery collects and saves remote publishers
// Intended for discovery of available publishers and for verification of signatures of
// configuration and input messages received from these publishers.
// Handle the following trust scenarios:
//  A: Discovery of the DSS. Address protection or use a CA.
//  B: Trust address protection - always accept the publisher if its message is signed by itself
//  C: Trust DSS signing - verify identity is signed by DSS
//
// address contains the publisher's identity address: <domain>/<publisher>/$identity
// message contains the publisher identity message
func (publisher *Publisher) handlePublisherDiscovery(address string, message string) {
	var pubIdentityMsg *iotc.PublisherIdentityMessage
	var payload string

	// message can be signed or not signed so start with trying to parse
	jseSignature, err := jose.ParseSigned(string(message))
	if err != nil {
		// message isn't signed
		if publisher.signingMethod == SigningMethodJWS {
			// message must be signed though. Discard
			publisher.logger.Warnf("handlePublisherDiscovery: Publisher update isn't signed but only signed updates are accepted. Publisher: %s", address)
			return
		}
		// accept the unsigned message as signing isn't required
		payload = message
	} else {
		// message is signed. The signature must verify with the publisher signing key included in the message
		payload = string(jseSignature.UnsafePayloadWithoutVerification())
	}

	err = json.Unmarshal([]byte(payload), &pubIdentityMsg)
	if err != nil {
		publisher.logger.Warnf("handlePublisherDiscovery: Failed parsing json payload [unsigned]: %s", err)
		// abort
		return
	}

	// Handle the DSS publisher separately
	dssAddress := fmt.Sprintf("%s/%s/%s", publisher.Domain(), iotc.DSSPublisherID, iotc.MessageTypeIdentity)
	if address == dssAddress {
		publisher.handleDSSDiscovery(pubIdentityMsg)
		return
	}

	// So we have a publisher identity update. Determine if it is trusted.
	// 1: No DSS, assume address protection is in place
	// 2: Do we have a DSS? If so, require the identity is signed by the DSS
	dssSigningKey := publisher.domainPublishers.GetPublisherKey(dssAddress)
	if dssSigningKey == nil {
		// 1: No DSS, assume address protection is in place
		publisher.domainPublishers.UpdatePublisher(pubIdentityMsg)
		publisher.logger.Infof("handlePublisherDiscovery: Discovered publisher %s. [No DSS present]", address)

	} else {
		// 2: We have a DSS. Require the publisher identity is signed by the DSS
		// Create base64 encoded identity
		identityAsJSON, err := json.Marshal(pubIdentityMsg.Public)
		if err != nil {
			publisher.logger.Infof("handlePublisherDiscovery: Missing identity for %s", address)
			return
		}
		base64URLIdentity := base64.URLEncoding.EncodeToString(identityAsJSON)
		valid := messenger.VerifyEcdsaSignature(base64URLIdentity, pubIdentityMsg.IdentitySignature, dssSigningKey)
		if !valid {
			publisher.logger.Infof("handlePublisherDiscovery: Identity for %s doesn't have a valid DSS signature", address)
			return
		}
		// finally, The newly published identity is correctly signed by the DSS
		publisher.domainPublishers.UpdatePublisher(pubIdentityMsg)
		publisher.logger.Infof("Discovered publisher %s. [DSS verified]", address)
	}
}
