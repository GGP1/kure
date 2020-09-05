package db

import (
	"encoding/json"
	"strings"

	"github.com/GGP1/kure/entry"

	bolt "go.etcd.io/bbolt"
)

// CreateEntry creates a new record.
func CreateEntry(entry *entry.Entry) error {
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)

		entry.Title = strings.Title(entry.Title)

		buf, err := json.Marshal(entry)
		if err != nil {
			return err
		}

		title := strings.ToLower(entry.Title)

		return b.Put([]byte(title), buf)
	})
	if err != nil {
		return err
	}
	return nil
}

// DeleteEntry removes an entry from the database.
func DeleteEntry(titles []string) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		for _, t := range titles {
			t = strings.ToLower(t)
			err := b.Delete([]byte(t))
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// GetEntry retrieves all the entries stored in the database.
func GetEntry(title string) (entry.Entry, error) {
	var e entry.Entry

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		t := strings.ToLower(title)

		entry := b.Get([]byte(t))

		if err := json.Unmarshal(entry, &e); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return entry.Entry{}, err
	}

	return e, nil
}

// ListEntries retrieves a list with all the entries.
func ListEntries() ([]entry.Entry, error) {
	var (
		entries []entry.Entry
		entry   entry.Entry
	)

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		c := b.Cursor()

		// Place cursor in the first line of the bucket and move it to the next one
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if err := json.Unmarshal(v, &entry); err != nil {
				return err
			}

			entries = append(entries, entry)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return entries, nil
}
