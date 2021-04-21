package totp

import (
	"strings"

	"github.com/GGP1/kure/crypt"
	"github.com/GGP1/kure/pb"

	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"
)

var totpBucket = []byte("kure_totp")

// Create a new TOTP.
func Create(db *bolt.DB, totp *pb.TOTP) error {
	return db.Batch(func(tx *bolt.Tx) error {
		// Ensure the name does not contain null characters
		if strings.ContainsRune(totp.Name, '\x00') {
			return errors.Errorf("TOTP name contains null characters")
		}

		b := tx.Bucket(totpBucket)

		buf, err := proto.Marshal(totp)
		if err != nil {
			return errors.Wrap(err, "marshal TOTP")
		}

		encTOTP, err := crypt.Encrypt(buf)
		if err != nil {
			return errors.Wrap(err, "encrypt TOTP")
		}

		if err := b.Put([]byte(totp.Name), encTOTP); err != nil {
			return errors.Wrap(err, "saving TOTP")
		}

		return nil
	})
}

// Get retrieves the TOTP with the specified name.
func Get(db *bolt.DB, name string) (*pb.TOTP, error) {
	totp := &pb.TOTP{}

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(totpBucket)

		encTOTP := b.Get([]byte(name))
		if encTOTP == nil {
			return errors.Errorf("%q has no TOTP set", name)
		}

		decTOTP, err := crypt.Decrypt(encTOTP)
		if err != nil {
			return errors.Wrap(err, "decrypt TOTP")
		}

		if err := proto.Unmarshal(decTOTP, totp); err != nil {
			return errors.Wrap(err, "unmarshal TOTP")
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return totp, nil
}

// List returns a list with all the TOTPs.
// Since this function is used more frequently than other objects' List
// functions, make it as efficient as possible.
func List(db *bolt.DB) ([]*pb.TOTP, error) {
	tx, err := db.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	b := tx.Bucket(totpBucket)
	totps := make([]*pb.TOTP, 0, b.Stats().KeyN)

	err = b.ForEach(func(k, v []byte) error {
		totp := &pb.TOTP{}

		decTOTP, err := crypt.Decrypt(v)
		if err != nil {
			return errors.Wrap(err, "decrypt TOTP")
		}

		if err := proto.Unmarshal(decTOTP, totp); err != nil {
			return errors.Wrap(err, "unmarshal TOTP")
		}

		totps = append(totps, totp)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return totps, nil
}

// ListNames returns a slice with all the totps names.
func ListNames(db *bolt.DB) ([]string, error) {
	tx, err := db.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// b will be nil only if the user attempts to add
	// a file on registration
	b := tx.Bucket(totpBucket)
	if b == nil {
		return nil, nil
	}

	totps := make([]string, 0, b.Stats().KeyN)
	_ = b.ForEach(func(k, _ []byte) error {
		totps = append(totps, string(k))
		return nil
	})

	return totps, nil
}

// Remove removes a totp from the database.
func Remove(db *bolt.DB, name string) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(totpBucket)

		if err := b.Delete([]byte(name)); err != nil {
			return errors.Wrap(err, "remove TOTP")
		}

		return nil
	})
}
