package publisher

import (
	"reflect"

	"github.com/hspaay/iotc.golang/messenger"
	"github.com/hspaay/iotc.golang/persist"
	"github.com/sirupsen/logrus"
)

// NewAppPublisher function for all the boilerplate. This:
// 1. Loads messenger config and create messenger instance
// 2. Load app config from appID.yaml and extract optional field PublisherID
// 3. Create a publisher using Zone from messenger config
// 4. Set to persist nodes and load previously saved nodes
//
// appID is the application ID, used as default publisher ID
// configFolder location, use "" for default location (.config/iotc)
// appConfig address for storing loaded application, use "" if not to load an appConfig
//    if appConfig has a field named 'PublisherID' it will be used instead of appID
// persistNodes flags whether to save discovered nodes and their configuration changes
// returns publisher instance or error if messenger fails to load
func NewAppPublisher(appID string, configFolder string, appConfig interface{}, persistNodes bool) (*Publisher, error) {
	logger := logrus.New()
	var messengerConfig = messenger.MessengerConfig{}

	err := persist.LoadMessengerConfig(configFolder, &messengerConfig)
	messenger := messenger.NewMessenger(&messengerConfig, logger)

	// appconfig is optional
	if appConfig != nil {
		persist.LoadAppConfig(configFolder, appID, appConfig)
	}
	ac := reflect.ValueOf(appConfig)
	field := reflect.Indirect(ac).FieldByName("PublisherID")
	pubID := field.String()
	if pubID == "" {
		pubID = appID
	}
	pub := NewPublisher("", messengerConfig.Domain, pubID, messenger)

	pub.SetPersistNodes(configFolder, persistNodes)
	return pub, err
}
