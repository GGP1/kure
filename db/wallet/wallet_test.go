package wallet

import (
	"testing"
	"time"

	"github.com/GGP1/kure/pb"
	"github.com/awnumar/memguard"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	bolt "go.etcd.io/bbolt"
)

func TestWallet(t *testing.T) {
	db := setContext(t)
	defer db.Close()

	wallet := &pb.Wallet{
		Name:       "test",
		Type:       "Ethereum",
		PrivateKey: "public",
		PublicKey:  "private",
	}

	t.Run("Create", create(db, wallet))
	t.Run("Get", get(db, wallet))
	t.Run("List", list(db))
	t.Run("List by name", listByName(db, wallet.Name))
	t.Run("List names", listNames(db))
	t.Run("Remove", remove(db, wallet.Name))
}

func create(db *bolt.DB, wallet *pb.Wallet) func(*testing.T) {
	return func(t *testing.T) {
		if err := Create(db, wallet); err != nil {
			t.Fatalf("Create() failed: %v", err)
		}
	}
}

func get(db *bolt.DB, wallet *pb.Wallet) func(*testing.T) {
	return func(t *testing.T) {
		got, err := Get(db, wallet.Name)
		if err != nil {
			t.Fatalf("Get() failed: %v", err)
		}

		// They aren't DeepEqual
		if got.Name != wallet.Name {
			t.Errorf("Expected %s, got %s", wallet.Name, got.Name)
		}
	}
}

func list(db *bolt.DB) func(*testing.T) {
	return func(t *testing.T) {
		wallets, err := List(db)
		if err != nil {
			t.Fatalf("List() failed: %v", err)
		}

		if len(wallets) == 0 {
			t.Error("Expected one or more wallets, got 0")
		}
	}
}

func listByName(db *bolt.DB, name string) func(*testing.T) {
	return func(t *testing.T) {
		wallets, err := ListByName(db, name)
		if err != nil {
			t.Fatalf("ListByName() failed: %v", err)
		}

		if len(wallets) == 0 {
			t.Error("Expected one or more wallets, got 0")
		}
	}
}

func listNames(db *bolt.DB) func(*testing.T) {
	return func(t *testing.T) {
		wallets, err := ListNames(db)
		if err != nil {
			t.Fatalf("List() failed: %v", err)
		}

		expected := "test"
		got := wallets[0].Name

		if got != expected {
			t.Errorf("Expected %s, got %s", expected, got)
		}

		if len(wallets) == 0 {
			t.Error("Expected one or more wallets, got 0")
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

	wallet := &pb.Wallet{Name: "test create errors"}
	// Create the wallet to receive 'already exists' error
	t.Run("Create", create(db, wallet))

	if err := Create(db, &pb.Wallet{}); err == nil {
		t.Error("Expected 'save entry' error, got nil")
	}

	// Remove the wallet to not receive 'already exists' error
	viper.Set("user.password", nil)
	if err := Create(db, &pb.Wallet{}); err == nil {
		t.Error("Expected List 'decrypt wallet' error, got nil")
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

	wallet := &pb.Wallet{Name: "nil bucket"}

	db.Update(func(tx *bolt.Tx) error {
		tx.DeleteBucket([]byte("kure_wallet"))
		return nil
	})

	_, err := Get(db, wallet.Name)
	if err == nil {
		t.Error("Expected Get() to return an error but it didn't")
	}
	_, err = List(db)
	if err == nil {
		t.Error("Expected List() to return an error but it didn't")
	}
	_, err = ListByName(db, wallet.Name)
	if err == nil {
		t.Error("Expected ListByName() to return an error but it didn't")
	}
	if err := Remove(db, wallet.Name); err == nil {
		t.Error("Expected Remove() to return an error but it didn't")
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
