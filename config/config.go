package config

import (
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// Load configuration file.
func Load() error {
	configPath := os.Getenv("KURE_CONFIG")

	if configPath != "" {
		viper.AddConfigPath(configPath)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		viper.AddConfigPath(home)
	}

	viper.SetConfigName("config")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return errors.Wrap(err, "failed reading config")
	}

	return nil
}
