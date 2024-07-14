package main

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/GGP1/kure/auth"
	"github.com/GGP1/kure/config"
	dbutil "github.com/GGP1/kure/db"
	"github.com/GGP1/kure/db/bucket"
	"github.com/GGP1/kure/terminal"

	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
)

type record struct {
	key   []byte
	value []byte
}

const confMessage = "This script modifies all records names by performing a XOR operation " +
	"against the authentication key. Are you sure you want to proceed?"

func main() {
	if err := config.Init(); err != nil {
		log.Fatalf("couldn't initialize the configuration: %v", err)
	}

	dbPath := filepath.Clean(config.GetString("database.path"))
	db, err := bolt.Open(dbPath, 0o600, &bolt.Options{Timeout: 200 * time.Millisecond})
	if err != nil {
		log.Fatalf("couldn't open the database: %v", err)
	}

	if err := auth.Login(db); err != nil {
		log.Fatalf("couldn't log in: %v", err)
	}

	if err := xorNames(db, os.Stdin); err != nil {
		log.Fatalf("couldn't xor names: %v", err)
	}
}

func xorNames(db *bolt.DB, r io.Reader) error {
	if !terminal.Confirm(r, confMessage) {
		return nil
	}

	tx, err := db.Begin(true)
	if err != nil {
		return errors.Wrap(err, "starting transaction")
	}
	defer tx.Rollback()

	buckets := bucket.GetNames()
	for _, bucket := range buckets {
		b := tx.Bucket(bucket)
		cursor := b.Cursor()
		mp := make(map[string]record, b.Stats().KeyN)

		// The bucket mustn't be modified inside the loop; this will result in undefined behavior
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			mp[string(k)] = record{
				key:   dbutil.XorName(k),
				value: v,
			}
		}

		for oldName, newRecord := range mp {
			if err := b.Put(newRecord.key, newRecord.value); err != nil {
				return errors.Wrap(err, "saving new record")
			}

			if err := b.Delete([]byte(oldName)); err != nil {
				return errors.Wrap(err, "deleting old record")
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "commiting transaction")
	}

	return nil
}
