package card

import (
	"strings"

	"github.com/GGP1/kure/crypt"
	dbutil "github.com/GGP1/kure/db"
	"github.com/GGP1/kure/pb"

	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"
)

// Create a new bank card.
func Create(db *bolt.DB, card *pb.Card) error {
	if strings.ContainsRune(card.Name, '\x00') {
		return errors.New("card name contains null characters")
	}

	return db.Batch(func(tx *bolt.Tx) error {
		b := tx.Bucket(dbutil.CardBucket)
		return save(b, card)
	})
}

// Get retrieves the card with the specified name.
func Get(db *bolt.DB, name string) (*pb.Card, error) {
	card := &pb.Card{}

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(dbutil.CardBucket)

		encCard := b.Get([]byte(name))
		if encCard == nil {
			return errors.Errorf("card %q does not exist", name)
		}

		decCard, err := crypt.Decrypt(encCard)
		if err != nil {
			return errors.Wrap(err, "decrypt card")
		}

		if err := proto.Unmarshal(decCard, card); err != nil {
			return errors.Wrap(err, "unmarshal card")
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return card, nil
}

// List returns a list with all the cards.
func List(db *bolt.DB) ([]*pb.Card, error) {
	tx, err := db.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	b := tx.Bucket(dbutil.CardBucket)
	cards := make([]*pb.Card, 0, b.Stats().KeyN)

	err = b.ForEach(func(k, v []byte) error {
		card := &pb.Card{}

		decCard, err := crypt.Decrypt(v)
		if err != nil {
			return errors.Wrap(err, "decrypt card")
		}

		if err := proto.Unmarshal(decCard, card); err != nil {
			return errors.Wrap(err, "unmarshal card")
		}

		cards = append(cards, card)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return cards, nil
}

// ListNames returns a list with all the cards names.
func ListNames(db *bolt.DB) ([]string, error) {
	return dbutil.ListNames(db, dbutil.CardBucket)
}

// Remove removes a card from the database.
func Remove(db *bolt.DB, name string) error {
	return dbutil.Remove(db, dbutil.CardBucket, name)
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
		return save(b, card)
	})
}

func save(b *bolt.Bucket, card *pb.Card) error {
	buf, err := proto.Marshal(card)
	if err != nil {
		return errors.Wrap(err, "marshal card")
	}

	encCard, err := crypt.Encrypt(buf)
	if err != nil {
		return errors.Wrap(err, "encrypt card")
	}

	if err := b.Put([]byte(card.Name), encCard); err != nil {
		return errors.Wrap(err, "save card")
	}

	return nil
}
