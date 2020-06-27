package publisher

import "github.com/iotdomain/iotdomain-go/persist"

// PublishUpdatedInputs publishes pending updates to discovered outputs
func (publisher *Publisher) PublishUpdatedInputs() {
	updatedInputs := publisher.Inputs.GetUpdatedInputs(true)

	// publish updated input discovery
	for _, input := range updatedInputs {
		aliasAddress := publisher.getOutputAliasAddress(input.Address, "")
		publisher.logger.Infof("Publisher.PublishUpdates: publish input discovery: %s", aliasAddress)
		publisher.publishObject(aliasAddress, true, input, nil)
	}
	if len(updatedInputs) > 0 && publisher.cacheFolder != "" {
		allInputs := publisher.Inputs.GetAllInputs()
		persist.SaveInputs(publisher.cacheFolder, publisher.PublisherID(), allInputs)
	}
}
