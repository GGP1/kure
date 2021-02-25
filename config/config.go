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
		ext := filepath.Ext(envPath)
		if ext == "" || ext == "." {
			return errors.New("\"KURE_CONFIG\" environment variable must have an extension")
		}

		viper.SetConfigType(ext[1:])
		viper.SetConfigFile(envPath)

	default:
		home, err := os.UserHomeDir()
		if err != nil {
			return errors.Wrap(err, "couldn't find the home directory")
		}
		home = filepath.Join(home, ".kure")

		if err := os.MkdirAll(home, 0600); err != nil {
			return errors.Wrap(err, "couldn't create the configuration directory")
		}

		configPath := filepath.Join(home, "kure.yaml")
		viper.SetConfigType("yaml")

		if _, err := os.Stat(configPath); err != nil {
			if os.IsNotExist(err) {
				setDefaults(filepath.Join(home, "kure.db"))
				if err := viper.WriteConfigAs(configPath); err != nil {
					return errors.Wrap(err, "couldn't write the configuration file")
				}
			} else {
				return err
			}
		}

		viper.SetConfigFile(configPath)
	}

	viper.SetConfigPermissions(0600)
	if err := viper.ReadInConfig(); err != nil {
		return errors.Wrap(err, "couldn't read the configuration")
	}

	if viper.InConfig("auth") {
		return errors.New("found invalid key in the configuration file: \"auth\"")
	}

	return nil
}

func setDefaults(dbPath string) {
	var defaults = map[string]interface{}{
		"clipboard.timeout": "0s",
		"database.path":     dbPath,
		"editor":            "vim",
		"keyfile.path":      "",
		"session.prefix":    "kure:~ $",
		"session.timeout":   "0s",
	}

	for k, v := range defaults {
		viper.SetDefault(k, v)
	}
}
