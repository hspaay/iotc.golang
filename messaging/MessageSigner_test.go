package messaging_test

import (
	"crypto/ecdsa"
	"encoding/base64"
	"encoding/json"
	"log"
	"testing"
	"time"

	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/stretchr/testify/assert"
	"gopkg.in/square/go-jose.v2"
)

type TestObjectWithSender struct {
	Field1 string `json:"field1"`
	Field2 int    `json:"field2"`
	Sender string `json:"sender"`
}
type TestObjectNoSender struct {
	Address string `json:"address"`
	Field1  string `json:"field1"`
	Field2  int    `json:"field2"`
	// Sender string `json:"sender"`
}

const Pub1Address = "dom1.testpub.$identity"

var testObject = TestObjectWithSender{
	Field1: "The question",
	Field2: 42,
	Sender: Pub1Address,
}

var testObject2 = TestObjectNoSender{
	Address: "test/publisher1/node1/input1/0/$set",
	Field1:  "The answer",
	Field2:  43,
	// Sender: Pub1Address,
}

func TestEcdsaSigning(t *testing.T) {
	privKey := messaging.CreateAsymKeys()
	payload1, _ := json.Marshal(testObject)

	sig1 := messaging.CreateEcdsaSignature(payload1, privKey)
	err := messaging.VerifyEcdsaSignature(payload1, sig1, &privKey.PublicKey)
	assert.NoErrorf(t, err, "Verification of ECDSA signature failed")
	ident := &types.PublisherIdentityMessage{}
	ident.Address = "1"
	ident.Domain = "test"
	ident.PublisherID = "pub1"
	payload, _ := json.Marshal(ident)

	// sig := messaging.SignEncodeIdentity(ident, privKey)
	sig := messaging.CreateEcdsaSignature(payload, privKey)
	err = messaging.VerifyEcdsaSignature(payload, sig, &privKey.PublicKey)
	assert.NoErrorf(t, err, "Verification ECDSA signature failed")

	// error cases - test without pk
	sig = messaging.CreateEcdsaSignature(payload, nil)
	assert.Empty(t, sig, "Expected no signature without keys")
	// test with invalid payload
	err = messaging.VerifyEcdsaSignature([]byte("hello world"), sig, &privKey.PublicKey)
	assert.Error(t, err)
	// test with invalid signature
	err = messaging.VerifyEcdsaSignature(payload, "invalid sig", &privKey.PublicKey)
	assert.Error(t, err)
	// test with invalid public key
	sig = messaging.CreateEcdsaSignature(payload, privKey)
	err = messaging.VerifyEcdsaSignature(payload, sig, nil)
	assert.Error(t, err)
	newKey := messaging.CreateAsymKeys()
	err = messaging.VerifyEcdsaSignature(payload, sig, &newKey.PublicKey)
	assert.Error(t, err)

}

