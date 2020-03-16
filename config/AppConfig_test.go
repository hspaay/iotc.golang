// Package config - Load and save publisher application configuration
package config

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const PublisherID = "testpub"
const ConfigFolder = "./"

type TestConfig struct {
	ConfigString string `yaml:"cstring"`
	ConfigNumber int    `yaml:"cnumber"`
}

// TestLoad configuration
func TestLoad(t *testing.T) {
	logger := logrus.New()
	logger.SetReportCaller(true) // publisher logging includes caller and file:line#

	testConfig := TestConfig{}
	err := LoadAppConfig(ConfigFolder, PublisherID, nil, &testConfig)
	assert.NoError(t, err, "Failed loading app config")

	assert.Equal(t, testConfig.ConfigNumber, 5, "AppConfig does not contain number 5")
}

// TestSave configuration
func TestSave(t *testing.T) {
	logger := logrus.New()
	logger.SetReportCaller(true) // publisher logging includes caller and file:line#

	// configfile contains ConfigNumber 5
	testConfig := TestConfig{}
	err := LoadAppConfig(ConfigFolder, PublisherID, nil, &testConfig)
	assert.NoError(t, err, "Failed loading app config")
	assert.Equal(t, 5, testConfig.ConfigNumber, "AppConfig does not contain number 5")

	testConfig.ConfigNumber = 6
	err = SaveAppConfig("/tmp", PublisherID, &testConfig)
	assert.NoError(t, err, "Failed saving config")

	testConfig.ConfigNumber = 7
	err = LoadAppConfig("/tmp", PublisherID, nil, &testConfig)
	assert.NoError(t, err, "Failed loading app config")
	assert.Equal(t, 6, testConfig.ConfigNumber, "Previously written AppConfig not loaded")

}
