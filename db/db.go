package dbutil

import (
	"bytes"
	"crypto/rand"
	"encoding/gob"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/GGP1/kure/config"
	"github.com/GGP1/kure/crypt"
	"github.com/GGP1/kure/db/bucket"
	"github.com/GGP1/kure/pb"

	"github.com/awnumar/memguard"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	bolt "go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"
)

var namesKey = []byte("names")

const nullChar = string('\x00')

// Record is an interface that all Kure objects implement.
type Record interface {
	GetName() string
	proto.Message
}

// Get retrieves a record from the database, decrypts it and loads it into record.
func Get(db *bolt.DB, name string, record Record) error {
	return db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(GetBucketName(record))

		key, err := getKey(tx, name, record)
		if err != nil {
			return err
		}

		encRecord := b.Get(key)
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
		return bucket.Card.GetName()
	case *pb.Entry:
		return bucket.Entry.GetName()
	case *pb.File, *pb.FileCheap:
		return bucket.File.GetName()
	case *pb.TOTP:
		return bucket.TOTP.GetName()
	default:
		memguard.SafePanic("invalid object: " + r.GetName())
		return nil
	}
}

// GetNamesBucketName returns the name of the bucket that stores the names of a specific
// type of record.
func GetNamesBucketName(r Record) []byte {
	switch r.(type) {
	case *pb.Card:
		return bucket.CardNames.GetName()
	case *pb.Entry:
		return bucket.EntryNames.GetName()
	case *pb.File, *pb.FileCheap:
		return bucket.FileNames.GetName()
	case *pb.TOTP:
		return bucket.TOTPNames.GetName()
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

	mp, err := getNames(tx, bucketName)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(mp))
	for name := range mp {
		names = append(names, name)
	}

	// Sort alphabetically
	sort.Slice(names, func(i, j int) bool { return names[i] < names[j] })
	return names, nil
}

// Put encrypts and saves a record into the database.
func Put(tx *bolt.Tx, record Record) error {
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

	key, err := putKey(tx, name, record)
	if err != nil {
		return err
	}

	b := tx.Bucket(GetBucketName(record))
	if err := b.Put(key, encRecord); err != nil {
		return errors.Wrap(err, "store record")
	}

	return nil
}

// Remove removes records from the database.
func Remove(tx *bolt.Tx, record Record, names ...string) error {
	if len(names) == 0 {
		return nil
	}

	namesBucketName := GetNamesBucketName(record)
	namesMap, err := getNames(tx, namesBucketName)
	if err != nil {
		return err
	}

	b := tx.Bucket(GetBucketName(record))
	for _, name := range names {
		if err := b.Delete([]byte(name)); err != nil {
			return errors.Wrapf(err, "delete record %q", name)
		}
		delete(namesMap, name)
	}

	return saveNamesMap(tx, namesBucketName, namesMap)
}

// SetContext creates a bucket and its context to test the database operations.
func SetContext(t testing.TB, bucketNames ...[]byte) *bolt.DB {
	dbFile, err := os.CreateTemp("", "*")
	assert.NoError(t, err)

	db, err := bolt.Open(dbFile.Name(), 0o600, &bolt.Options{Timeout: 1 * time.Second})
	assert.NoError(t, err, "Failed connecting to the database")

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
		for _, bucketName := range bucketNames {
			tx.DeleteBucket(bucketName)
			if _, err := tx.CreateBucket(bucketName); err != nil {
				return errors.Wrapf(err, "couldn't create %q bucket", bucketName)
			}
		}
		return nil
	})
	assert.NoError(t, err)

	t.Cleanup(func() {
		err := db.Close()
		assert.NoError(t, err, "Failed closing the database")
	})

	return db
}

// getKey returns the key corresponding to the name provided.
func getKey(tx *bolt.Tx, name string, record Record) ([]byte, error) {
	bucketName := GetNamesBucketName(record)
	mp, err := getNames(tx, bucketName)
	if err != nil {
		return nil, err
	}

	return mp[name], nil
}

func getNames(tx *bolt.Tx, bucketName []byte) (map[string][]byte, error) {
	namesMap := make(map[string][]byte, 0)
	b := tx.Bucket(bucketName)
	encMap := b.Get(namesKey)
	if encMap == nil {
		return namesMap, nil
	}

	decMap, err := crypt.Decrypt(encMap)
	if err != nil {
		return nil, errors.Wrap(err, "decrypt map")
	}

	reader := bytes.NewBuffer(decMap)
	if err := gob.NewDecoder(reader).Decode(&namesMap); err != nil {
		return nil, err
	}

	return namesMap, nil
}

func putKey(tx *bolt.Tx, name string, record Record) ([]byte, error) {
	bucketName := GetNamesBucketName(record)
	namesMap, err := getNames(tx, bucketName)
	if err != nil {
		return nil, err
	}

	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}

	namesMap[name] = key
	if err := saveNamesMap(tx, bucketName, namesMap); err != nil {
		return nil, err
	}

	return key, nil
}

func saveNamesMap(tx *bolt.Tx, bucketName []byte, namesMap map[string][]byte) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(namesMap); err != nil {
		return err
	}

	updatedMap, err := crypt.Encrypt(buf.Bytes())
	if err != nil {
		return errors.Wrap(err, "encrypt names")
	}

	b := tx.Bucket(bucketName)
	return b.Put(namesKey, updatedMap)
}
