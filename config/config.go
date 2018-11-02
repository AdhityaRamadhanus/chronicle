package config

import (
	"fmt"

	"github.com/spf13/viper"
)

func Init(env string, configPaths []string) error {
	viper.SetConfigName("config")

	// default config path
	viper.AddConfigPath(fmt.Sprintf("config/%s/", env))
	for _, configPath := range configPaths {
		viper.AddConfigPath(configPath)
	}

	if env == "production" {
		viper.SetEnvPrefix("production")
		viper.AutomaticEnv()
	}

	return viper.ReadInConfig()
}
