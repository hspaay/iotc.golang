// Package nodes with command to configure a discovered domain node
package nodes

import (
	"crypto/ecdsa"
	"strings"
	"time"

	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/sirupsen/logrus"
)

// PublishConfigureNode updates the configuration of a remote node by this publisher
// If a key is given then the signed message will be encrypted with the given encryption key
func PublishConfigureNode(
	destinationAddress string, attr types.NodeAttrMap, sender string,
	messageSigner *messaging.MessageSigner, encryptionKey *ecdsa.PublicKey) {

	logrus.Infof("PublishSetConfigure: publishing encrypted configuration to %s", destinationAddress)
	// Check that address is one of our inputs
	segments := strings.Split(destinationAddress, "/")
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
		Sender:    sender,
		Timestamp: timeStampStr,
		Attr:      attr,
	}
	messageSigner.PublishObject(configAddr, false, &configureMessage, encryptionKey)
}
