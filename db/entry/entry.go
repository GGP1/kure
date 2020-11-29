package entry

import (
	"strings"
	"time"

	"github.com/GGP1/kure/crypt"
	"github.com/GGP1/kure/pb"

	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"
)

var (
	entryBucket      = []byte("kure_entry")
	errInvalidBucket = errors.New("invalid bucket")
)

// Create creates a new record.
func Create(db *bolt.DB, entry *pb.Entry) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(entryBucket)
		if b == nil {
			return errInvalidBucket
		}

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

// Edit edits an entry.
func Edit(db *bolt.DB, name string, entry *pb.Entry) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(entryBucket)
		if b == nil {
			return errInvalidBucket
		}
		name = strings.TrimSpace(strings.ToLower(name))

		if err := b.Delete([]byte(name)); err != nil {
			return errors.Wrap(err, "delete old entry")
		}

		buf, err := proto.Marshal(entry)
		if err != nil {
			return errors.Wrap(err, "marshal entry")
		}

		encEntry, err := crypt.Encrypt(buf)
		if err != nil {
			return errors.Wrap(err, "encrypt entry")
		}

		if err := b.Put([]byte(entry.Name), encEntry); err != nil {
			return errors.Wrap(err, "edit entry")
		}

		return nil
	})
}

// Get retrieves the entry with the specified name.
func Get(db *bolt.DB, name string) (*pb.Entry, error) {
	entry := &pb.Entry{}
	expired := make(chan bool, 1)
	errCh := make(chan error, 1)

	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(entryBucket)
		if b == nil {
			return errInvalidBucket
		}
		name = strings.TrimSpace(strings.ToLower(name))

		encEntry := b.Get([]byte(name))
		if encEntry == nil {
			return errors.Errorf("%q does not exist", name)
		}

		decEntry, err := crypt.Decrypt(encEntry)
		if err != nil {
			return errors.Wrap(err, "decrypt entry")
		}

		if err := proto.Unmarshal(decEntry, entry); err != nil {
			return errors.Wrap(err, "unmarshal entry")
		}

		go isExpired(entry.Expires, expired, errCh)

		select {
		case e := <-expired:
			if e {
				if err := b.Delete([]byte(entry.Name)); err != nil {
					return errors.Wrap(err, "remove expired entry")
				}
				return errors.Errorf("%q expired", name)
			}
		case err := <-errCh:
			return err
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
	var entries []*pb.Entry
	expired := make(chan bool, 1)
	errCh := make(chan error, 1)

	_, err := crypt.GetMasterPassword()
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(entryBucket)
		if b == nil {
			return errInvalidBucket
		}
		c := b.Cursor()

		// Place cursor in the first line of the bucket and move it to the next one
		for k, v := c.First(); k != nil; k, v = c.Next() {
			entry := &pb.Entry{}

			decEntry, err := crypt.Decrypt(v)
			if err != nil {
				return errors.Wrap(err, "decrypt entry")
			}

			if err := proto.Unmarshal(decEntry, entry); err != nil {
				return errors.Wrap(err, "unmarshal entry")
			}

			go isExpired(entry.Expires, expired, errCh)

			select {
			case e := <-expired:
				if !e {
					entries = append(entries, entry)
					continue
				}

				if err := b.Delete([]byte(entry.Name)); err != nil {
					return errors.Wrap(err, "remove expired entry")
				}
			case err := <-errCh:
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return entries, nil
}

// ListByName filters all the entries and returns those matching with the name passed.
func ListByName(db *bolt.DB, name string) ([]*pb.Entry, error) {
	var group []*pb.Entry
	name = strings.TrimSpace(strings.ToLower(name))

	entries, err := List(db)
	if err != nil {
		return nil, err
	}

	for _, e := range entries {
		if strings.Contains(e.Name, name) {
			group = append(group, e)
		}
	}

	return group, nil
}

// ListNames returns a list with all the entries names.
func ListNames(db *bolt.DB) ([]*pb.EntryList, error) {
	var entries []*pb.EntryList
	expired := make(chan bool, 1)
	errCh := make(chan error, 1)

	_, err := crypt.GetMasterPassword()
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(entryBucket)
		if b == nil {
			return errInvalidBucket
		}
		c := b.Cursor()

		// Place cursor in the first line of the bucket and move it to the next one
		for k, v := c.First(); k != nil; k, v = c.Next() {
			entry := &pb.EntryList{}

			decEntry, err := crypt.Decrypt(v)
			if err != nil {
				return errors.Wrap(err, "decrypt entry")
			}

			if err := proto.Unmarshal(decEntry, entry); err != nil {
				return errors.Wrap(err, "unmarshal entry")
			}

			go isExpired(entry.Expires, expired, errCh)

			select {
			case e := <-expired:
				if !e {
					entries = append(entries, entry)
					continue
				}

				if err := b.Delete([]byte(entry.Name)); err != nil {
					return errors.Wrap(err, "remove expired entry")
				}
			case err := <-errCh:
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return entries, nil
}

// Remove removes an entry from the database.
func Remove(db *bolt.DB, name string) error {
	_, err := Get(db, name)
	if err != nil {
		return err
	}

	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(entryBucket)
		name = strings.TrimSpace(strings.ToLower(name))

		if err := b.Delete([]byte(name)); err != nil {
			return errors.Wrap(err, "remove entry")
		}

		return nil
	})
}

// isExpired removes the expired entry from the database.
func isExpired(expires string, expired chan<- bool, errCh chan<- error) {
	if expires == "Never" {
		expired <- false
		return
	}

	// Format expires time and time.Now to compare them
	expiration, err := time.Parse(time.RFC1123Z, expires)
	if err != nil {
		errCh <- errors.Wrap(err, "expiration parse")
	}

	now, err := time.Parse(time.RFC1123Z, time.Now().Format(time.RFC1123Z))
	if err != nil {
		errCh <- errors.Wrap(err, "now parse")
	}

	// Delete expired entries
	if now.Sub(expiration) >= 0 {
		expired <- true
		return
	}

	expired <- false
}
