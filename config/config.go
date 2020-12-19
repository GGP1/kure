package config

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// Load configuration file.
func Load() error {
	envPath := os.Getenv("KURE_CONFIG")

	switch {
	case envPath != "":
		ext := filepath.Ext(envPath)
		if ext == "" {
			return errors.New("\"KURE_CONFIG\" env var must have an extension")
		}

		viper.SetConfigFile(envPath)
		viper.SetConfigType(ext[1:])

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

		_, err = os.Stat(home)
		if err != nil {
			if os.IsNotExist(err) {
				if err := viper.SafeWriteConfigAs(home); err != nil {
					return errors.Wrap(err, "failed writing config file")
				}
			} else {
				return err
			}
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
		"database.path":     "",
		"database.name":     "kure",
		"entry.format":      []int{1, 2, 3, 4, 5},
		"entry.repeat":      true,
		"http.port":         4000,
		"argon2.iterations": 1,
		"argon2.memory":     1048576,
		"argon2.threads":    runtime.NumCPU(),
	}

	for k, v := range defaults {
		viper.SetDefault(k, v)
	}
}
