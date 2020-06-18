// Package messenger with methods for JWE encryption of messages
package messenger

import (
	"crypto/ecdsa"

	"gopkg.in/square/go-jose.v2"
)

// EncryptMessage encrypts and serializes the message using JWE
func EncryptMessage(message string, publicKey *ecdsa.PublicKey) (serialized string, err error) {

	recpnt := jose.Recipient{Algorithm: jose.ECDH_ES, Key: publicKey}

	encrypter, err := jose.NewEncrypter(jose.A128CBC_HS256, recpnt, nil)

	if err != nil {
		return message, err
	}

	jwe, err := encrypter.Encrypt([]byte(message))
	if err != nil {
		return message, err
	}
	serialized, _ = jwe.CompactSerialize()
	return serialized, err
}

// DecryptMessage deserializes and decrypts the message using JWE
// This returns the decrypted message, or the input message if the message was not encrypted
func DecryptMessage(serialized string, privateKey *ecdsa.PrivateKey) (isEncrypted bool, message string, err error) {
	message = serialized
	decrypter, err := jose.ParseEncrypted(serialized)
	if err == nil {
		dmessage, err := decrypter.Decrypt(privateKey)
		message = string(dmessage)
		return true, message, err
	}
	return false, message, err
}
