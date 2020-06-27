// Package publisher with signing, encrypting and publishing of objects
package publisher

import (
	"crypto/ecdsa"
	"encoding/json"

	"github.com/iotdomain/iotdomain-go/messenger"
)

// publishObject encapsulates the message object in a payload, signs the message, and sends it.
// If an encryption key is provided then the signed message will be encrypted.
// address of the publication
// object to publish. This will be marshalled to JSON and signed by this publisher
func (publisher *Publisher) publishObject(address string, retained bool, object interface{}, encryptionKey *ecdsa.PublicKey) error {
	// payload, err := json.Marshal(object)
	payload, err := json.MarshalIndent(object, " ", " ")
	if err != nil {
		publisher.logger.Errorf("Publisher.publishMessage: Error marshalling message for address %s: %s", address, err)
		return err
	}
	if encryptionKey != nil {
		err = publisher.publishEncrypted(address, retained, string(payload), encryptionKey)
	} else {
		err = publisher.publishSigned(address, retained, string(payload))
	}
	return err
}

// publishEncrypted sign and encrypts the payload and publish the resulting message on the given address
// Signing only happens if the publisher's signingMethod is set to SigningMethodJWS
func (publisher *Publisher) publishEncrypted(address string, retained bool, payload string, publicKey *ecdsa.PublicKey) error {
	var err error
	message := payload
	// first sign, then encrypt as per RFC
	if publisher.signingMethod == SigningMethodJWS {
		message, err = messenger.CreateJWSSignature(string(payload), publisher.identityPrivateKey)
	}
	emessage, err := messenger.EncryptMessage(message, publicKey)
	err = publisher.messenger.Publish(address, retained, emessage)
	return err
}

// publishSigned sign the payload and publish the resulting message on the given address
// Signing only happens if the publisher's signingMethod is set to SigningMethodJWS
func (publisher *Publisher) publishSigned(address string, retained bool, payload string) error {
	var err error

	// default is unsigned
	message := payload

	if publisher.signingMethod == SigningMethodJWS {
		message, err = messenger.CreateJWSSignature(string(payload), publisher.identityPrivateKey)
		if err != nil {
			publisher.logger.Errorf("Publisher.publishMessage: Error signing message for address %s: %s", address, err)
		}
	}
	err = publisher.messenger.Publish(address, retained, message)
	return err
}
