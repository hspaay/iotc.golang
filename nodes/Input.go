// Package nodes with handling of inputs
package nodes

import (
	"fmt"

	"github.com/hspaay/iotc.golang/iotc"
)

// Input contains logic for using the data from the input discovery message
type Input struct {
	iotc.InputDiscoveryMessage
}

// MakeInputDiscoveryAddress creates the address for the input discovery
func MakeInputDiscoveryAddress(zone string, publisherID string, nodeID string, inputType string, instance string) string {
	address := fmt.Sprintf("%s/%s/%s/"+iotc.MessageTypeInputDiscovery+"/%s/%s",
		zone, publisherID, nodeID, inputType, instance)
	return address
}

// MakeInputSetAddress creates the address used to update an input value
func MakeInputSetAddress(zone string, publisherID string, nodeID string, ioType string, instance string) string {
	address := fmt.Sprintf("%s/%s/%s/"+iotc.MessageTypeSet+"/%s/%s",
		zone, publisherID, nodeID, ioType, instance)
	return address
}

// NewInput instance
func NewInput(node *Node, inputType string, instance string) *Input {
	address := MakeInputDiscoveryAddress(node.Zone, node.PublisherID, node.ID, inputType, instance)
	input := &Input{
		iotc.InputDiscoveryMessage{
			Address:   address,
			Instance:  instance,
			InputType: inputType,
			// History:  make([]*HistoryValue, 1),
		},
	}
	return input
}
