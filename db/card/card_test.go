package card

import (
	"testing"

	"github.com/GGP1/kure/config"
	"github.com/GGP1/kure/crypt"
	dbutil "github.com/GGP1/kure/db"
	"github.com/GGP1/kure/pb"

	"github.com/awnumar/memguard"
	"github.com/stretchr/testify/assert"
	bolt "go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"
)

func TestCard(t *testing.T) {
	db := setContext(t)

	c := &pb.Card{
		Name:         "test",
		Type:         "debit",
		Number:       "4403650939814064",
		SecurityCode: "1234",
		ExpireDate:   "16/2024",
	}

	// Create destroys the buffer, hence we cannot use their fields anymore
	t.Run("Create", create(db, c))
	t.Run("Get", get(db, c))
	t.Run("List", list(db, c))
	t.Run("List names", listNames(db, c))
	t.Run("Remove", remove(db, c.Name))
	t.Run("Update", update(db))
}

func create(db *bolt.DB, c *pb.Card) func(*testing.T) {
	return func(t *testing.T) {
		err := Create(db, c)
		assert.NoError(t, err)
	}
}

func get(db *bolt.DB, expected *pb.Card) func(*testing.T) {
	return func(t *testing.T) {
		got, err := Get(db, expected.Name)
		assert.NoError(t, err)

		if !proto.Equal(expected, got) {
			t.Errorf("Expected %v, got %v", expected, got)
		}
	}
}

func list(db *bolt.DB, expected *pb.Card) func(*testing.T) {
	return func(t *testing.T) {
		cards, err := List(db)
		assert.NoError(t, err)

		assert.NotZero(t, len(cards), "Expected one or more cards")

		got := cards[0]
		if !proto.Equal(expected, got) {
			t.Errorf("Expected %v, got %v", expected, got)
		}
	}
}

func listNames(db *bolt.DB, expected *pb.Card) func(*testing.T) {
	return func(t *testing.T) {
		cards, err := ListNames(db)
		assert.NoError(t, err)

		if len(cards) == 0 {
			t.Fatal("Expected one or more cards, got 0")
		}

		got := cards[0]
		if got != expected.Name {
			t.Errorf("Expected %s, got %s", expected.Name, got)
		}
	}
}

func remove(db *bolt.DB, name string) func(*testing.T) {
	return func(t *testing.T) {
		err := Remove(db, name)
		assert.NoError(t, err)
	}
}

func update(db *bolt.DB) func(*testing.T) {
	return func(t *testing.T) {
		oldCard := &pb.Card{Name: "old"}
		err := Create(db, oldCard)
		assert.NoError(t, err)

		newCard := &pb.Card{Name: "new"}
		err = Update(db, oldCard.Name, newCard)
		assert.NoError(t, err)

		_, err = Get(db, newCard.Name)
		assert.NoError(t, err)
	}
}

func TestCreateErrors(t *testing.T) {
	db := setContext(t)

	cases := []struct {
		desc string
		name string
	}{
		{
			desc: "Invalid name",
			name: "",
		},
		{
			desc: "Null characters",
			name: string('\x00'),
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			err := Create(db, &pb.Card{Name: tc.name})
			assert.Error(t, err)
		})
	}
}

func TestGetError(t *testing.T) {
	db := setContext(t)

	_, err := Get(db, "non-existent")
	assert.Error(t, err)
}

func TestCryptErrors(t *testing.T) {
	db := setContext(t)

	name := "crypt-errors"
	err := Create(db, &pb.Card{Name: name})
	assert.NoError(t, err)

	// Try to get the card with another password
	config.Set("auth.password", memguard.NewEnclave([]byte("invalid")))

	_, err = Get(db, name)
	assert.Error(t, err)
	_, err = List(db)
	assert.Error(t, err)
}

func TestProtoErrors(t *testing.T) {
	db := setContext(t)

	name := "unformatted"
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(dbutil.CardBucket))
		buf := make([]byte, 64)
		encBuf, _ := crypt.Encrypt(buf)
		return b.Put([]byte(name), encBuf)
	})
	assert.NoError(t, err, "Failed writing invalid type")

	_, err = Get(db, name)
	assert.Error(t, err)
	_, err = List(db)
	assert.Error(t, err)
}

func TestKeyError(t *testing.T) {
	db := setContext(t)

	err := Create(db, &pb.Card{Name: ""})
	assert.Error(t, err)
}

func setContext(t testing.TB) *bolt.DB {
	return dbutil.SetContext(t, "../testdata/database", dbutil.CardBucket)
}
