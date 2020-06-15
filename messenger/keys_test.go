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

	privPem, pubPem := KeysToPem(privKey, &privKey.PublicKey)
	_ = privPem
	pubKey2 := PublicKeyFromPem(pubPem)
	pubPem2 := PublicKeyToPem(pubKey2)
	assert.Equal(t, pubPem, pubPem2, "Public keys should be equal")

	// create public key for verification of publisher
	// used each time a publisher message is received
	start := time.Now()
	for count := 0; count < 10000; count++ {
		pubKey2 = PublicKeyFromPem(pubPem)
		_ = pubKey2
	}
	duration := time.Since(start).Seconds()
	log.Printf("10K public keys generated from PEM in %f seconds", duration)

}
