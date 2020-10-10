package db

import (
	"fmt"
	"strings"
	"time"

	"github.com/GGP1/kure/crypt"
	"github.com/GGP1/kure/model/entry"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
)

// cleanExpired removes the expired entries from the database.
func cleanExpired(b *bolt.Bucket, name, expires string, expired chan bool, errCh chan error) {
	if expires == "Never" {
		expired <- false
		return
	}

	// Format expires time and time.Now to compare them
	expiration, err := time.Parse(time.RFC3339, expires)
	if err != nil {
		errCh <- errors.Wrap(err, "expiration time parse")
	}

	now, err := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	if err != nil {
		errCh <- errors.Wrap(err, "time now parse")
	}

	// Clean expired entries
	if now.Sub(expiration) > 0 {
		if err := b.Delete([]byte(name)); err != nil {
			errCh <- errors.Wrap(err, "delete entry")
		}
		expired <- true
	}

	expired <- false
}

// CreateEntry creates a new record.
func CreateEntry(entry *entry.Entry) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(entryBucket)

		exists := b.Get([]byte(entry.Name))
		if exists != nil {
			return errors.New("there is an existing entry with this name, to edit it please use <kure edit>")
		}

		buf, err := proto.Marshal(entry)
		if err != nil {
			return errors.Wrap(err, "marshal entry")
		}

		encEntry, err := crypt.Encrypt(buf)
		if err != nil {
			return errors.Wrap(err, "encrypt entry")
		}

		if err := b.Put([]byte(entry.Name), encEntry); err != nil {
			return errors.Wrap(err, "save entry")
		}

		return nil
	})
}

// DeleteEntry removes an entry from the database.
func DeleteEntry(name string) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(entryBucket)
		t := strings.ToLower(name)

		if err := b.Delete([]byte(t)); err != nil {
			return errors.Wrap(err, "delete entry")
		}

		return nil
	})
}

// EditEntry edits an entry.
func EditEntry(entry *entry.Entry) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(entryBucket)

		buf, err := proto.Marshal(entry)
		if err != nil {
			return errors.Wrap(err, "marshal entry")
		}

		if err := b.Put([]byte(entry.Name), buf); err != nil {
			return errors.Wrap(err, "edit entry")
		}

		return nil
	})
}

// GetEntry retrieves the entry with the specified name.
func GetEntry(name string) (*entry.Entry, error) {
	e := &entry.Entry{}

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(entryBucket)
		t := strings.ToLower(name)

		result := b.Get([]byte(t))
		if result == nil {
			return errors.Errorf("\"%s\" does not exist", name)
		}

		decEntry, err := crypt.Decrypt(result)
		if err != nil {
			return errors.Wrap(err, "decrypt entry")
		}

		if err := proto.Unmarshal(decEntry, e); err != nil {
			return errors.Wrap(err, "unmarshal entry")
		}

		if e.Name == "" {
			return fmt.Errorf("%s does not exist", name)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return e, nil
}

// ListEntries returns a list with all the entries.
func ListEntries() ([]*entry.Entry, error) {
	var entries []*entry.Entry

	expired := make(chan bool, 1)
	errCh := make(chan error, 1)

	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(entryBucket)
		c := b.Cursor()

		// Place cursor in the first line of the bucket and move it to the next one
		for k, v := c.First(); k != nil; k, v = c.Next() {
			entry := &entry.Entry{}

			decEntry, err := crypt.Decrypt(v)
			if err != nil {
				return errors.Wrap(err, "decrypt entry")
			}

			if err := proto.Unmarshal(decEntry, entry); err != nil {
				return errors.Wrap(err, "unmarshal entry")
			}

			go cleanExpired(b, entry.Name, entry.Expires, expired, errCh)

			select {
			case e := <-expired:
				if !e {
					entries = append(entries, entry)
				}
			case err := <-errCh:
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return entries, nil
}
