package card

import (
	"testing"
	"time"

	"github.com/GGP1/kure/config"
	"github.com/GGP1/kure/crypt"
	"github.com/GGP1/kure/pb"

	"github.com/awnumar/memguard"
	"github.com/pkg/errors"
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

func TestListNameNil(t *testing.T) {
	db := setContext(t)
	err := db.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket(cardBucket)
	})
	if err != nil {
		t.Fatalf("Failed deleting the card bucket: %v", err)
	}

	list, err := ListNames(db)
	if err != nil || list != nil {
		t.Errorf("Expected to receive a nil list and error, got: %v list, %v error", list, err)
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

func setContext(t *testing.T) *bolt.DB {
	db, err := bolt.Open("../testdata/database", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		t.Fatalf("Failed connecting to the database: %v", err)
	}

	config.Reset()
	// Reduce argon2 parameters to speed up tests
	auth := map[string]interface{}{
		"password":   memguard.NewEnclave([]byte("1")),
		"iterations": 1,
		"memory":     1,
		"threads":    1,
	}
	config.Set("auth", auth)

	err = db.Update(func(tx *bolt.Tx) error {
		bucket := "kure_card"
		tx.DeleteBucket([]byte(bucket))
		if _, err := tx.CreateBucketIfNotExists([]byte(bucket)); err != nil {
			return errors.Wrapf(err, "couldn't create %q bucket", bucket)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("Failed closing the database: %v", err)
		}
	})

	return db
}
