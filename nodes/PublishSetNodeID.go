// Package nodes with command to set a remote node's alias
package nodes

import (
	"crypto/ecdsa"
	"strings"
	"time"

	"github.com/iotdomain/iotdomain-go/lib"
	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/sirupsen/logrus"
)

// PublishSetNodeID publishes the command to change a remote node ID using the existing
// node address. This signs and encrypts the message for the destination
func PublishSetNodeID(
	nodeAddress string, newNodeID string, sender string,
	messageSigner *messaging.MessageSigner, encryptionKey *ecdsa.PublicKey) error {

	logrus.Infof("PublishSetNodeID: publishing encrypted message to %s", nodeAddress)
	segments := strings.Split(nodeAddress, "/")
	if len(segments) < 3 {
		return lib.MakeErrorf("PublishNodeAlias: Node address %s is invalid", nodeAddress)
	}
	setNodeIDAddr := MakeSetNodeIDAddress(segments[0], segments[1], segments[2])
	// Encecode the SetMessage
	timeStampStr := time.Now().Format("2006-01-02T15:04:05.000-0700")
	var message = types.SetNodeIDMessage{
		Address:   setNodeIDAddr,
		Sender:    sender,
		Timestamp: timeStampStr,
		NodeID:    newNodeID,
	}
	err := messageSigner.PublishObject(setNodeIDAddr, false, &message, encryptionKey)
	return err
}
