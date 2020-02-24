package iotzone

import (
	publisher "iotzone/publisher"
	"testing"
)

const PublisherID = "test1"

type Test1 struct {
	// iotzone.Publisher   // this makes test1 a publisher
}

func TestInitialize(t *testing.T) {
	publisher := publisher.New(publisher.LocalZoneID, PublisherID)
	if publisher == nil {
		t.Fatal("Failed to create a publisher")
	}
	t.Log("Completed TestInitialize")
}

func Terminate() {
	// publisher.log.warn("Stopping test1")
}

func Discover() {
	// Discovery of node. Discovery can be updated any time.
	// node = UpdateNode(NodeId) // add/update node with ID forecast
}

func Poll() {
	// node = publisher.GetNode(NodeId)
	// configValues = publisher.GetNodeConfigValues(node)
	// publisher.UpdateOutput(node, OutputTypeForecast, "0", forecastObject.value)

	// iotzone.UpdateOutputError(node, OutputTypeForecast, "0", "Server provided no forecast")
}

func main() {
	// RunPublisher(Initialize)
}
