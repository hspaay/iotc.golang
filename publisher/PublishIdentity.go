// Package publisher with publishing of this publisher's identity
package publisher

import (
	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/sirupsen/logrus"
)

// PublishIdentity publishes this publisher's identity on startup or update
func (publisher *Publisher) PublishIdentity(signer messaging.MessageSigner) {
	identity := publisher.fullIdentity
	logrus.Infof("Publisher.PublishIdentity: publish identity: %s", publisher.fullIdentity.Address)
	signer.PublishObject(identity.Address, true, identity.PublisherIdentityMessage, nil)
}
