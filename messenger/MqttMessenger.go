// Package messenger - Publish and Subscribe to message using the MQTT message bus
package messenger

// MqttMessenger that implements IMessenger
type MqttMessenger struct {
}

// NewMqttMessenger provides a messager for the MQTT message bus
// url with host and port of the server, eg "mqtt://mqtt.iotzone.network:5883/"
// login name: zone/id
// cred with password
// zone namespace for multi-tenant message busses
func NewMqttMessenger(url string, login string, cred string, zone string) *MqttMessenger {
	mqttMessenger := &MqttMessenger{}
	return mqttMessenger
}

// Connect the messenger
func (messenger *MqttMessenger) Connect(lastWillAddress string, lastWillValue string) {
}

// Disconnect gracefully disconnects the messenger
func (messenger *MqttMessenger) Disconnect() {
}

// Publish a message
func (messenger *MqttMessenger) Publish(address string, payload struct{}) {
}

// Subscribe to a message by address
func (messenger *MqttMessenger) Subscribe(address string, onMessage func(address string, payload interface{})) {
}
