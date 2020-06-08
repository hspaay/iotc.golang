package persist

// AppConfigSuffix to append to the publisher ID to load the application configuration
const KeyfileSuffix = ".keys"

// DefaultConfigFolder for IoTConnect publisher configuration files: ~/.config/iotc
// var DefaultConfigFolder = path.Join(UserHomeDir, ".config", "iotc")

// LoadAppConfig loads the application configuration from a configuration file
//
// configFolder contains the location for the configuration files.
// appID to load from file <appID>.yaml
// appConfig is the object to store the application configuration using yaml. This is optional.
// func LoadAppConfig(configFolder string, appID string, appConfig interface{}) error {
// 	configFile := appID + AppConfigSuffix
// 	err := LoadYamlConfig(configFolder, configFile, appID, appConfig)
// 	return err
// }
