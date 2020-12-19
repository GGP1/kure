package card

import (
	"testing"
	"time"

	"github.com/GGP1/kure/pb"
	"github.com/awnumar/memguard"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	bolt "go.etcd.io/bbolt"
)

func TestCard(t *testing.T) {
	db := setContext(t)
	defer db.Close()

	c := &pb.Card{
		Name:         "test",
		Type:         "debit",
		Number:       "4403650939814064",
		SecurityCode: "1234",
		ExpireDate:   "16/2024",
	}

	t.Run("Create", create(db, c))
	t.Run("Get", get(db, c))
	t.Run("List", list(db))
	t.Run("List fastest", listFastest(db))
	t.Run("List names", listNames(db))
	t.Run("Remove", remove(db, c.Name))
}

func create(db *bolt.DB, card *pb.Card) func(*testing.T) {
	return func(t *testing.T) {
		if err := Create(db, card); err != nil {
			t.Fatalf("Create() failed: %v", err)
		}
	}
}

func get(db *bolt.DB, card *pb.Card) func(*testing.T) {
	return func(t *testing.T) {
		got, err := Get(db, card.Name)
		if err != nil {
			t.Fatalf("Get() failed: %v", err)
		}

		// They aren't DeepEqual
		if got.Name != card.Name {
			t.Errorf("Expected %s, got %s", card.Name, got.Name)
		}
	}
}

func list(db *bolt.DB) func(*testing.T) {
	return func(t *testing.T) {
		cards, err := List(db)
		if err != nil {
			t.Fatalf("List() failed: %v", err)
		}

		if len(cards) == 0 {
			t.Error("Expected one or more cards, got 0")
		}
	}
}

func listFastest(db *bolt.DB) func(*testing.T) {
	return func(t *testing.T) {
		if !ListFastest(db) {
			t.Error("Failed decrypting cards")
		}
	}
}

func listNames(db *bolt.DB) func(*testing.T) {
	return func(t *testing.T) {
		cards, err := ListNames(db)
		if err != nil {
			t.Fatalf("List() failed: %v", err)
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
			t.Fatalf("Remove() failed: %v", err)
		}
	}
}

func TestCreateErrors(t *testing.T) {
	db := setContext(t)
	defer db.Close()

	card := &pb.Card{Name: "test create errors"}
	// Create the card to receive 'already exists' error
	t.Run("Create", create(db, card))

	if err := Create(db, &pb.Card{}); err == nil {
		t.Error("Expected 'save card' error, got nil")
	}

	viper.Set("user.password", nil)
	if err := Create(db, &pb.Card{}); err == nil {
		t.Error("Expected List 'decrypt card' error, got nil")
	}
}

func TestGetError(t *testing.T) {
	db := setContext(t)
	defer db.Close()

	_, err := Get(db, "non-existent")
	if err == nil {
		t.Error("Expected 'does not exist' error, got nil")
	}
}

func TestBucketError(t *testing.T) {
	db := setContext(t)
	defer db.Close()

	card := &pb.Card{Name: "nil bucket"}

	db.Update(func(tx *bolt.Tx) error {
		tx.DeleteBucket([]byte("kure_card"))
		return nil
	})

	if err := Create(db, card); err == nil {
		t.Error("Create() didn't return 'invalid bucket' error")
	}
	_, err := Get(db, card.Name)
	if err == nil {
		t.Error("Get() didn't return 'invalid bucket' error")
	}
	_, err = List(db)
	if err == nil {
		t.Error("List() didn't return 'invalid bucket' error")
	}
	_, err = ListNames(db)
	if err == nil {
		t.Error("ListNames() didn't return 'invalid bucket' error")
	}
	if err := Remove(db, card.Name); err == nil {
		t.Error("Remove() didn't return 'invalid bucket' error")
	}
}

func TestDecryptError(t *testing.T) {
	db := setContext(t)
	defer db.Close()

	card := &pb.Card{Name: "test decrypt error"}
	if err := Create(db, card); err != nil {
		t.Fatal(err)
	}

	viper.Set("user.password", nil)

	_, err := Get(db, card.Name)
	if err == nil {
		t.Error("Get() didn't return 'decrypt card' error")
	}
	_, err = List(db)
	if err == nil {
		t.Error("List() didn't return 'decrypt card' error")
	}
	if ListFastest(db) {
		t.Error("Expected ListFastest() to return false and returned true")
	}
}

func TestKeyError(t *testing.T) {
	db := setContext(t)
	defer db.Close()

	card := &pb.Card{Name: ""}

	if err := Create(db, card); err == nil {
		t.Error("Create() didn't fail")
	}
}

func setContext(t *testing.T) *bolt.DB {
	db, err := bolt.Open("../testdata/database", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		t.Fatalf("Failed connecting to the database: %v", err)
	}

	viper.Reset()
	password := memguard.NewBufferFromBytes([]byte("test"))
	defer password.Destroy()
	viper.Set("user.password", password.Seal())

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

	return db
}
