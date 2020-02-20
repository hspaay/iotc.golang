// Package messenger - Interface of messengers for publishers and subscribers
package messenger

// IMessenger interface for messenger implementations
type IMessenger interface {

	// Connect the messenger.
	// This contains the last-will & testament information which is useful to inform subscribers
	//  when a publisher is unintentionally disconnected. Non MQTT busses can replace this with
	// their equivalent if available. Subscribers-only leave this empty.
	//
	// lastWillAddress optional last will & testament address for publishing device state
	//                 on accidental disconnect. Subscribers use "" to ignore.
	// lastWillValue payload to use with the last will publication
	Connect(lastWillAddress string, lastWillValue string)

	// Gracefully disconnect the messenger.
	// This will prevent the LWT publication so publishers must publish a graceful disconnect
	// message.
	Disconnect()

	// Sign and Publish a message
	// address to subscribe top as per MyZone standard
	// message object to transmit, this is an object that will be converted into a JSON and signed before being sent.
	Publish(address string, message interface{})

	// Subscribe to a message
	// address to subscribe to with support for wildcards '+' and '#'. Non MQTT busses must conver to equivalent
	// onMessage callback is invoked when a message on this address is received
	Subscribe(address string, onMessage func(address string, payload interface{}))
}
