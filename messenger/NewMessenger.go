package messenger

import (
	"github.com/sirupsen/logrus"
)

// NewMessenger creates a new messenger instance
// Create a messenger instance using configuration setting:
//    "DummyMessenger" (default)
//    MQTTMessenger, requires server, login and credentials properties set
//
// config holds the messenger configuration. If no server is given, 'localhost' will be used.
// logger is optional in case you want to use a predefined logging. Default is a default logger.
func NewMessenger(messengerConfig *MessengerConfig, logger *logrus.Logger) IMessenger {
	var m IMessenger
	if logger == nil {
		logger = logrus.New()
	}

	if messengerConfig.Server == "" {
		messengerConfig.Server = "localhost"
	}
	if messengerConfig.Messenger == "MQTTMessenger" {
		m = NewMqttMessenger(messengerConfig, logger)
	} else {
		m = NewDummyMessenger(messengerConfig, logger)
	}
	return m
}
