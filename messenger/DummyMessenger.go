// Package messenger - Dummy in-memory messenger for testing
package messenger

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/asn1"
	"encoding/base64"
	"encoding/json"
	"math/big"

	log "github.com/sirupsen/logrus"
)

// ECDSASignature ...
type ECDSASignature struct {
	R, S *big.Int
}

// DummyMessenger that implements IMessenger
type DummyMessenger struct {
	Logger         *log.Logger
	Publications   []*Publication
	signPrivateKey *ecdsa.PrivateKey
}

// Publication ...
type Publication struct {
	Address   string `json:"address"`
	Signature string `json:"signature"`
	Message   string `json:"message"`
}

// ECDSAsign the message and return the base64 encoded signature
// This requires the signing private key to be set
func (messenger *DummyMessenger) ECDSAsign(message []byte) string {
	if messenger.signPrivateKey == nil {
		return ""
	}
	hashed := sha256.Sum256(message)
	r, s, err := ecdsa.Sign(rand.Reader, messenger.signPrivateKey, hashed[:])
	if err != nil {
		return ""
	}
	sig, err := asn1.Marshal(ECDSASignature{r, s})
	return base64.StdEncoding.EncodeToString(sig)
}

// Connect the messenger
func (messenger *DummyMessenger) Connect(lastWillAddress string, lastWillValue string) {
}

// Disconnect gracefully disconnects the messenger
func (messenger *DummyMessenger) Disconnect() {
}

// FindPublication with the address
func (messenger *DummyMessenger) FindPublication(addr string) *Publication {
	for _, p := range messenger.Publications {
		if p.Address == addr {
			return p
		}
	}
	return nil
}

// Publish a message
func (messenger *DummyMessenger) Publish(address string, message interface{}) {
	buffer, err := json.MarshalIndent(message, " ", " ")
	signature := messenger.ECDSAsign(buffer)
	payload := Publication{
		Address:   address,
		Message:   string(buffer),
		Signature: string(signature),
	}
	if err != nil {
		messenger.Logger.Errorf("Messenger Publish: Error marshalling object on address %s' to json:", address, err)
	} else {
		messenger.Logger.Infof("Messenger Publish address=%s", address)
		messenger.Publications = append(messenger.Publications, &payload)
	}
}

// Subscribe to a message by address
func (messenger *DummyMessenger) Subscribe(address string, onMessage func(address string, payload interface{})) {
}

// NewDummyMessenger provides a messenger for messages that go no.where...
// logger to use for debug messages
func NewDummyMessenger() *DummyMessenger {
	var logger = log.New()
	logger.SetReportCaller(true) // publisher logging includes caller and file:line#

	// generate private/public key for signing
	rng := rand.Reader
	curve := elliptic.P256()
	signPrivateKey, err := ecdsa.GenerateKey(curve, rng)
	if err != nil {
		logger.Errorf("Failed to create keys for signing: ", err)
	}

	var messenger = &DummyMessenger{
		Logger:         logger,
		Publications:   make([]*Publication, 0),
		signPrivateKey: signPrivateKey,
	}
	return messenger
}
