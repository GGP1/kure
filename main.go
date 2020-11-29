package main

import (
	"path/filepath"
	"time"

	// jpeg and png imported for displaying images on the terminal
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"strings"

	"github.com/GGP1/kure/cmd/add"
	"github.com/GGP1/kure/cmd/backup"
	"github.com/GGP1/kure/cmd/card"
	"github.com/GGP1/kure/cmd/clear"
	configcmd "github.com/GGP1/kure/cmd/config"
	"github.com/GGP1/kure/cmd/copy"
	"github.com/GGP1/kure/cmd/edit"
	"github.com/GGP1/kure/cmd/file"
	"github.com/GGP1/kure/cmd/gen"
	"github.com/GGP1/kure/cmd/ls"
	"github.com/GGP1/kure/cmd/rm"
	"github.com/GGP1/kure/cmd/root"
	"github.com/GGP1/kure/cmd/session"
	"github.com/GGP1/kure/cmd/stats"
	"github.com/GGP1/kure/cmd/wallet"
	"github.com/GGP1/kure/config"

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
		buckets := [4]string{"kure_card", "kure_entry", "kure_file", "kure_wallet"}
		for _, bucket := range buckets {
			_, err := tx.CreateBucketIfNotExists([]byte(bucket))
			if err != nil {
				return errors.Wrapf(err, "couldn't create %q bucket", bucket)
			}
		}

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	root.Register(add.NewCmd(db, os.Stdin))
	root.Register(backup.NewCmd(db))
	root.Register(card.NewCmd(db))
	root.Register(clear.NewCmd())
	root.Register(configcmd.NewCmd(db, os.Stdin))
	root.Register(copy.NewCmd(db))
	root.Register(edit.NewCmd(db, os.Stdin))
	root.Register(file.NewCmd(db))
	root.Register(gen.NewCmd())
	root.Register(ls.NewCmd(db))
	root.Register(rm.NewCmd(db, os.Stdin))
	root.Register(session.NewCmd(db, os.Stdin))
	root.Register(stats.NewCmd(db))
	root.Register(wallet.NewCmd(db))

	root.Execute(db)
}
