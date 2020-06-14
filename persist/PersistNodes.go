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

// InputsFileSuffix to append to name of the file containing saved inputs
const InputsFileSuffix = "-inputs.json"

// OutputsFileSuffix to append to name of the file containing saved output
const OutputsFileSuffix = "-outputs.json"

// LoadNodes loads previously saved publisher node messages from JSON file.
// Existing nodes are replaced if they exist in the JSON file. Custom nodes must be updated
// after lodaing nodes from file as previously saved versions will be loaded here.
//
// altConfigFolder contains a alternate location for the configuration files, intended for testing.
//   Use "" for default, which is <userhome>/.config/iotc
// publisherID determines the filename: <publisherID-nodes.json>
// nodelist is the address of a list that holds nodes
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
	log.Infof("LoadNodes: Node list loaded successfully from %s", nodesFile)
	return nil
}

// SaveNodes saves the nodelist to a JSON file
// configFolder contains the location for the configuration files
// publisherID determines the filename: <publisherID-nodes.json>
// nodelist is a list of nodes to save
func SaveNodes(configFolder string, publisherID string, nodeList interface{}) error {
	return SaveToJSON(configFolder, publisherID+NodesFileSuffix, nodeList)
}

// SaveInputs saves the inputlist to a JSON file
// configFolder contains the location for the configuration files
//   Use "" for default, which is <userhome>/.config/iotconnect
// publisherID determines the filename: <publisherID-nodes.json>
// nodelist is the object to hold list of nodes
func SaveInputs(configFolder string, publisherID string, inputList interface{}) error {
	return SaveToJSON(configFolder, publisherID+InputsFileSuffix, inputList)
}

// SaveOutputs saves the outputlist to a JSON file
// configFolder contains the location for the configuration files
//   Use "" for default, which is <userhome>/.config/iotconnect
// publisherID determines the filename: <publisherID-nodes.json>
// nodelist is the object to hold list of nodes
func SaveOutputs(configFolder string, publisherID string, outputList interface{}) error {
	return SaveToJSON(configFolder, publisherID+OutputsFileSuffix, outputList)
}

// SaveToJSON saves the given collection to a JSON file
// configFolder contains the location for the configuration files
// filename is the name to save the collection under
// nodelist is the object to hold list of nodes
func SaveToJSON(configFolder string, fileName string, collection interface{}) error {
	jsonText, err := json.MarshalIndent(collection, "", "  ")
	if err != nil {
		log.Errorf("Save: Error Marshalling JSON collection '%s': %v", fileName, err)
		return err
	}
	fullPath := path.Join(configFolder, fileName)
	err = ioutil.WriteFile(fullPath, jsonText, 0664)
	if err != nil {
		log.Errorf("Save: Error saving collection to JSON file %s: %v", fullPath, err)
		return err
	}
	log.Infof("Save: Collection saved successfully to JSON file %s", fullPath)
	return nil
}
