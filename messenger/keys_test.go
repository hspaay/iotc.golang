package messenger

import (
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestKeys tests public/private key conversion to and from pem
func TestKeysFromPem(t *testing.T) {

	privKey := CreateAsymKeys()

	privPem := PrivateKeyToPem(privKey)
	_ = privPem
	privKey2 := PrivateKeyFromPem(privPem)
	assert.Equal(t, privKey, privKey2, "Private keys should be equal")

	// create public key for verification of publisher
	// used each time a publisher message is received
	pubPem := PublicKeyToPem(&privKey.PublicKey)
	start := time.Now()
	for count := 0; count < 10000; count++ {
		pubKey2 := PublicKeyFromPem(pubPem)
		_ = pubKey2
	}
	duration := time.Since(start).Seconds()
	log.Printf("10K public keys generated from PEM in %f seconds", duration)

}
