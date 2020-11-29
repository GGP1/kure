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
		Name:       "test",
		Type:       "debit",
		Number:     "4403650939814064",
		CVC:        "1234",
		ExpireDate: "16/2024",
	}

	t.Run("Create", create(db, c))
	t.Run("Get", get(db, c))
	t.Run("List", list(db))
	t.Run("List by name", listByName(db, c.Name))
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

func listByName(db *bolt.DB, name string) func(*testing.T) {
	return func(t *testing.T) {
		cards, err := ListByName(db, name)
		if err != nil {
			t.Fatalf("ListByName() failed: %v", err)
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
			t.Fatalf("List() failed: %v", err)
		}

		expected := "test"
		got := cards[0].Name

		if got != expected {
			t.Errorf("Expected %s, got %s", expected, got)
		}

		if len(cards) == 0 {
			t.Error("Expected one or more cards, got 0")
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
		t.Error("Expected 'save entry' error, got nil")
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

func TestRemoveError(t *testing.T) {
	db := setContext(t)
	defer db.Close()

	if err := Remove(db, "non-existent"); err == nil {
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

	_, err := Get(db, card.Name)
	if err == nil {
		t.Error("Get() failed")
	}
	_, err = List(db)
	if err == nil {
		t.Error("List() failed")
	}
	_, err = ListByName(db, card.Name)
	if err == nil {
		t.Error("ListByName() failed")
	}
	if err := Remove(db, card.Name); err == nil {
		t.Error("Remove() failed")
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
		buckets := [4]string{"kure_card", "kure_entry", "kure_file", "kure_wallet"}
		for _, bucket := range buckets {
			tx.DeleteBucket([]byte(bucket))
			_, err := tx.CreateBucketIfNotExists([]byte(bucket))
			if err != nil {
				return errors.Wrapf(err, "couldn't create %q bucket", bucket)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	return db
}
