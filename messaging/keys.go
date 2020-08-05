// Package messaging for handling encryption and signing keys
package messaging

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

// PrivateKeyFromPem converts PEM encoded private keys into a ECDSA object for use in the application
// See also PrivateKeyToPem for the opposite.
// Returns nil if the encoded pem source isn't a pem format
func PrivateKeyFromPem(pemEncodedPriv string) *ecdsa.PrivateKey {
	if pemEncodedPriv == "" {
		return nil
	}
	block, _ := pem.Decode([]byte(pemEncodedPriv))
	x509Encoded := block.Bytes
	privateKey, _ := x509.ParseECPrivateKey(x509Encoded)

	return privateKey
}

// PrivateKeyToPem converts a private key into their PEM encoded ascii format
// see also https://stackoverflow.com/questions/21322182/how-to-store-ecdsa-private-key-in-go
func PrivateKeyToPem(privateKey *ecdsa.PrivateKey) string {
	x509Encoded, _ := x509.MarshalECPrivateKey(privateKey)
	pemEncoded := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: x509Encoded})

	return string(pemEncoded)
}

// PublicKeyFromPem converts a ascii encoded public key into a ECDSA public key
func PublicKeyFromPem(pemEncodedPub string) *ecdsa.PublicKey {
	if pemEncodedPub == "" {
		return nil
	}
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
