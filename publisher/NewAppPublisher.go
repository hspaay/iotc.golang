package publisher

import (
	"github.com/iotdomain/iotdomain-go/lib"
	"github.com/iotdomain/iotdomain-go/messaging"
)

// NewAppPublisher function for all the boilerplate. This:
//  1. Loads messenger config and create messenger instance
//  2. Load PublisherConfig from <appID>.yaml
//  3. Load appconfig from <appID>.yaml (yes same file)
//  4. Create a publisher using the domain from messenger config and publisherID from <appID>.yaml
//  5. Set to persist nodes and load previously saved nodes
//
//  'appID' is the application ID, used as publisher ID unless overridden in <appID>.yaml. The 'configFolder'
// location contains the messenger and application configuration files. Use "" for default location (.config/iotc)
//  configFolder location contains saved publisher, nodes, inputs and outputs, use "" for default
// location (.config/iotdomain).
//  appConfig optional application object to load <appID>.yaml configuration into
//  autosave automatically saves discovered identities and node configuration.
// This returns publisher instance or error if messenger fails to load
func NewAppPublisher(appID string, configFolder string,
	appConfig interface{}, autosave bool) (*Publisher, error) {

	// 1: load messenger config shared with other publishers
	var messengerConfig = messaging.MessengerConfig{}
	err := lib.LoadMessengerConfig(configFolder, &messengerConfig)
	messenger := messaging.NewMessenger(&messengerConfig)

	// 2: load Publisher config fields from appconfig
	pubConfig := &PublisherConfig{
		Autosave:     autosave,
		ConfigFolder: configFolder,
		Loglevel:     "warning",
		Domain:       messengerConfig.Domain,
		PublisherID:  appID,
	}
	lib.LoadAppConfig(configFolder, appID, &pubConfig)

	// 3: load application configuration itself
	if appConfig != nil {
		lib.LoadAppConfig(configFolder, appID, appConfig)
	}
	// 4: create the publisher. Reload its identity if available.
	pub := NewPublisher(pubConfig, messenger)

	// Load configuration of previously registered nodes from config
	pub.LoadRegisteredNodes()

	// Load discovered domain publishers from cache
	pub.LoadDomainIdentities()
	return pub, err
}
