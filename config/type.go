package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// config is the variable used to manage Kure's configuration.
var config *Config

func init() {
	config = New()
}

// Config contains the elements for handling the configuration.
type Config struct {
	filename  string
	mp        map[string]interface{}
	separator string
}

// New returns a new Config.
func New() *Config {
	return &Config{
		filename:  "",
		mp:        make(map[string]interface{}),
		separator: ".",
	}
}

// Get the value mapped to the specified key.
func (c *Config) Get(key string) interface{} {
	if key == "" {
		return nil
	}

	path := strings.Split(key, c.separator)
	return search(c.mp, path)
}

// Load reads the file and populates the config map.
func (c *Config) Load(filename string) error {
	if filename == "" {
		return errors.New("no configuration file was specified")
	}
	config.filename = filename

	data, err := os.ReadFile(config.filename)
	if err != nil {
		return err
	}

	return config.populateMap(data, filepath.Ext(config.filename))
}

// Set sets a value for the key passed.
func (c *Config) Set(key string, value interface{}) {
	if key == "" {
		return
	}

	path := strings.Split(key, c.separator)
	insert(c.mp, path, value)
}

// Write creates a new file and writes the configuration map content to it.
func (c *Config) Write(filename string, flags int) error {
	// Avoid including the auth parameters in the configuration file
	temp := c.Get("auth")
	defer c.Set("auth", temp)
	delete(config.mp, "auth")

	content, err := c.marshal(filepath.Ext(filename))
	if err != nil {
		return err
	}

	f, err := os.OpenFile(filename, flags, 0600)
	if err != nil {
		return err
	}

	if _, err := f.Write(content); err != nil {
		return errors.Wrap(err, "writing config")
	}

	if err := f.Close(); err != nil {
		return errors.Wrap(err, "closing file")
	}

	return nil
}

func (c *Config) marshal(ext string) ([]byte, error) {
	switch ext {
	case ".json":
		return json.MarshalIndent(c.mp, "", "   ")

	case ".toml":
		return toml.Marshal(c.mp)

	case ".yaml", ".yml":
		return yaml.Marshal(c.mp)

	default:
		return nil, errors.Errorf("unsupported file type: %q", ext)
	}
}

// populateMap parses the data and populates the configuration map.
func (c *Config) populateMap(data []byte, ext string) error {
	switch ext {
	case ".json":
		return json.Unmarshal(data, &c.mp)

	case ".toml":
		return toml.Unmarshal(data, &c.mp)

	case ".yaml", ".yml":
		return yaml.Unmarshal(data, &c.mp)

	default:
		return errors.Errorf("unsupported file type: %q", ext)
	}
}

// insert inserts a value to the map passed.
func insert(mp map[string]interface{}, path []string, v interface{}) {
	if len(path) == 0 {
		return
	}

	key := path[0]

	if len(path) > 1 {
		current := mp[key]

		switch c := current.(type) {
		case map[string]interface{}:
			// Use the current map to continue building the path
			insert(c, path[1:], v)

		default:
			// Replace the current value (or nil) with a map
			next := make(map[string]interface{})
			mp[key] = next
			insert(next, path[1:], v)
		}
		return
	}

	mp[key] = v
}

// search looks for a value in the map specified.
func search(mp map[string]interface{}, path []string) interface{} {
	if len(path) == 0 {
		return nil
	}

	current, ok := mp[path[0]]
	if ok {
		if len(path) == 1 {
			return current
		}

		switch c := current.(type) {
		case map[string]interface{}:
			// Search on a deeper level
			return search(c, path[1:])
		}
	}

	return nil
}
