package config

import (
	"github.com/robertantonyjaikumar/hangover-common/logger"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"os"
)

const (
	varLogLevel     = "log.level"
	varPathToConfig = "../config.yaml"
)

var (
	CFG = GetConfig()
)

type Configuration struct {
	V *viper.Viper
}

// GetConfig return a Configuration struct with allows to
// get viper configurations from yaml and env variables
func GetConfig() *Configuration {
	c := Configuration{
		V: viper.GetViper(),
	}
	c.V.SetDefault(varLogLevel, "info")
	c.V.AutomaticEnv()
	c.V.SetConfigType("yaml")
	c.V.SetConfigName("config")
	c.V.AddConfigPath("./")
	err := c.V.ReadInConfig() // Find and read the config file

	logger.Info("loading config")
	if _, ok := err.(*os.PathError); ok {
		logger.Error(
			"no config file not found. Using default values",
			zap.String("config_path", c.GetPathToConfig()),
		)
	} else if err != nil { // Handle other errors that occurred while reading the config file
		logger.Error("fatal error while reading the config file", zap.Error(err))
	}
	return &c

}

// GetServiceName returns service name
func (c *Configuration) GetServiceName() string {
	return c.V.GetString("service.name")

}

func (c *Configuration) GetPathToConfig() string {
	return c.V.GetString(varPathToConfig)
}
