// Package persist with configuration for IoTConnect publishers and/or subscribers
package persist

import (
	"encoding/json"
	"io/ioutil"
	"path"

	log "github.com/sirupsen/logrus"
)

// NodesFileSuffix to append to name of the file containing saved nodes
const NodesFileSuffix = "-nodes.json"

// LoadNodes loads previously saved publisher node messages from JSON file.
// Existing nodes are replaced if they exist in the JSON file. Custom nodes must be updated
// after lodaing nodes from file as previously saved versions will be loaded here.
//
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
