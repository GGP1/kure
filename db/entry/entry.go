package entry

import (
	dbutil "github.com/GGP1/kure/db"
	"github.com/GGP1/kure/pb"

	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
)

// Create new entries.
func Create(db *bolt.DB, entries ...*pb.Entry) error {
	if len(entries) == 0 {
		return nil
	}

	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(dbutil.EntryBucket)
		for _, entry := range entries {
			if err := dbutil.Put(b, entry); err != nil {
				return err
			}
		}

		return nil
	})
}

// Get retrieves the entry with the specified name.
func Get(db *bolt.DB, name string) (*pb.Entry, error) {
	entry := &pb.Entry{}
	if err := dbutil.Get(db, name, entry); err != nil {
		return nil, err
	}

	return entry, nil
}

// List returns a list with all the entries.
func List(db *bolt.DB) ([]*pb.Entry, error) {
	return dbutil.List(db, &pb.Entry{})
}

// ListNames returns a list with all the entries names.
func ListNames(db *bolt.DB) ([]string, error) {
	return dbutil.ListNames(db, dbutil.EntryBucket)
}

// Remove removes one or more entries from the database.
func Remove(db *bolt.DB, names ...string) error {
	return dbutil.Remove(db, dbutil.EntryBucket, names...)
}

// Update updates an entry, it removes the old one if the name differs.
func Update(db *bolt.DB, oldName string, entry *pb.Entry) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(dbutil.EntryBucket)
		if oldName != entry.Name {
			if err := b.Delete([]byte(oldName)); err != nil {
				return errors.Wrap(err, "remove old entry")
			}
		}
		return dbutil.Put(b, entry)
	})
}
