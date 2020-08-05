package publisher

import (
	"reflect"

	"github.com/iotdomain/iotdomain-go/lib"
	"github.com/iotdomain/iotdomain-go/messaging"
)

// NewAppPublisher function for all the boilerplate. This:
//  1. Loads messenger config and create messenger instance
//  2. Load app config from <appID>.yaml and extract field PublisherID
//  3. Create a publisher using the domain from messenger config and publisherID from <appID>.yaml
//  4. Set to persist nodes and load previously saved nodes
//
// 'appID' is the application ID, used as publisher ID unless overridden in <appID>.yaml. The 'configFolder'
// location contains the messenger and application configuration files. Use "" for default location (.config/iotc)
//
// The cache folder location contains saved publisher, nodes, inputs and outputs, use "" for default location (.cache/iotc)
// appConfig optional application object to load <appID>.yaml configuration into. If it contains
//   a field named 'PublisherID' it will allow override the default publisherID.
// persistNodes flags whether to save discovered nodes and their configuration changes.
//
// This returns publisher instance or error if messenger fails to load
func NewAppPublisher(appID string, configFolder string, cacheFolder string,
	appConfig interface{}, persistNodes bool) (*Publisher, error) {

	var messengerConfig = messaging.MessengerConfig{}
	err := lib.LoadMessengerConfig(configFolder, &messengerConfig)
	messenger := messaging.NewMessenger(&messengerConfig)

	// appconfig is optional
	// The publisherID can be overridden from the appConfig yaml file
	if appConfig != nil {
		lib.LoadAppConfig(configFolder, appID, appConfig)
	}
	ac := reflect.ValueOf(appConfig)
	field := reflect.Indirect(ac).FieldByName("PublisherID")
	pubID := field.String()
	if !field.IsValid() || pubID == "" {
		pubID = appID
	}
	// identity lives in the config folder
	pub := NewPublisher(messengerConfig.Domain, pubID, configFolder, false, messenger)

	// Load configuration of previously registered nodes from config
	pub.LoadRegisteredNodes()

	// Load discovered domain publishers from cache
	pub.LoadDomainIdentities()
	return pub, err
}
