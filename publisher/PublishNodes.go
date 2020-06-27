package publisher

import "github.com/iotdomain/iotdomain-go/persist"

// PublishUpdatedNodes publishes pending updates to discovered outputs and saves them to file
func (publisher *Publisher) PublishUpdatedNodes() {
	updatedNodes := publisher.Nodes.GetUpdatedNodes(true)

	// publish updated nodes
	for _, node := range updatedNodes {
		publisher.logger.Infof("Publisher.PublishUpdates: publish node discovery: %s", node.Address)
		publisher.publishObject(node.Address, true, node, nil)
	}
	if len(updatedNodes) > 0 && publisher.cacheFolder != "" {
		allNodes := publisher.Nodes.GetAllNodes()
		persist.SaveNodesToCache(publisher.cacheFolder, publisher.PublisherID(), allNodes)
	}

}
