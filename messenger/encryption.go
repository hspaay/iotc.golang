// Package messenger with ECDSA signing and encryption functions
package messenger

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"encoding/pem"
	"math/big"
)

// ECDSASignature ...
type ECDSASignature struct {
	R, S *big.Int
}

// DecodeKeys converts ascii encoded public and private keys into a ECDSA object for use in the application
func DecodeKeys(pemEncoded string, pemEncodedPub string) (*ecdsa.PrivateKey, *ecdsa.PublicKey) {
	block, _ := pem.Decode([]byte(pemEncoded))
	x509Encoded := block.Bytes
	privateKey, _ := x509.ParseECPrivateKey(x509Encoded)

	publicKey := DecodePublicKey(pemEncodedPub)
	return privateKey, publicKey
}

// DecodePublicKey converts a ascii encoded public key into a ECDSA public key
func DecodePublicKey(pemEncodedPub string) *ecdsa.PublicKey {
	blockPub, _ := pem.Decode([]byte(pemEncodedPub))
	x509EncodedPub := blockPub.Bytes
	genericPublicKey, _ := x509.ParsePKIXPublicKey(x509EncodedPub)
	publicKey := genericPublicKey.(*ecdsa.PublicKey)

	return publicKey
}

// EncodeKeys converts a public/private key pair into their PEM encoded ascii format
// see also https://stackoverflow.com/questions/21322182/how-to-store-ecdsa-private-key-in-go
func EncodeKeys(privateKey *ecdsa.PrivateKey, publicKey *ecdsa.PublicKey) (string, string) {
	x509Encoded, _ := x509.MarshalECPrivateKey(privateKey)
	pemEncoded := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: x509Encoded})

	x509EncodedPub, _ := x509.MarshalPKIXPublicKey(publicKey)
	pemEncodedPub := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: x509EncodedPub})

	return string(pemEncoded), string(pemEncodedPub)
}

// CreateEcdsaSignature the message and return the base64 encoded signature
// This requires the signing private key to be set
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
	return base64.StdEncoding.EncodeToString(sig)
}

// VerifyEcdsaSignature the message using the base64 encoded signature and public key
func VerifyEcdsaSignature(message []byte, signatureBase64 string, publicKey *ecdsa.PublicKey) bool {

	var rs ECDSASignature
	signature, err := base64.StdEncoding.DecodeString(signatureBase64)
	if err != nil {
		return false
	}
	if _, err = asn1.Unmarshal(signature, &rs); err != nil {
		return false
	}

	hashed := sha256.Sum256(message)
	return ecdsa.Verify(publicKey, hashed[:], rs.R, rs.S)
}
