package card

import (
	"strings"

	"github.com/GGP1/kure/crypt"
	"github.com/GGP1/kure/pb"

	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"
)

var (
	cardBucket       = []byte("kure_card")
	errInvalidBucket = errors.New("invalid bucket")
)

// Create creates a new bank card.
func Create(db *bolt.DB, card *pb.Card) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(cardBucket)
		if b == nil {
			return errInvalidBucket
		}

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
	})
}

// Get retrieves the card with the specified name.
func Get(db *bolt.DB, name string) (*pb.Card, error) {
	c := &pb.Card{}

	err := db.View(func(tx *bolt.Tx) error {
		name = strings.TrimSpace(strings.ToLower(name))
		b := tx.Bucket(cardBucket)
		if b == nil {
			return errInvalidBucket
		}

		encCard := b.Get([]byte(name))
		if encCard == nil {
			return errors.Errorf("%q does not exist", name)
		}

		decCard, err := crypt.Decrypt(encCard)
		if err != nil {
			return errors.Wrap(err, "decrypt card")
		}

		if err := proto.Unmarshal(decCard, c); err != nil {
			return errors.Wrap(err, "unmarshal card")
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return c, nil
}

// List returns a list with all the cards.
func List(db *bolt.DB) ([]*pb.Card, error) {
	var cards []*pb.Card

	_, err := crypt.GetMasterPassword()
	if err != nil {
		return nil, err
	}

	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(cardBucket)
		if b == nil {
			return errInvalidBucket
		}
		c := b.Cursor()

		// Place cursor in the first line of the bucket and move it to the next one
		for k, v := c.First(); k != nil; k, v = c.Next() {
			card := &pb.Card{}

			decCard, err := crypt.Decrypt(v)
			if err != nil {
				return errors.Wrap(err, "decrypt card")
			}

			if err := proto.Unmarshal(decCard, card); err != nil {
				return errors.Wrap(err, "unmarshal card")
			}

			cards = append(cards, card)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return cards, nil
}

// ListByName filters all the entries and returns those matching with the name passed.
func ListByName(db *bolt.DB, name string) ([]*pb.Card, error) {
	var group []*pb.Card
	name = strings.TrimSpace(strings.ToLower(name))

	cards, err := List(db)
	if err != nil {
		return nil, err
	}

	for _, c := range cards {
		if strings.Contains(c.Name, name) {
			group = append(group, c)
		}
	}

	return group, nil
}

// ListNames returns a list with all the cards names.
func ListNames(db *bolt.DB) ([]*pb.CardList, error) {
	var cards []*pb.CardList

	_, err := crypt.GetMasterPassword()
	if err != nil {
		return nil, err
	}

	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(cardBucket)
		if b == nil {
			return errInvalidBucket
		}
		c := b.Cursor()

		// Place cursor in the first line of the bucket and move it to the next one
		for k, v := c.First(); k != nil; k, v = c.Next() {
			card := &pb.CardList{}

			decCard, err := crypt.Decrypt(v)
			if err != nil {
				return errors.Wrap(err, "decrypt card")
			}

			if err := proto.Unmarshal(decCard, card); err != nil {
				return errors.Wrap(err, "unmarshal card")
			}

			cards = append(cards, card)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return cards, nil
}

// Remove removes a card from the database.
func Remove(db *bolt.DB, name string) error {
	_, err := Get(db, name)
	if err != nil {
		return err
	}

	return db.Update(func(tx *bolt.Tx) error {
		name = strings.TrimSpace(strings.ToLower(name))
		b := tx.Bucket(cardBucket)

		if err := b.Delete([]byte(name)); err != nil {
			return errors.Wrap(err, "remove card")
		}

		return nil
	})
}
