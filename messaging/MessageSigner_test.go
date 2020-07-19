package messaging_test

import (
	"crypto/ecdsa"
	"encoding/base64"
	"encoding/json"
	"log"
	"testing"
	"time"

	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/stretchr/testify/assert"
	"gopkg.in/square/go-jose.v2"
)

type TestObjectWithSender struct {
	Field1 string `json:"field1"`
	Field2 int    `json:"field2"`
	Sender string `json:"sender"`
}
type TestObjectNoSender struct {
	Field1 string `json:"field1"`
	Field2 int    `json:"field2"`
	// Sender string `json:"sender"`
}

const Pub1Address = "dom1.testpub.$identity"

var testObject = TestObjectWithSender{
	Field1: "The answer",
	Field2: 42,
	Sender: Pub1Address,
}

var testObject2 = TestObjectNoSender{
	Field1: "The answer",
	Field2: 42,
	// Sender: Pub1Address,
}

func TestEcdsaSigning(t *testing.T) {
	privKey := messaging.CreateAsymKeys()
	payload1, _ := json.Marshal(testObject)

	sig1 := messaging.CreateEcdsaSignature(string(payload1), privKey)
	match := messaging.VerifyEcdsaSignature(string(payload1), sig1, &privKey.PublicKey)
	assert.Truef(t, match, "Verification of ECDSA signature failed")
}

func TestJWSSigning(t *testing.T) {
	privKey := messaging.CreateAsymKeys()

	payload1, err := json.Marshal(testObject)
	assert.NoErrorf(t, err, "Serializing node1 failed")

	sig1, err := messaging.CreateJWSSignature(string(payload1), privKey)
	assert.NoErrorf(t, err, "signing node1 failed")
	assert.NotEmpty(t, sig1, "Signature is empty")

	sig2 := messaging.CreateEcdsaSignature(string(payload1), privKey)
	assert.NotEqual(t, sig1, sig2, "JWS Signature doesn't match with Ecdsa")
}

func TestSigningPerformance(t *testing.T) {
	privKey := messaging.CreateAsymKeys()

	payload1, err := json.Marshal(testObject)
	assert.NoErrorf(t, err, "Serializing node1 failed")

	// create sig of base64URL encoded publisher
	start := time.Now()
	for count := 0; count < 10000; count++ {
		payload1base64 := base64.URLEncoding.EncodeToString(payload1)
		sig := messaging.CreateEcdsaSignature(payload1base64, privKey)
		_ = sig
	}
	duration := time.Since(start).Seconds()
	log.Printf("10K CreateEcdsaSignature signatures generated in %f seconds", duration)

	// Create JWS signature using JOSE directly
	start = time.Now()
	joseSigner, _ := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.ES256, Key: privKey}, nil)
	for count := 0; count < 10000; count++ {
		jws, _ := joseSigner.Sign([]byte(payload1))
		payload, _ := jws.CompactSerialize()
		_ = payload
	}
	duration = time.Since(start).Seconds()
	log.Printf("10K JoseSigner signatures generated in %.2f seconds", duration)

	// generate JWS signature using my lib
	start = time.Now()
	for count := 0; count < 10000; count++ {
		messaging.CreateJWSSignature(string(payload1), privKey)
	}
	duration = time.Since(start).Seconds()
	log.Printf("10K CreateJWSSignature signatures generated in %.2f seconds", duration)

	// verify sig of base64URL encoded payload
	payload1base64 := base64.URLEncoding.EncodeToString(payload1)
	sig := messaging.CreateEcdsaSignature(payload1base64, privKey)
	start = time.Now()
	for count := 0; count < 10000; count++ {
		// verify signature
		// pubKey2 = messenger.PublicKeyFromPem(pubPem)
		match := messaging.VerifyEcdsaSignature(payload1base64, sig, &privKey.PublicKey)
		_ = match
	}
	duration = time.Since(start).Seconds()
	log.Printf("10K VerifyEcdsaSignature signatures verified in %f seconds", duration)

	// verify JWS signature using my lib
	sig1, err := messaging.CreateJWSSignature(string(payload1), privKey)
	start = time.Now()
	for count := 0; count < 10000; count++ {
		messaging.VerifyJWSMessage(sig1, &privKey.PublicKey)
	}
	duration = time.Since(start).Seconds()
	log.Printf("10K VerifyJWSMessage signatures verified in %.2f seconds", duration)

}

// Test the verification a
// This tests the
func TestVerifyPublisher(t *testing.T) {
	privKey := messaging.CreateAsymKeys()

	payload1, err := json.Marshal(testObject)
	assert.NoErrorf(t, err, "Serializing node1 failed")
	sig1, err := messaging.CreateJWSSignature(string(payload1), privKey)

	//
	var received TestObjectWithSender

	isSigned, err := messaging.VerifySenderSignature(sig1, &received, func(address string) *ecdsa.PublicKey {
		// return the public key of this publisher
		return &privKey.PublicKey
	})
	assert.NoErrorf(t, err, "Verification failed")
	assert.True(t, isSigned, "Message wasn't signed")
}

// Test the sender verification
func TestVerifySender(t *testing.T) {
	privKey := messaging.CreateAsymKeys()

	payload1, err := json.Marshal(testObject)
	assert.NoErrorf(t, err, "Serializing node1 failed")
	sig1, err := messaging.CreateJWSSignature(string(payload1), privKey)

	//
	var received TestObjectWithSender

	isSigned, err := messaging.VerifySenderSignature(sig1, &received, func(address string) *ecdsa.PublicKey {
		// return the public key of this publisher
		return &privKey.PublicKey
	})
	assert.NoErrorf(t, err, "Verification failed")
	assert.True(t, isSigned, "Message wasn't signed")
}
