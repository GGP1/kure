package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/GGP1/kure/auth"
	"github.com/GGP1/kure/commands/root"
	"github.com/GGP1/kure/config"
	"github.com/GGP1/kure/sig"

	"github.com/awnumar/memguard"
	bolt "go.etcd.io/bbolt"
)

func main() {
	if err := config.Init(); err != nil {
		fmt.Fprintln(os.Stderr, "couldn't initialize the configuration:", err)
		os.Exit(1)
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
