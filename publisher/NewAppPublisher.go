package publisher

import (
	"github.com/iotdomain/iotdomain-go/lib"
	"github.com/iotdomain/iotdomain-go/messaging"
)

// PublisherConfig defined configuration fields read from the application configuration
type PublisherConfig struct {
	Domain      string `yaml:"domain"` // optional override per publisher
	PublisherID string `yaml:"publisherId"`
	Loglevel    string `yaml:"loglevel"` // error, warning, info, debug
	Logfile     string `yaml:"logfile"`  //
}

// NewAppPublisher function for all the boilerplate. This:
//  1. Loads messenger config and create messenger instance
//  2. Load appconfig from <appID>.yaml and extract common fields:
//     loglevel, logfile and publisherId (see struct PublisherConfig)
//  3. Load appconfig into the application struct provided through appConfig
//  4. Create a publisher using the domain from messenger config and publisherID from <appID>.yaml
//  5. Set to persist nodes and load previously saved nodes
//
// 'appID' is the application ID, used as publisher ID unless overridden in <appID>.yaml. The 'configFolder'
// location contains the messenger and application configuration files. Use "" for default location (.config/iotc)
//
// The cache folder location contains saved publisher, nodes, inputs and outputs, use "" for default location (.cache/iotc)
//  appConfig optional application object to load <appID>.yaml configuration into. If it contains
//  persistNodes flag indicates whether to save discovered nodes and their configuration changes.
//
// This returns publisher instance or error if messenger fails to load
func NewAppPublisher(appID string, configFolder string, cacheFolder string,
	appConfig interface{}, persistNodes bool) (*Publisher, error) {

	// 1: load messenger config shared with other publishers
	var messengerConfig = messaging.MessengerConfig{}
	err := lib.LoadMessengerConfig(configFolder, &messengerConfig)
	messenger := messaging.NewMessenger(&messengerConfig)

	// 2: load Publisher config fields from appconfig
	pubConfig := PublisherConfig{}
	pubConfig.Loglevel = "warning"
	pubConfig.PublisherID = appID
	pubConfig.Domain = messengerConfig.Domain
	lib.LoadAppConfig(configFolder, appID, &pubConfig)

	if appConfig != nil {
		lib.LoadAppConfig(configFolder, appID, appConfig)
	}
	pub := NewPublisher(pubConfig.Domain, pubConfig.PublisherID, configFolder, false, messenger)
	pub.SetLogging(pubConfig.Loglevel, pubConfig.Logfile)

	// Load configuration of previously registered nodes from config
	pub.LoadRegisteredNodes()

	// Load discovered domain publishers from cache
	pub.LoadDomainIdentities()
	return pub, err
}
