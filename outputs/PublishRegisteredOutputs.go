// Package outputs with publication of discovery information of registered outputs that have been updated
package outputs

import (
	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/sirupsen/logrus"
)

// PublishRegisteredOutputs publishes output discovery messages
func PublishRegisteredOutputs(
	outputs []*types.OutputDiscoveryMessage,
	messageSigner *messaging.MessageSigner) {

	// publish updated output discovery
	for _, output := range outputs {
		logrus.Infof("PublishRegisteredOutputs: publish output discovery for: %s", output.Address)
		messageSigner.PublishObject(output.Address, true, output, nil)
	}
	// todo: move save output configuration
	// if len(outputs) > 0 && publisher.cacheFolder != "" {
	// 	allOutputs := publisher.Outputs.GetAllOutputs()
	// 	persist.SaveOutputs(publisher.cacheFolder, publisher.PublisherID(), allOutputs)
	// }
}
