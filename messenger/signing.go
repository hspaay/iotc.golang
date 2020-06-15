// Package messenger with methods for signing and verifying a message
package messenger

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/asn1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"reflect"

	"github.com/hspaay/iotc.golang/iotc"
	"gopkg.in/square/go-jose.v2"
)

// CreateEcdsaSignature creates a ECDSA256 signature from the payload using the provided private key
// This returns a base64url encoded signature
func CreateEcdsaSignature(payload string, privateKey *ecdsa.PrivateKey) string {
	if privateKey == nil {
		return ""
	}
	hashed := sha256.Sum256([]byte(payload))
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

// SignEncodeIdentity returns a base64URL encoded ECDSA256 signature of the given object
// Intended to sign the publisher identity
func SignEncodeIdentity(ident *iotc.PublisherIdentity, privKey *jose.SigningKey) string {
	signingKey := jose.SigningKey{Algorithm: jose.ES256, Key: privKey}
	signer, _ := jose.NewSigner(signingKey, nil)
	payload, _ := json.Marshal(ident)
	jws, _ := signer.Sign(payload)
	sig := jws.Signatures[0].Signature
	sigStr := base64.URLEncoding.EncodeToString(sig)
	return sigStr
}

// VerifyEcdsaSignature the payload using the base64url encoded signature and public key
// payload is a text or base64 encoded raw data
// signatureB64urlEncoded is the ecdsa 256 URL encoded signature
func VerifyEcdsaSignature(payload string, signatureB64urlEncoded string, publicKey *ecdsa.PublicKey) bool {
	var rs ECDSASignature
	signature, err := base64.URLEncoding.DecodeString(signatureB64urlEncoded)
	if err != nil {
		return false
	}
	if _, err = asn1.Unmarshal(signature, &rs); err != nil {
		return false
	}

	hashed := sha256.Sum256([]byte(payload))
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

// VerifySender verifies if a message is JWS signed and if so, verifies using the 'Sender' Attribute to
// get the public key to verify with.
// param message must contain a json attribute named 'Sender'.
// param getPublicKey lookup function that provides the public key from the given sender address.
// param object is the address of the expected object in the message where it will be unmarshalled.
// This returns a flag if the message was signed and if so, an error if the verification failed
func VerifySender(message string, object interface{}, getPublicKey func(address string) *ecdsa.PublicKey) (isSigned bool, err error) {
	jwsSignature, err := jose.ParseSigned(message)
	if err != nil {
		// message is not signed, try to unmarshal it directly
		err = json.Unmarshal([]byte(message), object)
		return false, err
	}
	payload := jwsSignature.UnsafePayloadWithoutVerification()
	err = json.Unmarshal([]byte(payload), object)
	if err != nil {
		// message doesn't have a json payload
		return true, err
	}
	// determine who the sender is and get its public key
	reflObject := reflect.ValueOf(object).Elem()
	reflSender := reflObject.FieldByName("Sender")
	if !reflSender.IsValid() {
		err := errors.New("VerifySender: object doesn't have a Sender field")
		return false, err
	}
	sender := reflSender.String()
	if sender == "" {
		err := errors.New("VerifySender: Missing 'Sender' attribute in message")
		return true, err
	}
	publicKey := getPublicKey(sender)
	if publicKey == nil {
		err := errors.New("VerifySender: No public key available for sender " + sender)
		return true, err
	}

	_, err = jwsSignature.Verify(publicKey)
	return true, err
}
