package card

import (
	"testing"

	"github.com/GGP1/kure/config"
	"github.com/GGP1/kure/crypt"
	dbutils "github.com/GGP1/kure/db"
	"github.com/GGP1/kure/pb"

	"github.com/awnumar/memguard"
	bolt "go.etcd.io/bbolt"
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
	t.Run("Get", get(db, c.Name))
	t.Run("List", list(db))
	t.Run("List names", listNames(db))
	t.Run("Remove", remove(db, c.Name))
	t.Run("Update", update(db))
}

func create(db *bolt.DB, c *pb.Card) func(*testing.T) {
	return func(t *testing.T) {
		if err := Create(db, c); err != nil {
			t.Error(err)
		}
	}
}

func get(db *bolt.DB, name string) func(*testing.T) {
	return func(t *testing.T) {
		got, err := Get(db, name)
		if err != nil {
			t.Error(err)
		}

		// They aren't DeepEqual
		if got.Name != name {
			t.Errorf("Expected %s, got %s", name, got.Name)
		}
	}
}

func list(db *bolt.DB) func(*testing.T) {
	return func(t *testing.T) {
		cards, err := List(db)
		if err != nil {
			t.Error(err)
		}

		if len(cards) == 0 {
			t.Error("Expected one or more cards, got 0")
		}
	}
}

func listNames(db *bolt.DB) func(*testing.T) {
	return func(t *testing.T) {
		cards, err := ListNames(db)
		if err != nil {
			t.Error(err)
		}

		if len(cards) == 0 {
			t.Fatal("Expected one or more cards, got 0")
		}

		expected := "test"
		got := cards[0]

		if got != expected {
			t.Errorf("Expected %s, got %s", expected, got)
		}
	}
}

func remove(db *bolt.DB, name string) func(*testing.T) {
	return func(t *testing.T) {
		if err := Remove(db, name); err != nil {
			t.Error(err)
		}
	}
}

func update(db *bolt.DB) func(*testing.T) {
	return func(t *testing.T) {
		oldCard := &pb.Card{Name: "old"}
		if err := Create(db, oldCard); err != nil {
			t.Fatal(err)
		}

		newCard := &pb.Card{Name: "new"}
		if err := Update(db, oldCard.Name, newCard); err != nil {
			t.Fatal(err)
		}

		if _, err := Get(db, newCard.Name); err != nil {
			t.Error(err)
		}
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
			if err := Create(db, &pb.Card{Name: tc.name}); err == nil {
				t.Error("Expected an error and got nil")
			}
		})
	}
}

func TestGetError(t *testing.T) {
	db := setContext(t)

	if _, err := Get(db, "non-existent"); err == nil {
		t.Error("Expected 'does not exist' error, got nil")
	}
}

func TestCryptErrors(t *testing.T) {
	db := setContext(t)

	name := "crypt-errors"
	if err := Create(db, &pb.Card{Name: name}); err != nil {
		t.Fatal(err)
	}

	// Try to get the card with another password
	config.Set("auth.password", memguard.NewEnclave([]byte("invalid")))

	if _, err := Get(db, name); err == nil {
		t.Error("Expected Get() to fail but it didn't")
	}
	if _, err := List(db); err == nil {
		t.Error("Expected List() to fail but it didn't")
	}
}

func TestProtoErrors(t *testing.T) {
	db := setContext(t)

	name := "unformatted"
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(cardBucket))
		buf := make([]byte, 64)
		encBuf, _ := crypt.Encrypt(buf)
		return b.Put([]byte(name), encBuf)
	})
	if err != nil {
		t.Fatalf("Failed writing invalid type: %v", err)
	}

	if _, err := Get(db, name); err == nil {
		t.Error("Expected Get() to fail but it didn't")
	}
	if _, err := List(db); err == nil {
		t.Error("Expected List() to fail but it didn't")
	}
}

func TestKeyError(t *testing.T) {
	db := setContext(t)

	if err := Create(db, &pb.Card{Name: ""}); err == nil {
		t.Error("Create() didn't fail")
	}
}

func setContext(t testing.TB) *bolt.DB {
	return dbutils.SetContext(t, "../testdata/database", cardBucket)
}