func TestJWSSigning(t *testing.T) {
	privKey := messaging.CreateAsymKeys()

	payload1, err := json.Marshal(testObject)
	assert.NoErrorf(t, err, "Serializing node1 failed")

	sig1, err := messaging.CreateJWSSignature(string(payload1), privKey)
	assert.NoErrorf(t, err, "signing node1 failed")
	assert.NotEmpty(t, sig1, "Signature is empty")

	sig2 := messaging.CreateEcdsaSignature(payload1, privKey)
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
		sig := messaging.CreateEcdsaSignature([]byte(payload1base64), privKey)
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
	sig := messaging.CreateEcdsaSignature([]byte(payload1base64), privKey)
	start = time.Now()
	for count := 0; count < 10000; count++ {
		// verify signature
		// pubKey2 = messenger.PublicKeyFromPem(pubPem)
		match := messaging.VerifyEcdsaSignature([]byte(payload1base64), sig, &privKey.PublicKey)
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

	// verify invalid jws mess
	messaging.VerifyJWSMessage("bad sig", &privKey.PublicKey)

}

// Test the sender verification
func TestVerifySender(t *testing.T) {
	privKey := messaging.CreateAsymKeys()

	payload1, err := json.Marshal(testObject)
	assert.NoErrorf(t, err, "Serializing node1 failed")
	sig1, err := messaging.CreateJWSSignature(string(payload1), privKey)

	var received TestObjectWithSender
	isSigned, err := messaging.VerifySenderJWSSignature(sig1, &received, func(address string) *ecdsa.PublicKey {
		// return the public key of this publisher
		return &privKey.PublicKey
	})
	assert.NoErrorf(t, err, "Verification failed")
	assert.True(t, isSigned, "Message wasn't signed")

	// using 'Address' instead of sender in the payload
	payload2, err := json.Marshal(testObject2)
	sig2, err := messaging.CreateJWSSignature(string(payload2), privKey)
	var received2 TestObjectNoSender
	isSigned, err = messaging.VerifySenderJWSSignature(sig2, &received2, func(address string) *ecdsa.PublicKey {
		// return the public key of this publisher
		return &privKey.PublicKey
	})
	assert.NoErrorf(t, err, "Verification failed")
	assert.True(t, isSigned, "Message wasn't signed")

	// using no public key lookup
	payload2, err = json.Marshal(testObject2)
	sig2, err = messaging.CreateJWSSignature(string(payload2), privKey)
	isSigned, err = messaging.VerifySenderJWSSignature(sig2, &received2, nil)
	assert.Errorf(t, err, "Verification with invalid message succeeded")
	assert.True(t, isSigned, "Message wasn't signed")

	// no public key for sender
	sig2, err = messaging.CreateJWSSignature(string(payload2), privKey)
	isSigned, err = messaging.VerifySenderJWSSignature(sig2, &received2, func(address string) *ecdsa.PublicKey {
		return nil
	})
	assert.Errorf(t, err, "Verification without public key succeeded")
	assert.True(t, isSigned, "Message wasn't signed")

	// using empty address
	testObject2.Address = ""
	payload2, err = json.Marshal(testObject2)
	sig2, err = messaging.CreateJWSSignature(string(payload2), privKey)
	isSigned, err = messaging.VerifySenderJWSSignature(sig2, &received2, nil)
	assert.Errorf(t, err, "Verification with message without Address should not succeed")
	assert.True(t, isSigned, "Message wasn't signed")

	// no sender or address
	var obj3 struct{ field1 string }
	payload3, err := json.Marshal(obj3)
	sig3, err := messaging.CreateJWSSignature(string(payload3), privKey)
	isSigned, err = messaging.VerifySenderJWSSignature(sig3, &obj3, nil)
	assert.Errorf(t, err, "Verification with message without sender should not succeed")
	assert.True(t, isSigned, "Message wasn't signed")

	// invalid message
	isSigned, err = messaging.VerifySenderJWSSignature("invalid", &received, nil)
	assert.Errorf(t, err, "Verification with invalid message should not succeed")
	// invalid payload
	payload4 := "this is not json"
	sig4, err := messaging.CreateJWSSignature(string(payload4), privKey)
	isSigned, err = messaging.VerifySenderJWSSignature(sig4, &received, nil)
	assert.Errorf(t, err, "Verification with non json payload should not succeed")

	// different public key
	newKeys := messaging.CreateAsymKeys()
	isSigned, err = messaging.VerifySenderJWSSignature(sig2, &received, func(address string) *ecdsa.PublicKey {
		// return the public key of this publisher
		return &newKeys.PublicKey
	})
	assert.Errorf(t, err, "Verification with wrong publickey should not succeed")
}

func TestSigner(t *testing.T) {
	const payload1 = "payload 1"
	const payload2 = "payload 2"
	var isEncrypted bool
	var isSigned bool
	var err error
	var received = ""

	config := messaging.MessengerConfig{}
	messenger := messaging.NewDummyMessenger(&config)
	privKey := messaging.CreateAsymKeys()
	getPubKey := func(address string) *ecdsa.PublicKey {
		return &privKey.PublicKey
	}
	signer := messaging.NewMessageSigner(true, getPubKey, messenger, privKey)

	handler := func(address string, rawMessage string) error {
		obj := TestObjectWithSender{}
		isEncrypted, isSigned, err = signer.DecodeMessage(rawMessage, &obj)
		assert.True(t, isSigned, "expected message to be signed")
		assert.NoErrorf(t, err, "%s", err)
		received = obj.Field1

		// additional sign test
		if !isEncrypted {
			isSigned2, err2 := signer.VerifySignedMessage(rawMessage, &obj)
			assert.NoErrorf(t, err2, "%s", err2)
			assert.Equal(t, isSigned, isSigned2, "Expected signature verification to match")
		}
		return nil
	}
	signer.SetSignMessages(true)
	signer.Subscribe("test/+/#", handler)

	obj := TestObjectWithSender{}
	obj.Field1 = payload1
	obj.Sender = "c'est moi"
	err = signer.PublishObject("test/bob/james", false, obj, &privKey.PublicKey)
	assert.NoError(t, err, "Error publishing encrypted object")
	assert.Equal(t, payload1, received)
	assert.True(t, isSigned, "Message not signed")
	assert.True(t, isEncrypted, "Not encrypted")

	// publish without encryption
	obj.Field1 = payload2
	err = signer.PublishObject("test/bob/james", false, obj, nil)
	assert.NoError(t, err, "Error publishing signing object")
	assert.Equal(t, payload2, received)
	assert.True(t, isSigned, "Message not signed")
	assert.False(t, isEncrypted, "Message is encrypted")

	// publish with error
	err = signer.PublishObject("test/bob/james", false, nil, &privKey.PublicKey)
	assert.Error(t, err, "No error publishing string as object")

	signer.Unsubscribe("test/+/#", nil)
}

func TestSignIdentity(t *testing.T) {
	dssKeys := messaging.CreateAsymKeys()
	newIdent := types.PublisherFullIdentity{}
	newIdent.IssuerName = "dss"
	newIdent.Organization = "iotdomain.org"
	newIdent.IdentitySignature = ""
	identSignature := messaging.CreateIdentitySignature(&newIdent.PublisherIdentityMessage, dssKeys)
	assert.NotNil(t, identSignature, "Signing identity fails")

	// verify the signature
	err := messaging.VerifyIdentitySignature(&newIdent.PublisherIdentityMessage, &dssKeys.PublicKey)
	assert.NotNil(t, err)
}
