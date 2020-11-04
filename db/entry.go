package db

import (
	"strings"
	"time"

	"github.com/GGP1/kure/crypt"
	"github.com/GGP1/kure/pb"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
)

// CreateEntry creates a new record.
func CreateEntry(entry *pb.Entry) error {
	entries, err := EntriesByName(entry.Name)
	if err == nil {
		return err
	}

	var exists bool
	for _, e := range entries {
		if strings.Split(e.Name, "/")[0] == entry.Name {
			exists = true
			break
		}
	}

	if exists {
		return errors.New("already exists an entry or folder with this name, use <kure edit> to edit")
	}

	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(entryBucket)

		exists := b.Get([]byte(entry.Name))
		if exists != nil {
			return errors.Errorf("there is an existing entry called %s, use <kure edit> to edit", entry.Name)
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

// EditEntry edits an entry.
func EditEntry(name string, entry *pb.Entry) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(entryBucket)
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

// EntriesByName filters all the entries and returns those matching with the name passed.
func EntriesByName(name string) ([]*pb.Entry, error) {
	var group []*pb.Entry
	name = strings.TrimSpace(strings.ToLower(name))

	entries, err := ListEntries()
	if err != nil {
		return nil, err
	}

	for _, e := range entries {
		if strings.Contains(e.Name, name) {
			group = append(group, e)
		}
	}

	if len(group) == 0 {
		return nil, errors.New("no entries were found")
	}

	return group, nil
}

// GetEntry retrieves the entry with the specified name.
func GetEntry(name string) (*pb.Entry, error) {
	entry := &pb.Entry{}
	expired := make(chan bool, 1)
	errCh := make(chan error, 1)

	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(entryBucket)
		name = strings.TrimSpace(strings.ToLower(name))

		encEntry := b.Get([]byte(name))
		if encEntry == nil {
			return errors.Errorf("\"%s\" does not exist", name)
		}

		decEntry, err := crypt.Decrypt(encEntry)
		if err != nil {
			return errors.Wrap(err, "decrypt entry")
		}

		if err := proto.Unmarshal(decEntry, entry); err != nil {
			return errors.Wrap(err, "unmarshal entry")
		}

		go cleanExpired(b, entry.Name, entry.Expires, expired, errCh)

		select {
		case e := <-expired:
			if e {
				return errors.Errorf("\"%s\" expired", name)
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

// ListEntries returns a list with all the entries.
func ListEntries() ([]*pb.Entry, error) {
	var entries []*pb.Entry
	expired := make(chan bool, 1)
	errCh := make(chan error, 1)

	password, err := crypt.GetMasterPassword()
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(entryBucket)
		c := b.Cursor()

		// Place cursor in the first line of the bucket and move it to the next one
		for k, v := c.First(); k != nil; k, v = c.Next() {
			entry := &pb.Entry{}

			decEntry, err := crypt.DecryptX(v, password)
			if err != nil {
				return errors.Wrap(err, "decrypt entry")
			}

			if err := proto.Unmarshal(decEntry, entry); err != nil {
				return errors.Wrap(err, "unmarshal entry")
			}

			go cleanExpired(b, entry.Name, entry.Expires, expired, errCh)

			select {
			case e := <-expired:
				if !e {
					entries = append(entries, entry)
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

// RemoveEntry removes an entry from the database.
func RemoveEntry(name string) error {
	_, err := GetEntry(name)
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

// cleanExpired removes the expired entry from the database.
func cleanExpired(b *bolt.Bucket, name, expires string, expired chan<- bool, errCh chan<- error) {
	if expires == "Never" {
		expired <- false
		return
	}

	// Format expires time and time.Now to compare them
	expiration, err := time.Parse(time.RFC1123Z, expires)
	if err != nil {
		errCh <- errors.Wrap(err, "expiration time parse")
	}

	now, err := time.Parse(time.RFC1123Z, time.Now().Format(time.RFC1123Z))
	if err != nil {
		errCh <- errors.Wrap(err, "time now parse")
	}

	// Delete expired entries
	if now.Sub(expiration) >= 0 {
		if err := b.Delete([]byte(name)); err != nil {
			errCh <- errors.Wrap(err, "remove expired entry")
		}
		expired <- true
		return
	}

	expired <- false
}
