// Package publisher with handling of input commands
package publisher

import (
	"crypto/ecdsa"
	"strings"
	"time"

	"github.com/iotdomain/iotdomain-go/types"
)

// PublishConfigureNode updates the configuration of a remote node by this publisher
// The signed message will be encrypted with the given encryption key
func (publisher *Publisher) PublishConfigureNode(remoteNodeAddress string, attr types.NodeAttrMap, encryptionKey *ecdsa.PublicKey) {
	publisher.logger.Infof("PublishSetConfigure: publishing encrypted configuration to %s", remoteNodeAddress)
	// Check that address is one of our inputs
	segments := strings.Split(remoteNodeAddress, "/")
	// a full address is required
	if len(segments) < 4 {
		return
	}
	// domain/publisherID/nodeID/$configure
	segments[3] = types.MessageTypeConfigure
	configAddr := strings.Join(segments, "/")

	// Encecode the SetMessage
	timeStampStr := time.Now().Format("2006-01-02T15:04:05.000-0700")
	var configureMessage = types.NodeConfigureMessage{
		Address:   configAddr,
		Sender:    publisher.Address(),
		Timestamp: timeStampStr,
		Attr:      attr,
	}
	publisher.publishObject(configAddr, false, &configureMessage, encryptionKey)
}
