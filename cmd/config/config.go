package config

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	cmdutil "github.com/GGP1/kure/cmd"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	bolt "go.etcd.io/bbolt"
)

var (
	create bool
	path   string
)

var example = `
* Read config file
kure config

* Create a config file and save it 
kure config -c -p path/to/config`

// NewCmd returns a new command.
func NewCmd(db *bolt.DB, r io.Reader) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "config [-c create] [-p path]",
		Short:   "Read or create the configuration file",
		Aliases: []string{"cfg"},
		Example: example,
		RunE:    runConfig(db, r),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset flags defaults (session)
			create = false
			path = ""
		},
	}

	f := cmd.Flags()
	f.StringVarP(&path, "path", "p", "", "set config file path")
	f.BoolVarP(&create, "create", "c", false, "create a config file")

	return cmd
}

func runConfig(db *bolt.DB, r io.Reader) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		if path == "" {
			cfgPath, err := cmdutil.GetConfigPath()
			if err != nil {
				return err
			}
			path = cfgPath
		}

		if create {
			if err := setConfig(r); err != nil {
				return err
			}

			if filepath.Ext(path) == "" {
				path += ".yaml"
			}

			if err := viper.WriteConfigAs(path); err != nil {
				return errors.Errorf("failed creating config file: %v", err)
			}

			return nil
		}

		if err := cmdutil.RequirePassword(db); err != nil {
			return err
		}

		data, err := ioutil.ReadFile(path)
		if err != nil {
			return errors.Errorf("failed reading config file: %v", err)
		}

		file := strings.TrimSpace(string(data))
		fmt.Printf(`
File location: %s
		
%s
`, path, file)

		return nil
	}
}

func setConfig(r io.Reader) error {
	scanner := bufio.NewScanner(r)

	name := cmdutil.Scan(scanner, "Database name")
	dbPath := cmdutil.Scan(scanner, "Database path")
	format := cmdutil.Scan(scanner, "Entry format")
	repeat := cmdutil.Scan(scanner, "Repeat characters")
	port := cmdutil.Scan(scanner, "Http port")
	prefix := cmdutil.Scan(scanner, "Session prefix")
	timeout := cmdutil.Scan(scanner, "Session timeout")
	mem := cmdutil.Scan(scanner, "Argon2id memory")
	iter := cmdutil.Scan(scanner, "Argon2id iterations")

	httpPort, err := strconv.Atoi(port)
	if err != nil {
		return errors.New("invalid port number")
	}

	memory, err := strconv.Atoi(mem)
	if err != nil {
		return errors.New("invalid argon2id memory number")
	}

	iterations, err := strconv.Atoi(iter)
	if err != nil {
		return errors.New("invalid argon2id iteration number")
	}

	viper.Set("database.name", name)
	viper.Set("database.path", dbPath)
	viper.Set("http.port", httpPort)
	viper.Set("entry.repeat", repeat)
	viper.Set("session.prefix", prefix)
	viper.Set("session.timeout", timeout)
	viper.Set("argon2id.memory", memory)
	viper.Set("argon2id.iterations", iterations)

	levels := strings.Split(format, ",")

	var f []int
	for _, v := range levels {
		l, err := strconv.Atoi(v)
		if err != nil {
			return errors.New("invalid level")
		}

		if l > 0 && l < 6 {
			f = append(f, l)
		}
	}

	viper.Set("entry.format", f)

	return nil
}
