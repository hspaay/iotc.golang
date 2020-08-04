package inputs

import (
	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/sirupsen/logrus"
)

// PublishRegisteredInputs publishes input discovery messages
// This will clear the updated inputs list
func PublishRegisteredInputs(
	inputs []*types.InputDiscoveryMessage,
	messageSigner *messaging.MessageSigner) {

	// publish updated registered inputs
	for _, input := range inputs {
		logrus.Infof("PublishRegisteredInputs: publish input discovery: %s", input.Address)
		// no encryption as this is for everyone to see
		messageSigner.PublishObject(input.Address, true, input, nil)
	}
	// todo move save input configuration
	// if len(updatedInputs) > 0 && publisher.cacheFolder != "" {
	// 	allInputs := inpub.registeredInputs.GetAllInputs()
	// 	persist.SaveInputs(publisher.cacheFolder, publisher.PublisherID(), allInputs)
	// }
}
