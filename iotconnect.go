package iotconnect

import (
	"iotconnect/messenger"
	"iotconnect/publisher"
	"iotconnect/standard"
)

const publisherID = "test1"

func main() {
	messenger := messenger.NewDummyMessenger()
	publisher := publisher.NewPublisher(standard.LocalZoneID, publisherID, messenger)
	publisher.Start(false, nil, nil)
}
