package db

import (
	"strings"

	"github.com/GGP1/kure/crypt"
	"github.com/GGP1/kure/pb"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
)

// CardsByName filters all the entries and returns those matching with the name passed.
func CardsByName(name string) ([]*pb.Card, error) {
	var group []*pb.Card
	name = strings.TrimSpace(strings.ToLower(name))

	cards, err := ListCards()
	if err != nil {
		return nil, err
	}

	for _, c := range cards {
		if strings.Contains(c.Name, name) {
			group = append(group, c)
		}
	}

	if len(group) == 0 {
		return nil, errors.New("no cards were found")
	}

	return group, nil
}

// CreateCard creates a new bank card.
func CreateCard(card *pb.Card) error {
	cards, err := CardsByName(card.Name)
	if err != nil {
		return err
	}

	var exists bool
	for _, c := range cards {
		if strings.Split(c.Name, "/")[0] == card.Name {
			exists = true
			break
		}
	}

	if exists {
		return errors.New("already exists a card or folder with this name")
	}

	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(cardBucket)

		exists := b.Get([]byte(card.Name))
		if exists != nil {
			return errors.Errorf("already exists a card named %s", card.Name)
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

// GetCard retrieves the card with the specified name.
func GetCard(name string) (*pb.Card, error) {
	c := &pb.Card{}

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(cardBucket)
		name = strings.TrimSpace(strings.ToLower(name))

		encCard := b.Get([]byte(name))
		if encCard == nil {
			return errors.Errorf("\"%s\" does not exist", name)
		}

		decCard, err := crypt.Decrypt(encCard)
		if err != nil {
			return errors.Wrap(err, "decrypt card")
		}

		if err := proto.Unmarshal(decCard, c); err != nil {
			return errors.Wrap(err, "unmarshal card")
		}

		if c.Name == "" {
			return errors.Errorf("%s does not exist", name)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return c, nil
}

// ListCards returns a list with all the cards.
func ListCards() ([]*pb.Card, error) {
	var cards []*pb.Card

	password, err := crypt.GetMasterPassword()
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(cardBucket)
		c := b.Cursor()

		// Place cursor in the first line of the bucket and move it to the next one
		for k, v := c.First(); k != nil; k, v = c.Next() {
			card := &pb.Card{}

			decCard, err := crypt.DecryptX(v, password)
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

// RemoveCard removes a card from the database.
func RemoveCard(name string) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(cardBucket)
		name = strings.TrimSpace(strings.ToLower(name))

		_, err := GetCard(name)
		if err != nil {
			return err
		}

		if err := b.Delete([]byte(name)); err != nil {
			return errors.Wrap(err, "remove card")
		}

		return nil
	})
}
