package cmd

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
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
				path = os.Getenv("KURE_CONFIG")
			} else {
				fatal(errors.New("a path to the configuration file is required"))
			}

			filename := fmt.Sprintf("%s/config.yaml", path)

			data, err := ioutil.ReadFile(filename)
			if err != nil {
				fatalf("failed reading config file: %v", err)
			}

			file := strings.TrimSpace(string(data))

			fmt.Println(file)
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
				fatal(err)
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

	viper.Set("database.name", name)
	viper.Set("database.path", dbPath)
	viper.Set("user.password", password)
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
