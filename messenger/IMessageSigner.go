// Package messenger - Interface of messengers for publishers and subscribers
package messenger

// IMessageSigner interface for signing and verification of published messages
type IMessageSigner interface {
	// Sign the given payload
	// This returns the serialized message for publication. The content depends on the signing algorithm
	Sign(payload []byte) (message []byte)

	// Verify a signed message for publication. The content depends on the signing algorithm
	// This returns the deserialized payload
	Verify(message []byte) (payload []byte, err error)
}
