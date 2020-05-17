package publisher

import (
	"github.com/hspaay/iotc.golang/messenger"
	"github.com/hspaay/iotc.golang/persist"
	"github.com/sirupsen/logrus"
)

// NewAppPublisher function for all the boilerplate. This:
// 1. Loads messenger config and create messenger instance
// 2. Load app config from appID.yaml
// 3. Create a publisher
// 4. Set to persist nodes and load previously saved nodes
// appID is the application ID, used as default publisher ID
// configFolder location, use "" for default location (.config/iotc)
// appConfig address for storing loaded application
// returns publisher instance
func NewAppPublisher(appID string, configFolder string, appConfig interface{}) *Publisher {
	logger := logrus.New()
	var messengerConfig = messenger.MessengerConfig{}

	persist.LoadMessengerConfig(configFolder, &messengerConfig)
	messenger := messenger.NewMessenger(&messengerConfig, logger)

	persist.LoadAppConfig(configFolder, appID, appConfig)

	pub := NewPublisher(messengerConfig.Zone, appID, messenger)

	pub.SetPersistNodes(configFolder, true)
	return pub
}
