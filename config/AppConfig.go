// Package config with configuration for IoTConnect publishers and/or subscribers
package config

import (
	"encoding/json"
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

// NodesFileSuffix to append to name of the file containing saved nodes
const NodesFileSuffix = "-nodes.json"

// UserHomeDir is the user's home folder for default config
var UserHomeDir, _ = os.UserHomeDir()

// DefaultConfigFolder for IoTConnect publisher configuration files: ~/.config/iotconnect
var DefaultConfigFolder = path.Join(UserHomeDir, ".config", "iotconnect")

// LoadAppConfig loads the application configuration from a configuration file
//
// altConfigFolder contains a alternate location for the configuration files, intended for testing.
//   Use "" for default, which is <userhome>/.config/iotconnect
// publisherID to load <publisherID>.conf
// appConfig is the object to store the application configuration using yaml. This is optional.
func LoadAppConfig(altConfigFolder string, publisherID string, appConfig interface{}) error {
	configFile := publisherID + AppConfigSuffix
	err := LoadYamlConfig(altConfigFolder, configFile, publisherID, appConfig)
	return err
}

// LoadMessengerConfig loads the message bus messenger configuration from a configuration file
//
// altConfigFolder contains a alternate location for the configuration files, intended for testing.
//   Use "" for default, which is <userhome>/.config/iotconnect
// messengerConfig is the object to store messenger configuration parameters using yaml.
func LoadMessengerConfig(altConfigFolder string, messengerConfig interface{}) error {
	publisherID := ""
	err := LoadYamlConfig(altConfigFolder, MessengerConfigFile, publisherID, messengerConfig)
	return err
}

// LoadNodes loads previously saved publisher node messages from JSON file
// altConfigFolder contains a alternate location for the configuration files, intended for testing.
//   Use "" for default, which is <userhome>/.config/iotconnect
// publisherID determines the filename: <publisherID-nodes.json>
// nodelist is the object to hold list of nodes
func LoadNodes(altConfigFolder string, publisherID string, nodelist interface{}) error {
	configFolder := altConfigFolder
	if altConfigFolder == "" {
		configFolder = DefaultConfigFolder
	}
	nodesFile := path.Join(configFolder, publisherID+NodesFileSuffix)

	jsonNodes, err := ioutil.ReadFile(nodesFile)
	if err != nil {
		log.Warningf("LoadNodes: Unable to open configuration file %s: %s", nodesFile, err)
		return err
	}
	err = json.Unmarshal(jsonNodes, nodelist)
	if err != nil {
		log.Errorf("LoadNodes: Error parsing JSON node file %s: %v", nodesFile, err)
		return err
	}
	log.Infof("Save: Node list loaded successfully from %s", nodesFile)
	return nil
}

// LoadYamlConfig parses the content of a yaml configuration file into the target object
// It performs template substitution of expressions {publisher} and {hostname}
//
// altConfigFolder contains a alternate location for the configuration files, intended for testing.
//   Use "" for default, which is <userhome>/.config/iotconnect
// filename has the name of the file to load
// publisherID is used for possible substitution use in the config file
// target is the destination object. This must have a yaml encoding set for the fields
// returns error or nil on success
func LoadYamlConfig(altConfigFolder string, filename string, publisherID string, target interface{}) error {
	configFolder := altConfigFolder
	if altConfigFolder == "" {
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

// SaveNodes saves the nodelist to a JSON file
// altConfigFolder contains a alternate location for the configuration files, intended for testing.
//   Use "" for default, which is <userhome>/.config/iotconnect
// publisherID determines the filename: <publisherID-nodes.json>
// nodelist is the object to hold list of nodes
func SaveNodes(altConfigFolder string, publisherID string, nodelist interface{}) error {
	configFolder := altConfigFolder
	if altConfigFolder == "" {
		configFolder = DefaultConfigFolder
	}
	rawNodes, err := json.MarshalIndent(nodelist, "", "  ")
	if err != nil {
		log.Errorf("Save: Error Marshalling YAML node list '%s' configuration: %v", publisherID, err)
		return err
	}
	nodesFile := path.Join(configFolder, publisherID+NodesFileSuffix)
	err = ioutil.WriteFile(nodesFile, rawNodes, 0664)
	if err != nil {
		log.Errorf("Save: Error saving node list file %s: %v", nodesFile, err)
		return err
	}
	log.Infof("Save: Node list saved successfully to %s", nodesFile)
	return nil
}
