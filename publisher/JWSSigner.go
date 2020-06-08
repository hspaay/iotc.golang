package publisher

import (
	"crypto/ecdsa"

	"github.com/hspaay/iotc.golang/iotc"
	"github.com/sirupsen/logrus"
	"github.com/square/go-jose"
)

// JWSSigner is a message signer using the JWS signing algorithm
type JWSSigner struct {
	alg        jose.SignatureAlgorithm
	logger     *logrus.Logger
	publishers map[string]*iotc.PublisherIdentityMessage
	privateKey *ecdsa.PrivateKey
	issuer     string // node address of the publisher that is signing the payload
	joseSigner jose.Signer
}

// Sign and serialize the given payload
// This returns the serialized payload for publication
func (signer *JWSSigner) Sign(payload []byte) (message []byte) {
	// TODO: configure signing algorithm
	// joseSigner, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.ES256, Key: signer.privateKey}, nil)
	// if err != nil {
	// 	signer.logger.Errorf("Publisher.publishMessage: signer not available: %s", err)
	// 	return
	// }
	// jsonObject, err := json.Marshal(object)
	// if err != nil {
	// 	signer.logger.Errorf("Publisher.publishMessage: Marshalling object to JSON failed: %s", err)
	// 	return
	// }
	signedObject, err := signer.joseSigner.Sign(payload)
	if err != nil {
		signer.logger.Errorf("Publisher.publishMessage: signing failed: %s", err)
	}
	// TODO: configure serialization
	serialized := signedObject.FullSerialize()
	return []byte(serialized)

}

// Verify a signed message for publication. The content depends on the signing algorithm
// This returns the deserialized signed payload with an error indicating whether the message is valid
// Even
func (signer *JWSSigner) Verify(message []byte) (payload []byte, err error) {
	jseSignature, err := jose.ParseSigned(string(message))
	if err != nil {
		return nil, err
	}
	iss := jseSignature.Signatures[0].Protected.ExtraHeaders["iss"]
	_ = iss
	payload = jseSignature.UnsafePayloadWithoutVerification()
	_, err = jseSignature.Verify(&signer.privateKey.PublicKey)
	return payload, err
}

// NewJWSSigner creates a message signer using JWS standard
// logger to use, use nil for internal logger
// issuer is the publisher address signing the message
// privateKey to use in signing
func NewJWSSigner(logger *logrus.Logger, issuer string, privateKey *ecdsa.PrivateKey) *JWSSigner {

	if logger == nil {
		logger = logrus.New()
	}
	// claims := jwt.Claims{}
	options := jose.SignerOptions{}
	options.WithHeader("iss", issuer)

	// header := make(map[jose.HeaderKey]interface{})

	joseSigner, err := jose.NewSigner(
		jose.SigningKey{
			Algorithm: jose.ES256,
			Key:       privateKey,
		}, &options)
	if err != nil {
		logger.Errorf("Publisher.publishMessage: signer not available: %s", err)
	}
	return &JWSSigner{
		logger:     logger,
		issuer:     issuer,
		privateKey: privateKey,
		joseSigner: joseSigner,
		publishers: make(map[string]*iotc.PublisherIdentityMessage),
	}
}
