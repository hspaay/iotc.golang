package messaging_test

import (
	"log"
	"testing"
	"time"

	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/stretchr/testify/assert"
)

// TestKeys tests public/private key conversion to and from pem
func TestKeysFromPem(t *testing.T) {

	privKey := messaging.CreateAsymKeys()

	privPem := messaging.PrivateKeyToPem(privKey)
	_ = privPem
	privKey2 := messaging.PrivateKeyFromPem(privPem)
	assert.Equal(t, privKey, privKey2, "Private keys should be equal")

	// error cases
	privKey3 := messaging.PrivateKeyFromPem("")
	assert.Nil(t, privKey3)

	pubKey3 := messaging.PublicKeyFromPem("")
	assert.Nil(t, pubKey3)

	// create public key for verification of publisher
	// used each time a publisher message is received
	pubPem := messaging.PublicKeyToPem(&privKey.PublicKey)
	start := time.Now()
	for count := 0; count < 10000; count++ {
		pubKey2 := messaging.PublicKeyFromPem(pubPem)
		_ = pubKey2
	}
	duration := time.Since(start).Seconds()
	log.Printf("10K public keys generated from PEM in %f seconds", duration)

}
