package card

import (
	"strings"

	dbutil "github.com/GGP1/kure/db"
	"github.com/GGP1/kure/pb"

	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
)

// Create a new bank card.
func Create(db *bolt.DB, card *pb.Card) error {
	return db.Batch(func(tx *bolt.Tx) error {
		b := tx.Bucket(dbutil.CardBucket)
		return dbutil.Put(b, card)
	})
}

// Get retrieves the card with the specified name.
func Get(db *bolt.DB, name string) (*pb.Card, error) {
	card := &pb.Card{}
	if err := dbutil.Get(db, name, card); err != nil {
		return nil, err
	}

	return card, nil
}

// List returns a list with all the cards.
func List(db *bolt.DB) ([]*pb.Card, error) {
	return dbutil.List(db, &pb.Card{})
}

// ListNames returns a list with all the cards names.
func ListNames(db *bolt.DB) ([]string, error) {
	return dbutil.ListNames(db, dbutil.CardBucket)
}

// Remove removes one or more cards from the database.
func Remove(db *bolt.DB, names ...string) error {
	return dbutil.Remove(db, dbutil.CardBucket, names...)
}

// Update updates a card, it removes the old one if the name differs.
func Update(db *bolt.DB, oldName string, card *pb.Card) error {
	if strings.ContainsRune(card.Name, '\x00') {
		return errors.New("entry name contains null characters")
	}

	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(dbutil.CardBucket)
		if oldName != card.Name {
			if err := b.Delete([]byte(oldName)); err != nil {
				return errors.Wrap(err, "remove old card")
			}
		}
		return dbutil.Put(b, card)
	})
}
