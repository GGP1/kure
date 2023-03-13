package totp

import (
	dbutil "github.com/GGP1/kure/db"
	"github.com/GGP1/kure/pb"

	bolt "go.etcd.io/bbolt"
)

// Create a new TOTP.
func Create(db *bolt.DB, totp *pb.TOTP) error {
	return db.Update(func(tx *bolt.Tx) error {
		return dbutil.Put(tx, totp)
	})
}

// Get retrieves the TOTP with the specified name.
func Get(db *bolt.DB, name string) (*pb.TOTP, error) {
	totp := &pb.TOTP{}
	if err := dbutil.Get(db, name, totp); err != nil {
		return nil, err
	}

	return totp, nil
}

// List returns a list with all the TOTPs.
func List(db *bolt.DB) ([]*pb.TOTP, error) {
	return dbutil.List(db, &pb.TOTP{})
}

// ListNames returns a slice with all the totps names.
func ListNames(db *bolt.DB) ([]string, error) {
	return dbutil.ListNames[*pb.TOTP](db)
}

// Remove removes one or more totps from the database.
func Remove(db *bolt.DB, names ...string) error {
	return db.Update(func(tx *bolt.Tx) error {
		return dbutil.Remove[*pb.TOTP](tx, names...)
	})
}
