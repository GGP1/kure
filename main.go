package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/GGP1/kure/cmd/add"
	"github.com/GGP1/kure/cmd/backup"
	"github.com/GGP1/kure/cmd/card"
	"github.com/GGP1/kure/cmd/clear"
	configcmd "github.com/GGP1/kure/cmd/config"
	"github.com/GGP1/kure/cmd/copy"
	"github.com/GGP1/kure/cmd/export"
	"github.com/GGP1/kure/cmd/file"
	"github.com/GGP1/kure/cmd/gen"
	importt "github.com/GGP1/kure/cmd/import"
	"github.com/GGP1/kure/cmd/ls"
	"github.com/GGP1/kure/cmd/note"
	"github.com/GGP1/kure/cmd/restore"
	"github.com/GGP1/kure/cmd/rm"
	"github.com/GGP1/kure/cmd/root"
	"github.com/GGP1/kure/cmd/session"
	"github.com/GGP1/kure/cmd/stats"
	"github.com/GGP1/kure/config"

	"github.com/awnumar/memguard"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	bolt "go.etcd.io/bbolt"
)

func main() {
	// Load config file and env variables
	if err := config.Load(); err != nil {
		log.Fatal(err)
	}

	dbPath := strings.TrimSuffix(viper.GetString("database.path"), "/")
	dbName := viper.GetString("database.name")

	path := filepath.Join(dbPath, dbName)
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("couldn't find user home directory: %v", err)
		}
		path = filepath.Join(home, "kure.db")
	}

	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 2 * time.Second})
	if err != nil {
		log.Fatalf("couldn't connect to the database: %v", err)
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		buckets := [4]string{"kure_card", "kure_entry", "kure_file", "kure_note"}
		for _, bucket := range buckets {
			if _, err := tx.CreateBucketIfNotExists([]byte(bucket)); err != nil {
				return errors.Wrapf(err, "couldn't create %q bucket", bucket)
			}
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGHUP, syscall.SIGTERM)
	go signals(db, interrupt)

	root.Register(add.NewCmd(db, os.Stdin))
	root.Register(backup.NewCmd(db))
	root.Register(card.NewCmd(db, os.Stdin))
	root.Register(clear.NewCmd())
	root.Register(configcmd.NewCmd(db, os.Stdin))
	root.Register(copy.NewCmd(db))
	root.Register(export.NewCmd(db))
	root.Register(file.NewCmd(db, os.Stdin, os.Stdout))
	root.Register(gen.NewCmd())
	root.Register(importt.NewCmd(db))
	root.Register(ls.NewCmd(db))
	root.Register(note.NewCmd(db, os.Stdin))
	root.Register(restore.NewCmd(db))
	root.Register(rm.NewCmd(db, os.Stdin))
	root.Register(session.NewCmd(db, os.Stdin, interrupt))
	root.Register(stats.NewCmd(db))

	root.Execute(db)
}

// signals waits for a signal to release resources, delete any sensitive information and exit safely.
//
// db.Close() will block waiting for open transactions to finish before closing.
func signals(db *bolt.DB, interrupt chan os.Signal) {
	<-interrupt
	db.Close()
	fmt.Fprint(os.Stderr, "\nExiting...\n")
	memguard.SafeExit(0)
}
