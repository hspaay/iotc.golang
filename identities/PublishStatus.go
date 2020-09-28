package identities

import (
	"fmt"

	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/sirupsen/logrus"
)

// MakePublisherStatusAddress returns the publisher status message address
func MakePublisherStatusAddress(domain string, publisherID string) string {
	address := fmt.Sprintf("%s/%s/%s", domain, publisherID, types.MessageTypeStatus)
	return address
}

// PublishStatus publishes the publisher status value message
func PublishStatus(statusMsg *types.PublisherStatusMessage, signer *messaging.MessageSigner) {

	logrus.Infof("PublishIdentity: publish identity: %s", statusMsg.Address)

	signer.PublishObject(statusMsg.Address, true, statusMsg, nil)
}
