package config

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"runtime"
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
kure config -c -p path/to/file`

// NewCmd returns a new command.
func NewCmd(db *bolt.DB, r io.Reader) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "config",
		Short:   "Read or create the configuration file",
		Aliases: []string{"cfg"},
		Example: example,
		PreRunE: cmdutil.RequirePassword(db),
		RunE:    runConfig(db, r),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset flags (session)
			create = false
			path = ""
		},
	}

	f := cmd.Flags()
	f.StringVarP(&path, "path", "p", "", "set config file path")
	f.BoolVarP(&create, "create", "c", false, "create a config file")

	cmd.AddCommand(argon2SubCmd(db))

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
	fmt.Println("Leave blank to use default value")
	fmt.Print("\n")

	scanner := bufio.NewScanner(r)

	name := cmdutil.Scan(scanner, "Database name")
	dbPath := cmdutil.Scan(scanner, "Database path")
	format := cmdutil.Scan(scanner, "Entry format")
	port := cmdutil.Scan(scanner, "Http port")
	prefix := cmdutil.Scan(scanner, "Session prefix")
	timeout := cmdutil.Scan(scanner, "Session timeout")
	mem := cmdutil.Scan(scanner, "Argon2 memory")
	iter := cmdutil.Scan(scanner, "Argon2 iterations")
	ths := cmdutil.Scan(scanner, "Argon2 threads")

	if port == "" {
		port = "4000"
	}
	if mem == "" {
		mem = "1048576"
	}
	if iter == "" {
		iter = "1"
	}
	if ths == "" {
		ths = fmt.Sprintf("%d", runtime.NumCPU())
	}

	httpPort, err := strconv.Atoi(port)
	if err != nil {
		return errors.New("invalid port number")
	}

	memory, err := strconv.Atoi(mem)
	if err != nil {
		return errors.New("invalid argon2 memory number")
	}

	iterations, err := strconv.Atoi(iter)
	if err != nil {
		return errors.New("invalid argon2 iteration number")
	}

	threads, err := strconv.Atoi(ths)
	if err != nil {
		return errors.New("invalid argon2 thread number")
	}

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

	viper.Set("database.name", name)
	viper.Set("database.path", dbPath)
	viper.Set("entry.format", f)
	viper.Set("http.port", httpPort)
	viper.Set("session.prefix", prefix)
	viper.Set("session.timeout", timeout)
	viper.Set("argon2.memory", memory)
	viper.Set("argon2.iterations", iterations)
	viper.Set("argon2.threads", threads)

	return nil
}
