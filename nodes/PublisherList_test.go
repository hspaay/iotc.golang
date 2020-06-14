package nodes

import (
	"encoding/base64"
	"log"
	"testing"
	"time"

	"github.com/hspaay/iotc.golang/messenger"
	"github.com/square/go-jose/json"
	"github.com/stretchr/testify/assert"
)

// const node1ID = "node1"
// const node1AliasID = "alias1"
// const publisher1ID = "publisher1"
// const publisher2ID = "publisher2"
// const domain1ID = iotc.LocalDomainID

// TestNewNode instance
func TestNewPubList(t *testing.T) {
	pubList := NewPublisherList()

	if !assert.NotNilf(t, pubList, "Failed creating node") {
		return
	}
	// 1: Test converting keys to and from pem
	privKey := messenger.CreateAsymKeys()

	privPem, pubPem := messenger.KeysToPem(privKey, &privKey.PublicKey)
	_ = privPem
	pubKey2 := messenger.PublicKeyFromPem(pubPem)
	pubPem2 := messenger.PublicKeyToPem(pubKey2)
	assert.Equal(t, pubPem, pubPem2, "Public keys should be equal")

	// 2: Test creating signature for base64url encoded list
	pubListJSON, err := json.Marshal(pubList)
	pubListBase64 := base64.URLEncoding.EncodeToString(pubListJSON)
	assert.NoError(t, err, "Failed creating json publisher list")
	sig := messenger.CreateEcdsaSignature([]byte(pubListBase64), privKey)
	assert.NotEmpty(t, sig, "", "Missing signature")

	// 3: Test signature verification
	match := messenger.VerifyEcdsaSignature(pubListBase64, sig, pubKey2)
	assert.Truef(t, match, "Verification of ECDSA signature failed")

	//--- performance tests ---

	// create public key for verification of publisher
	// used each time a publisher message is received
	start := time.Now()
	for count := 0; count < 10000; count++ {
		pubKey2 = messenger.PublicKeyFromPem(pubPem)
		_ = pubKey2
	}
	duration := time.Since(start).Seconds()
	log.Printf("10K public keys generated from PEM in %f seconds", duration)

	// create sig of base64URL encoded publisher
	start = time.Now()
	for count := 0; count < 10000; count++ {
		pubListJSON, _ := json.Marshal(pubList)
		pubListBase64 = base64.URLEncoding.EncodeToString(pubListJSON)
		sig = messenger.CreateEcdsaSignature([]byte(pubListBase64), privKey)
		_ = sig
	}
	duration = time.Since(start).Seconds()
	log.Printf("10K signatures generated from marshalled object in %f seconds", duration)

	// verify sig of base64URL encoded publisher
	start = time.Now()
	for count := 0; count < 10000; count++ {
		// verify signature
		// pubKey2 = messenger.PublicKeyFromPem(pubPem)
		match = messenger.VerifyEcdsaSignature(pubListBase64, sig, pubKey2)
	}
	duration = time.Since(start).Seconds()
	log.Printf("10K signatures verified in %f seconds", duration)
}
