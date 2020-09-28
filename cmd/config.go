package cmd

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

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
		Format string
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
			p := os.Getenv("KURE_CONFIG")

			if path != "" {
				p = path
			}

			filename := fmt.Sprintf("%s/config.yaml", p)

			data, err := ioutil.ReadFile(filename)
			if err != nil {
				must(err)
			}

			fmt.Println(string(data))
			return
		}

		config := configInput()

		viper.Set("database.name", config.Database.Name)
		viper.Set("database.path", config.Database.Path)
		viper.Set("user.password", config.User.Password)
		viper.Set("http.port", config.HTTP.Port)

		var format []int

		levels := strings.Split(config.Entry.Format, ",")

		for _, v := range levels {
			integer, _ := strconv.Atoi(v)
			format = append(format, integer)
		}

		viper.Set("entry.format", format)

		if path != "" {
			if err := viper.WriteConfigAs(path); err != nil {
				must(err)
			}
		} else {
			home, err := os.UserHomeDir()
			if err != nil {
				must(err)
			}

			path = fmt.Sprintf("%s/config.yaml", home)

			if err := viper.WriteConfigAs(path); err != nil {
				must(err)
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(configCmd)
	configCmd.Flags().StringVarP(&path, "path", "p", "", "set config file path")
	configCmd.Flags().BoolVarP(&create, "create", "c", false, "create a config file")
}

func configInput() *config {
	var name, DBPath, password, format, port string

	scanner := bufio.NewScanner(os.Stdin)

	name = scan(scanner, "Database name", name)
	DBPath = scan(scanner, "Database path", DBPath)
	password = scan(scanner, "User password", password)
	format = scan(scanner, "Entry format", format)
	port = scan(scanner, "Http port", port)

	httpPort, err := strconv.Atoi(port)
	if err != nil {
		must(err)
	}

	config := &config{
		Database: struct {
			Name string
			Path string
		}{
			Name: name,
			Path: path,
		},
		User: struct{ Password string }{
			Password: password,
		},
		Entry: struct{ Format string }{
			Format: format,
		},
		HTTP: struct{ Port int }{
			Port: httpPort,
		},
	}

	return config
}
