// Package messenger for handling encryption and signing keys
package messenger

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"math/big"
)

// ECDSASignature ...
type ECDSASignature struct {
	R, S *big.Int
}

// CreateAsymKeys creates a asymmetric key set
// Returns a private key that contains its associated public key
func CreateAsymKeys() *ecdsa.PrivateKey {
	rng := rand.Reader
	curve := elliptic.P256()
	privKey, _ := ecdsa.GenerateKey(curve, rng)
	return privKey
}

// KeysFromPem converts PEM encoded public and private keys into a ECDSA object for use in the application
// See also EncodeKeysToPem for the opposite
func KeysFromPem(pemEncodedPriv string, pemEncodedPub string) (*ecdsa.PrivateKey, *ecdsa.PublicKey) {
	block, _ := pem.Decode([]byte(pemEncodedPriv))
	x509Encoded := block.Bytes
	privateKey, _ := x509.ParseECPrivateKey(x509Encoded)

	publicKey := PublicKeyFromPem(pemEncodedPub)
	return privateKey, publicKey
}

// KeysToPem converts a public/private key pair into their PEM encoded ascii format
// see also https://stackoverflow.com/questions/21322182/how-to-store-ecdsa-private-key-in-go
// See also DecodeKeysFromPem for the opposite
func KeysToPem(privateKey *ecdsa.PrivateKey, publicKey *ecdsa.PublicKey) (string, string) {
	x509Encoded, _ := x509.MarshalECPrivateKey(privateKey)
	pemEncoded := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: x509Encoded})

	x509EncodedPub, _ := x509.MarshalPKIXPublicKey(publicKey)
	pemEncodedPub := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: x509EncodedPub})

	return string(pemEncoded), string(pemEncodedPub)
}

// PublicKeyFromPem converts a ascii encoded public key into a ECDSA public key
func PublicKeyFromPem(pemEncodedPub string) *ecdsa.PublicKey {
	blockPub, _ := pem.Decode([]byte(pemEncodedPub))
	x509EncodedPub := blockPub.Bytes
	genericPublicKey, _ := x509.ParsePKIXPublicKey(x509EncodedPub)
	publicKey := genericPublicKey.(*ecdsa.PublicKey)

	return publicKey
}

// PublicKeyToPem converts a public key into PEM encoded ascii format
// See also PublicKeyFromPem for its counterpart
func PublicKeyToPem(publicKey *ecdsa.PublicKey) string {
	x509EncodedPub, _ := x509.MarshalPKIXPublicKey(publicKey)
	pemEncodedPub := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: x509EncodedPub})
	return string(pemEncodedPub)
}
