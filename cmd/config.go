package cmd

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/GGP1/kure/crypt"
	"github.com/GGP1/kure/db"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configCmd = &cobra.Command{
	Use:   "config [-c create] [-p path]",
	Short: "Read or create the configuration file",
	Run: func(cmd *cobra.Command, args []string) {
		if path == "" {
			path = getConfigPath()
		}

		if !create {
			if err := db.RequirePassword(); err != nil {
				fatal(err)
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

		if !strings.Contains(filepath.Base(path), ".") {
			path = filepath.Join(path, "config.yaml")
		}

		if err := viper.WriteConfigAs(path); err != nil {
			fatalf("failed creating config file: %v", err)
		}

	},
}

func init() {
	rootCmd.AddCommand(configCmd)

	configCmd.Flags().StringVarP(&path, "path", "p", "", "set config file path")
	configCmd.Flags().BoolVarP(&create, "create", "c", false, "create a config file")
}

func getConfigPath() string {
	var path string
	cfgPath := os.Getenv("KURE_CONFIG")

	if cfgPath != "" {
		base := filepath.Base(cfgPath)
		if strings.Contains(base, ".") {
			return cfgPath
		}

		path = filepath.Join(cfgPath, "config.yaml")
		return path
	}

	home, err := os.UserHomeDir()
	if err != nil {
		fatalf("couldn't find user home directory: %v", err)
	}

	path = filepath.Join(filepath.Clean(home), "config.yaml")

	return path
}

func setConfig() error {
	var name, dbPath, format, repeat, port string

	scanner := bufio.NewScanner(os.Stdin)

	scan(scanner, "Database name", &name)
	scan(scanner, "Database path", &dbPath)
	scan(scanner, "Entry format", &format)
	scan(scanner, "Repeat characters", &repeat)
	scan(scanner, "Http port", &port)

	password, err := crypt.AskPassword(true)
	if err != nil {
		return err
	}

	httpPort, err := strconv.Atoi(port)
	if err != nil {
		return errors.New("invalid port")
	}

	viper.Set("database.name", name)
	viper.Set("database.path", dbPath)
	viper.Set("user.password", password)
	viper.Set("http.port", httpPort)
	viper.Set("entry.repeat", repeat)

	levels := strings.Split(format, ",")

	var f []int
	for _, v := range levels {
		integer, err := strconv.Atoi(v)
		if err != nil {
			return errors.New("invalid levels")
		}
		f = append(f, integer)
	}

	viper.Set("entry.format", f)

	return nil
}
