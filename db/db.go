package db

import (
	"fmt"
	"time"

	bolt "go.etcd.io/bbolt"
)

var (
	db         *bolt.DB
	bucketName = []byte("kure")
)

// Init initializes the database connection.
func Init(path string) error {
	var err error

	p := fmt.Sprintf("%s/kure.db", path)

	db, err = bolt.Open(p, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return err
	}

	return CreateBucketIfNotExists()
}

// CreateBucketIfNotExists creates a new bucket in case it doesn't exist.
func CreateBucketIfNotExists() error {
	return db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketName)
		if err != nil {
			return fmt.Errorf("create bucket failed: %v", err)
		}

		return nil
	})
}
