package db

import (
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
)

var (
	db           *bolt.DB
	entryBucket  = []byte("kure_entry")
	cardBucket   = []byte("kure_card")
	fileBucket   = []byte("kure_file")
	walletBucket = []byte("kure_wallet")
)

// Init initializes the database connection.
func Init(path string) error {
	var err error

	db, err = bolt.Open(path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return errors.Wrap(err, "open the database")
	}

	return CreateBucketIfNotExists(cardBucket, entryBucket, fileBucket, walletBucket)
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
				return errors.Wrapf(err, "create bucket %s", bucket)
			}
		}

		return nil
	})
}

// HTTPBackup writes a consistent view of the database to a http endpoint.
func HTTPBackup(w http.ResponseWriter, r *http.Request) {
	err := db.View(func(tx *bolt.Tx) error {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", `attachment; filename="kure.db"`)
		w.Header().Set("Content-Length", strconv.Itoa(int(tx.Size())))
		_, err := tx.WriteTo(w)
		if err != nil {
			return errors.Wrap(err, "write database")
		}
		return nil
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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
		return nil, errors.Wrap(err, "failed listing buckets")
	}

	return buckets, nil
}

// Stats gives information about the database and its buckets.
func Stats() (map[string]int, error) {
	tx, err := db.Begin(false)
	if err != nil {
		return nil, errors.Wrap(err, "transaction failed")
	}
	defer tx.Rollback()

	cardStats := tx.Bucket(cardBucket).Stats()
	entryStats := tx.Bucket(entryBucket).Stats()
	fileStats := tx.Bucket(fileBucket).Stats()
	walletStats := tx.Bucket(walletBucket).Stats()

	stats := make(map[string]int, 4)
	stats["cards"] = cardStats.KeyN
	stats["entries"] = entryStats.KeyN
	stats["files"] = fileStats.KeyN
	stats["wallets"] = walletStats.KeyN

	return stats, nil
}

// WriteTo writes the entire database to a writer.
func WriteTo(w io.Writer) error {
	return db.View(func(tx *bolt.Tx) error {
		_, err := tx.WriteTo(w)
		if err != nil {
			return errors.Wrap(err, "write database")
		}
		return nil
	})
}
