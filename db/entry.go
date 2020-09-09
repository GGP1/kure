package db

import (
	"fmt"
	"strings"
	"time"

	"github.com/GGP1/kure/entry"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
)

// cleanExpired removes the expired entries from the database.
func cleanExpired(b *bolt.Bucket, title, expires []byte, expired chan bool, errCh chan error) {
	if string(expires) == "Never" {
		expired <- false
		return
	}

	// Format expires time and time.Now to compare them
	expiration, err := time.Parse(time.RFC3339, string(expires))
	if err != nil {
		errCh <- errors.Wrap(err, "expiration time parse")
	}

	now, err := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	if err != nil {
		errCh <- errors.Wrap(err, "time now parse")
	}

	// Clean expired entries
	if now.Sub(expiration) > 0 {
		if err := b.Delete(title); err != nil {
			errCh <- errors.Wrap(err, "delete entry")
		}
		expired <- true
	}

	expired <- false
}

// CreateEntry creates a new record.
func CreateEntry(entry *entry.Entry) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(defaultBucket)
		if b == nil {
			return fmt.Errorf("%s folder does not exist", defaultBucket)
		}

		exists := b.Get(entry.Title)
		if exists != nil {
			return errors.New("there is an existing entry with this title, to edit it please use <kure edit>")
		}

		buf, err := proto.Marshal(entry)
		if err != nil {
			return errors.Wrap(err, "marshal proto")
		}

		if err := b.Put(entry.Title, buf); err != nil {
			return errors.Wrap(err, "save entry")
		}

		return nil
	})
}

// DeleteEntry removes an entry from the database.
func DeleteEntry(title string) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(defaultBucket)
		if b == nil {
			return fmt.Errorf("%s folder does not exist", defaultBucket)
		}

		t := strings.ToLower(title)

		if err := b.Delete([]byte(t)); err != nil {
			return errors.Wrap(err, "delete entry")
		}

		return nil
	})
}

// EditEntry edits an entry
func EditEntry(entry *entry.Entry) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(defaultBucket)
		if b == nil {
			return fmt.Errorf("%s folder does not exist", defaultBucket)
		}

		buf, err := proto.Marshal(entry)
		if err != nil {
			return errors.Wrap(err, "marshal proto")
		}

		if err := b.Put(entry.Title, buf); err != nil {
			return errors.Wrap(err, "edit entry")
		}

		return nil
	})
}

// GetEntry retrieves all the entries stored in the database.
func GetEntry(title string) (*entry.Entry, error) {
	e := &entry.Entry{}
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(defaultBucket)
		if b == nil {
			return fmt.Errorf("%s folder does not exist", defaultBucket)
		}

		t := strings.ToLower(title)

		result := b.Get([]byte(t))

		if err := proto.Unmarshal(result, e); err != nil {
			return errors.Wrap(err, "unmarshal proto")
		}

		if string(e.Title) == "" {
			return fmt.Errorf("%s does not exist", title)
		}

		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "view entry")
	}

	return e, nil
}

// ListEntries retrieves a list with all the entries.
func ListEntries() ([]*entry.Entry, error) {
	var entries []*entry.Entry

	expired := make(chan bool)
	errCh := make(chan error)

	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(defaultBucket)
		if b == nil {
			return fmt.Errorf("%s folder does not exist", defaultBucket)
		}

		c := b.Cursor()

		// Place cursor in the first line of the bucket and move it to the next one
		for k, v := c.First(); k != nil; k, v = c.Next() {
			entry := &entry.Entry{}
			if err := proto.Unmarshal(v, entry); err != nil {
				return errors.Wrap(err, "unmarshal proto")
			}

			go cleanExpired(b, entry.Title, entry.Expires, expired, errCh)

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
		return nil, errors.Wrap(err, "list entries")
	}

	return entries, nil
}
