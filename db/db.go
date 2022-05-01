package dbutil

import (
	"strings"
	"testing"
	"time"

	"github.com/GGP1/kure/config"
	"github.com/GGP1/kure/crypt"
	"github.com/GGP1/kure/pb"

	"github.com/awnumar/memguard"
	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"
)

const nullChar = string('\x00')

// Database bucket
var (
	CardBucket  = []byte("kure_card")
	EntryBucket = []byte("kure_entry")
	FileBucket  = []byte("kure_file")
	TOTPBucket  = []byte("kure_totp")
)

// Record is an interface that all Kure objects implement.
type Record interface {
	GetName() string
	proto.Message
}

// Get retrieves a record from the database, decrypts it and loads it into record.
func Get(db *bolt.DB, name string, record Record) error {
	return db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(GetBucketName(record))

		encRecord := b.Get([]byte(name))
		if encRecord == nil {
			return errors.Errorf("record %q does not exist", name)
		}

		decRecord, err := crypt.Decrypt(encRecord)
		if err != nil {
			return errors.Wrap(err, "decrypt record")
		}

		if err := proto.Unmarshal(decRecord, record); err != nil {
			return errors.Wrap(err, "unmarshal record")
		}

		return nil
	})
}

// GetBucketName returns the bucket name depending on the type of the record passed.
func GetBucketName(r Record) []byte {
	switch r.(type) {
	case *pb.Card:
		return CardBucket
	case *pb.Entry:
		return EntryBucket
	case *pb.File, *pb.FileCheap:
		return FileBucket
	case *pb.TOTP:
		return TOTPBucket
	default:
		memguard.SafePanic("invalid object: " + r.GetName())
		return nil
	}
}

// List returns a list of decrypted records from the database.
func List[R Record](db *bolt.DB, record R) ([]R, error) {
	tx, err := db.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	b := tx.Bucket(GetBucketName(record))
	records := make([]R, 0, b.Stats().KeyN)

	err = b.ForEach(func(k, v []byte) error {
		decRecord, err := crypt.Decrypt(v)
		if err != nil {
			return errors.Wrap(err, "decrypt record")
		}

		if err := proto.Unmarshal(decRecord, record); err != nil {
			return errors.Wrap(err, "unmarshal record")
		}

		records = append(records, record)
		// Allocate a new protobuf object of type R
		record = record.ProtoReflect().New().Interface().(R)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return records, nil
}

// ListNames returns a list with all the records names.
func ListNames(db *bolt.DB, bucketName []byte) ([]string, error) {
	tx, err := db.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// b will be nil only if the user attempts to add
	// a record on registration as this method is being used
	// in checks previous to a command execution
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

// Put encrypts and saves a record into the database.
func Put(b *bolt.Bucket, record Record) error {
	name := strings.ReplaceAll(record.GetName(), nullChar, "")
	if name == "" {
		return errors.New("record name is empty")
	}

	buf, err := proto.Marshal(record)
	if err != nil {
		return errors.Wrap(err, "marshal record")
	}

	encRecord, err := crypt.Encrypt(buf)
	if err != nil {
		return errors.Wrap(err, "encrypt record")
	}

	if err := b.Put([]byte(name), encRecord); err != nil {
		return errors.Wrap(err, "store record")
	}

	return nil
}

// Remove removes records from the database.
func Remove(db *bolt.DB, bucketName []byte, names ...string) error {
	if len(names) == 0 {
		return nil
	}

	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		for _, name := range names {
			if err := b.Delete([]byte(name)); err != nil {
				return errors.Wrapf(err, "delete record %q", name)
			}
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
