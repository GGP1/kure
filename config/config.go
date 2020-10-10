package config

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var defaults = map[string]interface{}{
	"database.name": "kure",
	"database.path": os.UserHomeDir,
	"user.password": "",
	"entry.format":  []int{1, 2, 3, 4, 5},
	"entry.repeat":  false,
	"http.port":     4000,
}

// Load configuration file.
func Load() error {
	configPath := os.Getenv("KURE_CONFIG")

	if configPath != "" {
		viper.AddConfigPath(configPath)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return errors.Wrap(err, "couldn't find user home directory")
		}

		setDefaults()
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")

		path := fmt.Sprintf("%s/config.yaml", home)

		if err := viper.WriteConfigAs(path); err != nil {
			return errors.Wrap(err, "failed writing config")
		}

		viper.AddConfigPath(path)
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return errors.Wrap(err, "failed reading config")
	}

	return nil
}

func setDefaults() {
	for k, v := range defaults {
		viper.SetDefault(k, v)
	}
}
