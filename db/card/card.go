package card

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
	cardBucket       = []byte("kure_card")
	errInvalidBucket = errors.New("invalid bucket")
)

// Create a new bank card. It destroys the locked buffer passed.
func Create(db *bolt.DB, lockedBuf *memguard.LockedBuffer, card *pb.Card) error {
	return db.Update(func(tx *bolt.Tx) error {
		defer lockedBuf.Destroy()

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
func Get(db *bolt.DB, name string) (*memguard.LockedBuffer, *pb.Card, error) {
	buf, c := pb.SecureCard()

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
		return nil, nil, err
	}

	return buf, c, nil
}

// List returns a list with all the cards.
func List(db *bolt.DB) (*memguard.LockedBuffer, []*pb.Card, error) {
	cardsBuf, cards := pb.SecureCardSlice()

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(cardBucket)
		if b == nil {
			return errInvalidBucket
		}

		return b.ForEach(func(k, v []byte) error {
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
	})
	if err != nil {
		return nil, nil, err
	}

	return cardsBuf, cards, nil
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
		b := tx.Bucket(cardBucket)

		return b.ForEach(func(_, v []byte) error {
			go decrypt(v)
			return nil
		})
	})

	return <-succeed
}

// ListNames returns a list with all the cards names.
func ListNames(db *bolt.DB) ([]string, error) {
	var cards []string

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(cardBucket)
		if b == nil {
			return errInvalidBucket
		}

		return b.ForEach(func(k, _ []byte) error {
			cards = append(cards, string(k))
			return nil
		})
	})
	if err != nil {
		return nil, err
	}

	return cards, nil
}

// Remove removes a card from the database.
func Remove(db *bolt.DB, name string) error {
	return db.Update(func(tx *bolt.Tx) error {
		name = strings.TrimSpace(strings.ToLower(name))

		b := tx.Bucket(cardBucket)
		if b == nil {
			return errInvalidBucket
		}

		if err := b.Delete([]byte(name)); err != nil {
			return errors.Wrap(err, "remove card")
		}

		return nil
	})
}
