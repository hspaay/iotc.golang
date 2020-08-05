// Package persist - Load and save publisher application configuration
package lib_test

import (
	"testing"

	"github.com/iotdomain/iotdomain-go/lib"
	"github.com/stretchr/testify/assert"
)

const PublisherID = "publisher1"
const cacheFolder = "./test"
const configFolder = "./test"

type TestConfig struct {
	ConfigString string `yaml:"cstring"`
	ConfigNumber int    `yaml:"cnumber"`
}
type MessengerConfig struct {
	Server string `yaml:"server"`         // Messenger hostname or ip address
	Port   uint16 `yaml:"port,omitempty"` // optional port, default is 8883 for TLS
}

// TestAppConfig configuration
func TestAppConfig(t *testing.T) {

	appConfig := TestConfig{}
	err := lib.LoadAppConfig(configFolder, PublisherID, &appConfig)
	assert.NoError(t, err, "Failed loading app config")
	assert.Equal(t, appConfig.ConfigNumber, 5, "AppConfig does not contain number 5")

	// default config folder, file not found error
	err = lib.LoadAppConfig("", PublisherID, &appConfig)
	assert.Error(t, err)
	// unmarshal error
	inputConfig := "sss"
	err = lib.LoadAppConfig(configFolder, PublisherID, &inputConfig)
	assert.Error(t, err)

}

// TestMessengerConfig load configuration
func TestMessengerConfig(t *testing.T) {
	messengerConfig := MessengerConfig{}
	err := lib.LoadMessengerConfig(configFolder, &messengerConfig)
	assert.NoError(t, err, "Failed loading app config")
	assert.Equal(t, "localhost", messengerConfig.Server, "Messenger does not contain server address")
}
