package messaging

// NewMessenger creates a new messenger instance
// Create a messenger instance using configuration setting:
//    "DummyMessenger" (default)
//    MQTTMessenger, requires server, login and credentials properties set
//
// config holds the messenger configuration. If no server is given, 'localhost' will be used.
func NewMessenger(messengerConfig *MessengerConfig) IMessenger {
	var m IMessenger

	if messengerConfig.Server == "" {
		messengerConfig.Server = "localhost"
	}
	if messengerConfig.Messenger == "MQTTMessenger" {
		m = NewMqttMessenger(messengerConfig)
	} else {
		m = NewDummyMessenger(messengerConfig)
	}
	return m
}
