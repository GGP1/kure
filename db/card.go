package db

import (
	"fmt"
	"strings"

	"github.com/GGP1/kure/crypt"
	"github.com/GGP1/kure/model/card"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
)

// CreateCard creates a new bank card.
func CreateCard(card *card.Card) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(cardBucket)

		exists := b.Get([]byte(card.Name))
		if exists != nil {
			return errors.New("already exists a card with this name")
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

// DeleteCard removes a card from the database.
func DeleteCard(name string) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(cardBucket)
		n := strings.ToLower(name)

		if err := b.Delete([]byte(n)); err != nil {
			return errors.Wrap(err, "delete card")
		}

		return nil
	})
}

// GetCard retrieves the card with the specified name.
func GetCard(name string) (*card.Card, error) {
	c := &card.Card{}

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(cardBucket)
		n := strings.ToLower(name)

		result := b.Get([]byte(n))
		if result == nil {
			return errors.Errorf("\"%s\" does not exist", name)
		}

		decCard, err := crypt.Decrypt(result)
		if err != nil {
			return errors.Wrap(err, "decrypt card")
		}

		if err := proto.Unmarshal(decCard, c); err != nil {
			return errors.Wrap(err, "unmarshal card")
		}

		if c.Name == "" {
			return fmt.Errorf("%s does not exist", name)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return c, nil
}

// ListCards returns a list with all the cards.
func ListCards() ([]*card.Card, error) {
	var cards []*card.Card

	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(cardBucket)
		c := b.Cursor()

		// Place cursor in the first line of the bucket and move it to the next one
		for k, v := c.First(); k != nil; k, v = c.Next() {
			card := &card.Card{}

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
