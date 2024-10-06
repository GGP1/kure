package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/GGP1/kure/auth"
	"github.com/GGP1/kure/commands/root"
	"github.com/GGP1/kure/config"
	"github.com/GGP1/kure/sig"
	"github.com/spf13/pflag"

	"github.com/awnumar/memguard"
	bolt "go.etcd.io/bbolt"
)

func main() {
	if err := validateFlags(); err != nil {
		if errors.Is(err, pflag.ErrHelp) {
			os.Exit(0)
		}
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	if err := config.Init(); err != nil {
		fmt.Fprintln(os.Stderr, "couldn't initialize the configuration:", err)
		os.Exit(1)
	}

	// Check for and run stateless commands
	if len(os.Args) < 2 || root.IsStatelessCommand(os.Args[1]) {
		if err := root.NewCmd(nil).Execute(); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	dbPath := filepath.Clean(config.GetString("database.path"))
	db, err := bolt.Open(dbPath, 0o600, &bolt.Options{Timeout: 200 * time.Millisecond})
	if err != nil {
		fmt.Fprintln(os.Stderr, "couldn't open the database:", err)
		os.Exit(1)
	}

	if err := auth.Login(db); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		db.Close()
		memguard.SafeExit(1)
	}

	// Listen for a signal to release resources and delete sensitive information
	sig.Signal.Listen(db)

	if err := root.NewCmd(db).Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		db.Close()
		memguard.SafeExit(1)
	}

	db.Close()
	memguard.SafeExit(0)
}

// validateFlags looks for the command called and parses its flags. If the flag is `--help`,
// it will print the command's help message and return the error pflag.ErrHelp.
func validateFlags() error {
	cmd, args, err := root.NewCmd(nil).Find(os.Args[1:])
	if err != nil {
		return err
	}

	if err := cmd.ParseFlags(args); err != nil {
		if errors.Is(err, pflag.ErrHelp) {
			if err := cmd.Help(); err != nil {
				return err
			}
			return pflag.ErrHelp
		}
		return err
	}

	return nil
}
