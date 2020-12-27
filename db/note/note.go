package note

import (
	"strings"

	"github.com/GGP1/kure/crypt"
	"github.com/GGP1/kure/pb"

	"github.com/awnumar/memguard"
	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"
)

var (
	noteBucket       = []byte("kure_note")
	errInvalidBucket = errors.New("invalid bucket")
)

// Create a new note. It destroys the locked buffer passed.
func Create(db *bolt.DB, lockedBuf *memguard.LockedBuffer, note *pb.Note) error {
	return db.Update(func(tx *bolt.Tx) error {
		defer lockedBuf.Destroy()

		b := tx.Bucket(noteBucket)
		if b == nil {
			return errInvalidBucket
		}

		buf, err := proto.Marshal(note)
		if err != nil {
			return errors.Wrap(err, "marshal note")
		}

		encNote, err := crypt.Encrypt(buf)
		if err != nil {
			return errors.Wrap(err, "encrypt note")
		}

		if err := b.Put([]byte(note.Name), encNote); err != nil {
			return errors.Wrap(err, "save note")
		}
		return nil
	})
}

// Get retrieves the note with the specified name.
func Get(db *bolt.DB, name string) (*memguard.LockedBuffer, *pb.Note, error) {
	buf, n := pb.SecureNote()

	err := db.View(func(tx *bolt.Tx) error {
		name = strings.TrimSpace(strings.ToLower(name))
		b := tx.Bucket(noteBucket)
		if b == nil {
			return errInvalidBucket
		}

		encNote := b.Get([]byte(name))
		if encNote == nil {
			return errors.Errorf("%q does not exist", name)
		}

		decNote, err := crypt.Decrypt(encNote)
		if err != nil {
			return errors.Wrap(err, "decrypt note")
		}

		if err := proto.Unmarshal(decNote, n); err != nil {
			return errors.Wrap(err, "unmarshal note")
		}

		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return buf, n, nil
}

// List returns a list with all the notes.
func List(db *bolt.DB) (*memguard.LockedBuffer, []*pb.Note, error) {
	buf, notes := pb.SecureNoteSlice()

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(noteBucket)
		if b == nil {
			return errInvalidBucket
		}

		return b.ForEach(func(k, v []byte) error {
			note := &pb.Note{}

			decNote, err := crypt.Decrypt(v)
			if err != nil {
				return errors.Wrap(err, "decrypt note")
			}

			if err := proto.Unmarshal(decNote, note); err != nil {
				return errors.Wrap(err, "unmarshal note")
			}

			notes = append(notes, note)
			return nil
		})
	})
	if err != nil {
		return nil, nil, err
	}

	return buf, notes, nil
}

// ListFastest is used to check if the user entered the correct password
// by trying to decrypt every record and returning the fastest result.
func ListFastest(db *bolt.DB) bool {
	succeed := make(chan bool)

	decrypt := func(v []byte) {
		_, err := crypt.Decrypt(v)
		if err != nil {
			succeed <- false
		}

		succeed <- true
	}

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(noteBucket)

		return b.ForEach(func(_, v []byte) error {
			go decrypt(v)
			return nil
		})
	})

	return <-succeed
}

// ListNames returns a list with all the notes names.
func ListNames(db *bolt.DB) ([]string, error) {
	var notes []string

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(noteBucket)
		if b == nil {
			return errInvalidBucket
		}

		return b.ForEach(func(k, _ []byte) error {
			notes = append(notes, string(k))
			return nil
		})
	})
	if err != nil {
		return nil, err
	}

	return notes, nil
}

// Remove removes a note from the database.
func Remove(db *bolt.DB, name string) error {
	return db.Update(func(tx *bolt.Tx) error {
		name = strings.TrimSpace(strings.ToLower(name))

		b := tx.Bucket(noteBucket)
		if b == nil {
			return errInvalidBucket
		}

		if err := b.Delete([]byte(name)); err != nil {
			return errors.Wrap(err, "remove note")
		}

		return nil
	})
}
