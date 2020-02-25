// Package publisher with handling commands for my node inputs
package publisher

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/asn1"
	"encoding/json"
	"iotzone/messenger"
	"iotzone/standard"
)

// VerifyECDSASignature Verify a ECDSA signature using the given public key
// See also https://leanpub.com/gocrypto/read#leanpub-auto-cryptographic-hashing-algorithms
func VerifyECDSASignature(message []byte, signature []byte, pub *ecdsa.PublicKey) bool {
	var rs ECDSASignature
	if _, err := asn1.Unmarshal(signature, &rs); err != nil {
		return false
	}

	hashed := sha256.Sum256(message)
	return ecdsa.Verify(pub, hashed[:], rs.R, rs.S)
}

// VerifyMessageSignature Verify the sender's signature
func (publisher *ThisPublisherState) VerifyMessageSignature(
	sender string, message json.RawMessage, signature string) bool {

	publisher.updateMutex.Lock()
	node := publisher.zonePublishers[sender]
	publisher.updateMutex.Unlock()

	if node == nil {
		return false
	}
	return false
}

// handle an incoming a set command for one of our nodes. This:
// - check if the signature is valid
// - check if the node is valid
// - pass the input value update to the adapter's callback method set in Start()
// TODO: track which nodes are allowed to control input
func (publisher *ThisPublisherState) handleNodeInput(address string, publication *messenger.Publication) {
	// TODO: authorization check
	input := publisher.GetInput(address)
	if input == nil || publication.Message == nil {
		publisher.Logger.Infof("handleNodeInput unknown input for address %s or missing message", address)
		return
	}

	var setMessage standard.SetMessage
	err := json.Unmarshal([]byte(publication.Message), &setMessage)
	if err != nil {
		publisher.Logger.Infof("Unable to unmarshal SetMessage in %s", address)
		return
	}

	isValid := publisher.VerifyMessageSignature(setMessage.Sender, publication.Message, publication.Signature)
	isValid = true
	if !isValid {
		publisher.Logger.Warningf("Incoming message verification failed for sender: %s", setMessage.Sender)
		return
	}
	if publisher.onSetMessage != nil {
		publisher.onSetMessage(input, &setMessage)
	}
}
