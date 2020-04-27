// Package nodes with handling of node outputs objects
package nodes

import (
	"fmt"

	"github.com/hspaay/iotconnect.golang/messaging"
)

// Output contains logic for using the data from the input discovery message
type Output struct {
	messaging.OutputDiscoveryMessage
}

// MakeOutputDiscoveryAddress for publishing or subscribing
func MakeOutputDiscoveryAddress(zone string, publisherID string, nodeID string, ioType string, instance string) string {
	address := fmt.Sprintf("%s/%s/%s/"+messaging.MessageTypeOutputDiscovery+"/%s/%s",
		zone, publisherID, nodeID, ioType, instance)
	return address
}

// NewOutput instance
func NewOutput(node *Node, outputType string, instance string) *Output {
	address := MakeOutputDiscoveryAddress(node.Zone, node.PublisherID, node.ID, outputType, instance)
	output := &Output{
		messaging.OutputDiscoveryMessage{
			Address:    address,
			Instance:   instance,
			OutputType: outputType,
			// History:  make([]*HistoryValue, 1),
		},
	}
	return output
}
