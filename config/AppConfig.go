// Package config with configuration for IoTConnect publishers and/or subscribers
package config

import (
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/hspaay/iotconnect.golang/messenger"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// IotConnectConfig is the filename of the shared configuration used by all publishers
const IotConnectConfig = "iotconnect.conf"

// UserHomeDir is the user's home folder for default config
var UserHomeDir, _ = os.UserHomeDir()

// DefaultConfigFolder for IoTConnect publisher configuration files
var DefaultConfigFolder = path.Join(UserHomeDir, "bin", "iotconnect", "config")

// // AppConfig with combined configuration for IoTConnect applications
// type AppConfig struct {
// 	PublisherID  string          // The publisher ID used as filename publisherID.conf
// 	ConfigFolder string          // folder with all configuration files
// 	Logger       log.Logger      // The application logger to use
// 	Messenger    MessengerConfig // Messenger configuration
// 	// PublisherConfig interface{}     // publisher whose configuration to load
// }

// LoadAppConfig loads the application configuration from a configuration file named <publisherId>.conf in
// the 'DefaultConfigFolder' config folder.
//
// altConfigFolder contains a alternate location for the configuration files.
//     This is optional for non-default folders. Use "" for default: <userhome>/bin/iotconnect/conf
// publisherID to load <publisherID>.conf
// messengerConfig is the object to store messenger configuration parameters using yaml. This is optional.
// appConfig is the object to store the application configuration using yaml. This is optional.
func LoadAppConfig(altConfigFolder string, publisherID string, messengerConfig *messenger.MessengerConfig, appConfig interface{}) error {
	configFolder := altConfigFolder
	if altConfigFolder == "" {
		configFolder = DefaultConfigFolder
	}
	// first load the global iotconnect messenger config
	iotcConfigFile := path.Join(configFolder, IotConnectConfig)
	submap := make(map[string]string, 0)
	hostname, _ := os.Hostname()
	submap["hostname"] = hostname
	submap["publisher"] = publisherID

	var err error
	if messengerConfig != nil {
		err = LoadYamlConfig(iotcConfigFile, submap, messengerConfig)
	}
	if appConfig != nil {
		pubConfigFile := path.Join(configFolder, publisherID+".conf")
		err = LoadYamlConfig(pubConfigFile, submap, appConfig)
	}
	return err
}

// SaveAppConfig saves the application configuration
func SaveAppConfig(altConfigFolder string, publisherID string, appConfig interface{}) error {
	configFolder := altConfigFolder
	if altConfigFolder == "" {
		configFolder = DefaultConfigFolder
	}
	rawConfig, err := yaml.Marshal(appConfig)
	if err != nil {
		log.Errorf("Save: Error Marshalling YAML publisher '%s' configuration: %v", publisherID, err)
		return err
	}
	pubConfigFile := path.Join(configFolder, publisherID+".conf")
	err = ioutil.WriteFile(pubConfigFile, rawConfig, 0664)
	if err != nil {
		log.Errorf("Save: Error saving publisher YAML configuration file %s: %v", pubConfigFile, err)
		return err
	}
	log.Infof("Save: publisher YAML configuration file %s saved successfully", pubConfigFile)
	return nil
}

// LoadYamlConfig parses the content of a yaml configuration file into the target object
// It performs template substitution of expressions {publisher} and {hostname}
// filename contains the path to the file to load
// substitution is a map for template substitution for use in the config file
// target is the destination object. This must have a yaml parser encoding
// returns nil or error if file not found or parsing error
func LoadYamlConfig(filename string, substitution map[string]string, target interface{}) error {
	rawConfig, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Warningf("UnmarshalConfigFile: Unable to open configuration file %s: %s", filename, err)
		return err
	}
	substituted := string(rawConfig)
	for key, val := range substitution {
		substituted = strings.ReplaceAll(substituted, "{"+key+"}", val)
	}

	err = yaml.Unmarshal([]byte(substituted), target)
	if err != nil {
		log.Errorf("UnmarshalConfigFile: Error parsing YAML configuration file %s: %v", filename, err)
		return err
	}
	return nil
}
