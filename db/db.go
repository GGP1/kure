package dbutil

import (
	"testing"
	"time"

	"github.com/GGP1/kure/config"

	"github.com/awnumar/memguard"
	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
)

// Database bucket
var (
	CardBucket  = []byte("kure_card")
	EntryBucket = []byte("kure_entry")
	FileBucket  = []byte("kure_file")
	TOTPBucket  = []byte("kure_totp")
)

// ListNames returns a list with all the records names.
func ListNames(db *bolt.DB, bucketName []byte) ([]string, error) {
	tx, err := db.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// b will be nil only if the user attempts to add
	// a record on registration
	b := tx.Bucket(bucketName)
	if b == nil {
		return nil, nil
	}

	records := make([]string, 0, b.Stats().KeyN)
	_ = b.ForEach(func(k, _ []byte) error {
		records = append(records, string(k))
		return nil
	})

	return records, nil
}

// Remove removes a record from the database.
func Remove(db *bolt.DB, bucketName []byte, name string) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		if err := b.Delete([]byte(name)); err != nil {
			return errors.Wrap(err, "remove record")
		}
		return nil
	})
}

// SetContext creates a bucket and its context to test the database operations.
func SetContext(t testing.TB, path string, bucketName []byte) *bolt.DB {
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		t.Fatalf("Failed connecting to the database: %v", err)
	}

	config.Reset()
	// Reduce argon2 parameters to speed up tests
	auth := map[string]interface{}{
		"password":   memguard.NewEnclave([]byte("1")),
		"iterations": 1,
		"memory":     1,
		"threads":    1,
	}
	config.Set("auth", auth)

	err = db.Update(func(tx *bolt.Tx) error {
		tx.DeleteBucket(bucketName)
		if _, err := tx.CreateBucketIfNotExists(bucketName); err != nil {
			return errors.Wrapf(err, "couldn't create %q bucket", bucketName)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("Failed closing the database: %v", err)
		}
	})

	return db
}
