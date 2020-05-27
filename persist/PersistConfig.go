// Package persist with configuration for IoTConnect publishers and/or subscribers
package persist

import (
	"io/ioutil"
	"os"
	"path"
	"strings"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// AppConfigSuffix to append to the publisher ID to load the application configuration
const AppConfigSuffix = ".yaml"

// MessengerConfigFile is the filename of the shared messenger configuration used by all publishers
const MessengerConfigFile = "messenger.yaml"

// UserHomeDir is the user's home folder for default config
var UserHomeDir, _ = os.UserHomeDir()

// DefaultConfigFolder for IoTConnect publisher configuration files: ~/.config/iotc
var DefaultConfigFolder = path.Join(UserHomeDir, ".config", "iotc")

// LoadAppConfig loads the application configuration from a configuration file
//
// configFolder contains the location for the configuration files.
// appID to load from file <appID>.yaml
// appConfig is the object to store the application configuration using yaml. This is optional.
func LoadAppConfig(configFolder string, appID string, appConfig interface{}) error {
	configFile := appID + AppConfigSuffix
	err := LoadYamlConfig(configFolder, configFile, appID, appConfig)
	return err
}

// LoadMessengerConfig loads the message bus messenger configuration from a configuration file
//
// configFolder location of configuration files. Use persist.DefaultConfigFolder for default
// messengerConfig is the object to store messenger configuration parameters using yaml.
func LoadMessengerConfig(configFolder string, messengerConfig interface{}) error {
	publisherID := ""
	err := LoadYamlConfig(configFolder, MessengerConfigFile, publisherID, messengerConfig)
	return err
}

// LoadYamlConfig parses the content of a yaml configuration file into the target object
// It performs template substitution of expressions {publisher} and {hostname}
//
// altConfigFolder contains the location for the configuration files.
//   Use "" for default, which is <userhome>/.config/iotconnect
// filename has the name of the file to load
// publisherID is used for possible substitution use in the config file
// target is the destination object. This must have a yaml encoding set for the fields
// returns error or nil on success
func LoadYamlConfig(configFolder string, filename string, publisherID string, target interface{}) error {
	if configFolder == "" {
		configFolder = DefaultConfigFolder
	}
	fullPath := path.Join(configFolder, filename)

	rawConfig, err := ioutil.ReadFile(fullPath)
	if err != nil {
		log.Warningf("LoadYamlConfig: Unable to open configuration file %s: %s", filename, err)
		return err
	}
	submap := make(map[string]string, 0)
	submap["hostname"], _ = os.Hostname()
	if publisherID != "" {
		submap["publisher"] = publisherID
	}

	substituted := string(rawConfig)
	for key, val := range submap {
		substituted = strings.ReplaceAll(substituted, "{"+key+"}", val)
	}

	err = yaml.Unmarshal([]byte(substituted), target)
	if err != nil {
		log.Errorf("UnmarshalConfigFile: Error parsing YAML configuration file %s: %v", filename, err)
		return err
	}
	return nil
}
