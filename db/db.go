package db

import (
	"time"

	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
)

var (
	db           *bolt.DB
	entryBucket  = []byte("kure_entry")
	cardBucket   = []byte("kure_card")
	walletBucket = []byte("kure_wallet")
)

// Init initializes the database connection.
func Init(path string) error {
	var err error

	db, err = bolt.Open(path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return errors.Wrap(err, "open the database")
	}

	return CreateBucketIfNotExists(cardBucket, entryBucket, walletBucket)
}

// CreateBucketIfNotExists creates a new bucket if it doesn't already exist.
func CreateBucketIfNotExists(bucketName ...[]byte) error {
	if bucketName == nil {
		return errors.New("invalid bucket name")
	}

	return db.Update(func(tx *bolt.Tx) error {
		for _, bucket := range bucketName {
			_, err := tx.CreateBucketIfNotExists(bucket)
			if err != nil {
				return errors.Wrapf(err, "create bucket %s", bucketName)
			}
		}

		return nil
	})
}

// ListOfBuckets returns a list of the existing buckets.
func ListOfBuckets() ([]string, error) {
	var buckets []string
	err := db.View(func(tx *bolt.Tx) error {
		return tx.ForEach(func(name []byte, _ *bolt.Bucket) error {
			buckets = append(buckets, string(name))
			return nil
		})
	})
	if err != nil {
		return nil, errors.Wrap(err, "list of buckets")
	}

	return buckets, nil
}
