package entry

import (
	"strings"

	"github.com/GGP1/kure/crypt"
	"github.com/GGP1/kure/pb"

	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"
)

var entryBucket = []byte("kure_entry")

// Create a new entry.
func Create(db *bolt.DB, entry *pb.Entry) error {
	return db.Batch(func(tx *bolt.Tx) error {
		// Ensure the name does not contain null characters
		if strings.ContainsRune(entry.Name, '\x00') {
			return errors.New("entry name contains null characters")
		}

		b := tx.Bucket(entryBucket)

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

// Get retrieves the entry with the specified name.
func Get(db *bolt.DB, name string) (*pb.Entry, error) {
	entry := &pb.Entry{}

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(entryBucket)

		encEntry := b.Get([]byte(name))
		if encEntry == nil {
			return errors.Errorf("entry %q does not exist", name)
		}

		decEntry, err := crypt.Decrypt(encEntry)
		if err != nil {
			return errors.Wrap(err, "decrypt entry")
		}

		if err := proto.Unmarshal(decEntry, entry); err != nil {
			return errors.Wrap(err, "unmarshal entry")
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return entry, nil
}

// List returns a list with all the entries.
func List(db *bolt.DB) ([]*pb.Entry, error) {
	tx, err := db.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	b := tx.Bucket(entryBucket)
	entries := make([]*pb.Entry, 0, b.Stats().KeyN)

	err = b.ForEach(func(k, v []byte) error {
		entry := &pb.Entry{}

		decEntry, err := crypt.Decrypt(v)
		if err != nil {
			return errors.Wrap(err, "decrypt entry")
		}

		if err := proto.Unmarshal(decEntry, entry); err != nil {
			return errors.Wrap(err, "unmarshal entry")
		}

		entries = append(entries, entry)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return entries, nil
}

// ListNames returns a list with all the entries names.
func ListNames(db *bolt.DB) ([]string, error) {
	tx, err := db.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// b will be nil only if the user attempts to add
	// an entry on registration
	b := tx.Bucket(entryBucket)
	if b == nil {
		return nil, nil
	}

	entries := make([]string, 0, b.Stats().KeyN)
	_ = b.ForEach(func(k, _ []byte) error {
		entries = append(entries, string(k))
		return nil
	})

	return entries, nil
}

// Remove removes an entry from the database.
func Remove(db *bolt.DB, name string) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(entryBucket)

		if err := b.Delete([]byte(name)); err != nil {
			return errors.Wrap(err, "remove entry")
		}

		return nil
	})
}
