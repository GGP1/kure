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
	"github.com/spf13/viper"
	bolt "go.etcd.io/bbolt"
)

func main() {
	if err := config.Load(); err != nil {
		log.Fatal(err)
	}

	dbPath := filepath.Clean(viper.GetString("database.path"))
	db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 200 * time.Millisecond})
	if err != nil {
		log.Fatalf("couldn't open the database: %v", err)
	}

	// Listen for a signal to release resources and delete sensitive information
	sig.Signal.Listen(db)

	if err := root.Execute(db); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		db.Close()
		memguard.SafeExit(1)
	}

	db.Close()
	memguard.SafeExit(0)
}
