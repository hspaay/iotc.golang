package messenger

import (
	"github.com/sirupsen/logrus"
)

// NewMessenger creates a new messenger instance
// Create a messenger instance using configuration setting:
//    "DummyMessenger" (default)
//    MQTTMessenger, requires server, login and credentials properties set
//
// config messenger configuration.
// logger to use. nil to use the internal logger
//
func NewMessenger(messengerConfig *MessengerConfig, logger *logrus.Logger) IMessenger {
	var m IMessenger
	if logger == nil {
		logger = logrus.New()
	}

	if messengerConfig.Type == "MQTTMessenger" {
		m = NewMqttMessenger(messengerConfig, logger)
	} else {
		m = NewDummyMessenger(messengerConfig, logger)
	}
	return m
}
