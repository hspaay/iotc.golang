// Package messenger with methods for signing and verifying a message
package messenger

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/asn1"
	"encoding/base64"

	"gopkg.in/square/go-jose.v2"
)

// CreateEcdsaSignature creates a  ECDSA256 signature from the message using the provided private key
// This returns a base64url encoded signature
func CreateEcdsaSignature(message []byte, privateKey *ecdsa.PrivateKey) string {
	if privateKey == nil {
		return ""
	}
	hashed := sha256.Sum256(message)
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hashed[:])
	if err != nil {
		return ""
	}
	sig, err := asn1.Marshal(ECDSASignature{r, s})
	return base64.URLEncoding.EncodeToString(sig)
}

// CreateJWSSignature signs the payload using JSE ES256 and return the JSE full serialized message
func CreateJWSSignature(payload string, privateKey *ecdsa.PrivateKey) (string, error) {
	joseSigner, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.ES256, Key: privateKey}, nil)
	signedObject, err := joseSigner.Sign([]byte(payload))
	if err != nil {
		return "", err
	}
	serialized := signedObject.FullSerialize()
	return serialized, err
}

// VerifyEcdsaSignature the message using the base64url encoded signature and public key
// message is a text message or base64 encoded raw data
// signatureB64urlEncoded is the ecdsa 256 URL encoded signature
func VerifyEcdsaSignature(message string, signatureB64urlEncoded string, publicKey *ecdsa.PublicKey) bool {
	var rs ECDSASignature
	signature, err := base64.URLEncoding.DecodeString(signatureB64urlEncoded)
	if err != nil {
		return false
	}
	if _, err = asn1.Unmarshal(signature, &rs); err != nil {
		return false
	}

	hashed := sha256.Sum256([]byte(message))
	return ecdsa.Verify(publicKey, hashed[:], rs.R, rs.S)
}

// VerifyJWSMessage verifies a signed message and returns its payload
// message is the message to verify
// publicKey from the signer. This must be known to verify the message.
func VerifyJWSMessage(message string, publicKey *ecdsa.PublicKey) (payload string, err error) {
	jwsSignature, err := jose.ParseSigned(message)
	if err != nil {
		return "", err
	}
	payloadB, err := jwsSignature.Verify(publicKey)
	return string(payloadB), err
}
