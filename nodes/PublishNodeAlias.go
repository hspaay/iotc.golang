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

// PublishNodeAliasCommand publishes the command to set a remote node alias using the existing
// node address. This signs and encrypts the message for the destination
func PublishNodeAliasCommand(
	nodeAddress string, alias, sender string,
	messageSigner *messaging.MessageSigner, encryptionKey *ecdsa.PublicKey) error {

	logrus.Infof("PublishNodeAlias: publishing encrypted alias to %s", nodeAddress)
	segments := strings.Split(nodeAddress, "/")
	if len(segments) < 3 {
		return lib.MakeErrorf("PublishNodeAlias: Node address %s is invalid", nodeAddress)
	}
	aliasAddr := MakeSetAliasAddress(segments[0], segments[1], segments[2])
	// Encecode the SetMessage
	timeStampStr := time.Now().Format("2006-01-02T15:04:05.000-0700")
	var aliasMessage = types.NodeAliasMessage{
		Address:   aliasAddr,
		Sender:    sender,
		Timestamp: timeStampStr,
		Alias:     alias,
	}
	err := messageSigner.PublishObject(aliasAddr, false, &aliasMessage, encryptionKey)
	return err
}
