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
	"algorithm":     "",
	"entry.format":  []int{1, 2, 3, 4},
	"http.port":     4000,
}

// Load configuration file.
func Load() error {
	configPath := os.Getenv("KURE_CONFIG")

	if configPath != "" {
		viper.AddConfigPath(configPath)
	} else {
		setDefaults()

		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		path := fmt.Sprintf("%s/config.yaml", home)

		viper.SetConfigName("config")
		viper.SetConfigType("yaml")

		if err := viper.WriteConfigAs(path); err != nil {
			return err
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
