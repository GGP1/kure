package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/GGP1/kure/commands/root"
	"github.com/GGP1/kure/config"
	"github.com/GGP1/kure/sig"

	"github.com/awnumar/memguard"
	bolt "go.etcd.io/bbolt"
)

var (
	version = "development"
	commit  = ""
)

func main() {
	if err := config.Init(); err != nil {
		log.Fatalf("couldn't initialize the configuration: %v", err)
	}

	dbPath := filepath.Clean(config.GetString("database.path"))
	db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 200 * time.Millisecond})
	if err != nil {
		log.Fatalf("couldn't open the database: %v", err)
	}

	// Listen for a signal to release resources and delete sensitive information
	sig.Signal.Listen(db)

	if err := root.Execute(version, commit, db); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		db.Close()
		memguard.SafeExit(1)
	}

	db.Close()
	memguard.SafeExit(0)
}
