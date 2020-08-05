package identities

import (
	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/sirupsen/logrus"
)

// PublishIdentity signs and publishes the public part of this identity using
// the given message signer.
func PublishIdentity(publicIdentity *types.PublisherIdentityMessage, signer *messaging.MessageSigner) {
	logrus.Infof("PublishIdentity: publish identity: %s", publicIdentity.Address)

	signer.PublishObject(publicIdentity.Address, true, publicIdentity, nil)
}
