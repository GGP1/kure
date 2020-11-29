package config

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// Load configuration file.
func Load() error {
	envPath := os.Getenv("KURE_CONFIG")

	switch {
	case envPath != "":
		if filepath.Ext(envPath) == "" {
			envPath += ".yaml"
		}

		viper.SetConfigFile(envPath)

	default:
		home, err := os.UserHomeDir()
		if err != nil {
			return errors.Wrap(err, "couldn't find user home directory")
		}

		setDefaults()
		viper.SetConfigName(".kure")
		viper.SetConfigType("yaml")
		viper.Set("database.path", home)
		viper.Set("database.name", "kure.db")

		home = filepath.Join(home, ".kure.yaml")

		if err := viper.WriteConfigAs(home); err != nil {
			return errors.Wrap(err, "failed writing config")
		}

		viper.SetConfigFile(home)
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return errors.Wrap(err, "failed reading config")
	}

	return nil
}

func setDefaults() {
	var defaults = map[string]interface{}{
		"database.name": "kure",
		"user.password": "",
		"entry.format":  []int{1, 2, 3, 4, 5},
		"entry.repeat":  false,
		"http.port":     4000,
	}

	for k, v := range defaults {
		viper.SetDefault(k, v)
	}
}
