package cmd

import (
	"bufio"
	"crypto/sha512"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type config struct {
	Database struct {
		Name string
		Path string
	}
	User struct {
		Password string
	}
	Entry struct {
		Format []int
		Repeat bool
	}
	HTTP struct {
		Port int
	}
}

var create bool

var configCmd = &cobra.Command{
	Use:   "config [-c create] [-p path]",
	Short: "Read or create the configuration file",
	Run: func(cmd *cobra.Command, args []string) {
		if !create {
			if path == "" {
				path = getConfigPath()
			}

			if !strings.Contains(path, ".") {
				path = filepath.Join(path, "config.yaml")
			}

			data, err := ioutil.ReadFile(path)
			if err != nil {
				fatalf("failed reading config file: %v", err)
			}

			file := strings.TrimSpace(string(data))
			fmt.Println("\n" + file)
			return
		}

		if err := setConfig(); err != nil {
			fatal(err)
		}

		if path != "" {
			if err := viper.WriteConfigAs(path); err != nil {
				fatalf("%s: %v", errCreatingConfig, err)
			}
		} else {
			home, err := os.UserHomeDir()
			if err != nil {
				fatalf("couldn't find user home directory: %v", err)
			}

			path = fmt.Sprintf("%s/config.yaml", home)

			if err := viper.WriteConfigAs(path); err != nil {
				fatalf("%s: %v", errCreatingConfig, err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(configCmd)

	configCmd.Flags().StringVarP(&path, "path", "p", "", "set config file path")
	configCmd.Flags().BoolVarP(&create, "create", "c", false, "create a config file")
}

func getConfigPath() string {
	var filename string
	configPath := os.Getenv("KURE_CONFIG")

	if configPath != "" {
		base := filepath.Base(configPath)
		if strings.Contains(base, ".") {
			return configPath
		}

		filename = fmt.Sprintf("%s/config.yaml", configPath)
		return filename
	}

	home, err := os.UserHomeDir()
	if err != nil {
		fatalf("couldn't find user home directory: %v", err)
	}

	filename = fmt.Sprintf("%s/config.yaml", home)

	return filename
}

func setConfig() error {
	var name, dbPath, password, format, repeat, port string

	scanner := bufio.NewScanner(os.Stdin)

	scan(scanner, "Database name", &name)
	scan(scanner, "Database path", &dbPath)
	scan(scanner, "User password", &password)
	scan(scanner, "Entry format", &format)
	scan(scanner, "Repeat characters", &repeat)
	scan(scanner, "Http port", &port)

	httpPort, err := strconv.Atoi(port)
	if err != nil {
		return errors.New("converting port to an integer")
	}

	h := sha512.New()
	_, err = h.Write([]byte(password))
	if err != nil {
		return errors.Wrap(err, "creating the password hash")
	}
	p := string(h.Sum(nil))

	viper.Set("database.name", name)
	viper.Set("database.path", dbPath)
	viper.Set("user.password", p)
	viper.Set("http.port", httpPort)
	viper.Set("entry.repeat", repeat)

	levels := strings.Split(format, ",")

	var f []int
	for _, v := range levels {
		integer, _ := strconv.Atoi(v)
		f = append(f, integer)
	}

	viper.Set("entry.format", f)

	return nil
}
