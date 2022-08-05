package config

import (
	"os"
	"path/filepath"
	"time"

	"github.com/awnumar/memguard"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
)

// Init intializes the configuration.
func Init() error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	if err := Load(configPath); err != nil {
		return err
	}

	if IsSet("auth") {
		return errors.New("found invalid key: \"auth\"")
	}

	return nil
}

func getConfigPath() (string, error) {
	configPath := os.Getenv("KURE_CONFIG")
	if configPath != "" {
		ext := filepath.Ext(configPath)
		if ext == "" || ext == "." {
			return "", errors.New("\"KURE_CONFIG\" environment variable path must have an extension")
		}

		return configPath, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", errors.Wrap(err, "finding home directory")
	}

	homeDir = filepath.Join(homeDir, ".kure")
	configPath = filepath.Join(homeDir, "kure.yaml")

	if _, err := os.Stat(configPath); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
		if err := createDefaultConfigFile(homeDir, configPath); err != nil {
			return "", err
		}
	}

	return configPath, nil
}

func createDefaultConfigFile(homeDir, configPath string) error {
	if err := os.MkdirAll(homeDir, 0700); err != nil {
		return errors.Wrap(err, "creating the directory")
	}
	SetDefaults(filepath.Join(homeDir, "kure.db"))
	return Write(configPath, true)
}

// Filename returns the name of the file that the configuration is using.
func Filename() string {
	return config.filename
}

// Get returns an uncasted value from the config map.
func Get(key string) interface{} {
	return config.Get(key)
}

// GetEnclave returns an enclave from the config map.
func GetEnclave(key string) *memguard.Enclave {
	v := config.Get(key)
	if v == nil {
		return nil
	}

	return v.(*memguard.Enclave)
}

// GetDuration returns a duration from the config map.
func GetDuration(key string) time.Duration {
	return cast.ToDuration(config.Get(key))
}

// GetString returns a string from the config map.
func GetString(key string) string {
	return cast.ToString(config.Get(key))
}

// GetStringMapString returns a map[string]string from the config map.
func GetStringMapString(key string) map[string]string {
	return cast.ToStringMapString(config.Get(key))
}

// GetUint32 returns an uint32 from the config map.
func GetUint32(key string) uint32 {
	return cast.ToUint32(config.Get(key))
}

// IsSet returns if the key exists in the config map or not.
func IsSet(key string) bool {
	return config.Get(key) != nil
}

// Load reads the configuration file and populates the map.
func Load(configPath string) error {
	return config.Load(configPath)
}

// Reset sets config to its initial value.
func Reset() {
	config = New()
}

// Set sets a value to the config map.
func Set(key string, value interface{}) {
	config.Set(key, value)
}

// SetDefaults populates the config map with the default values.
func SetDefaults(dbPath string) {
	var defaults = map[string]interface{}{
		"clipboard.timeout": "0s",
		"database.path":     dbPath,
		"editor":            "vim",
		"keyfile.path":      "",
		"session.prefix":    "kure:~ $",
		"session.scripts":   map[string]string{},
		"session.timeout":   "0s",
	}

	for k, v := range defaults {
		Set(k, v)
	}
}

// SetFilename sets the configuration filename.
func SetFilename(filename string) {
	config.filename = filename
}

// Write encodes and writes the config map to the specified file.
//
// If exclusive is true an error will be returned if the file
// already exists, otherwise it will truncate the file and write to it.
func Write(filename string, exclusive bool) error {
	flags := os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	if exclusive {
		flags = os.O_CREATE | os.O_WRONLY | os.O_EXCL
	}

	return config.Write(filename, flags)
}

// WriteStruct writes the configuration empty structure to the given file.
func WriteStruct(filename string) error {
	temp := config.mp
	config.mp = map[string]interface{}{
		"clipboard": map[string]interface{}{
			"timeout": "",
		},
		"database": map[string]interface{}{
			"path": "",
		},
		"editor": "",
		"keyfile": map[string]interface{}{
			"path": "",
		},
		"session": map[string]interface{}{
			"prefix":  "",
			"scripts": map[string]string{},
			"timeout": "",
		},
	}

	if err := Write(filename, true); err != nil {
		return err
	}

	// Return the map to its previous state
	config.mp = temp

	return nil
}
