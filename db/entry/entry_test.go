package entry

import (
	"testing"
	"time"

	"github.com/GGP1/kure/pb"

	"github.com/awnumar/memguard"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	bolt "go.etcd.io/bbolt"
)

func TestEntry(t *testing.T) {
	db := setContext(t)
	defer db.Close()

	entry := &pb.Entry{
		Name:     "test",
		Username: "testing",
		URL:      "golang.org",
		Notes:    "",
		Expires:  "Never",
	}

	t.Run("Create", create(db, entry))
	t.Run("Edit", edit(db, entry.Name))
	t.Run("Get", get(db, entry))
	t.Run("List", list(db))
	t.Run("List by name", listByName(db, entry.Name))
	t.Run("List names", listNames(db))
	t.Run("Remove", remove(db, entry.Name))
}

func create(db *bolt.DB, entry *pb.Entry) func(*testing.T) {
	return func(t *testing.T) {
		if err := Create(db, entry); err != nil {
			t.Fatalf("Create() failed: %v", err)
		}
	}
}

func edit(db *bolt.DB, name string) func(*testing.T) {
	return func(t *testing.T) {
		entry := &pb.Entry{
			Name:     "test",
			Username: "edit username",
			URL:      "golang.org",
			Notes:    "",
			Expires:  "Never",
		}

		if err := Edit(db, name, entry); err != nil {
			t.Fatalf("Edit() failed: %v", err)
		}
	}
}

func get(db *bolt.DB, entry *pb.Entry) func(*testing.T) {
	return func(t *testing.T) {
		got, err := Get(db, entry.Name)
		if err != nil {
			t.Fatalf("Get() failed: %v", err)
		}

		// They aren't DeepEqual
		if got.Name != entry.Name {
			t.Errorf("Expected %s, got %s", entry.Name, got.Name)
		}
	}
}

func list(db *bolt.DB) func(*testing.T) {
	return func(t *testing.T) {
		entries, err := List(db)
		if err != nil {
			t.Fatalf("List() failed: %v", err)
		}

		if len(entries) == 0 {
			t.Error("Expected one or more entries, got 0")
		}
	}
}

func listByName(db *bolt.DB, name string) func(*testing.T) {
	return func(t *testing.T) {
		entries, err := ListByName(db, name)
		if err != nil {
			t.Fatalf("ListByName() failed: %v", err)
		}

		if len(entries) == 0 {
			t.Error("Expected one or more entries, got 0")
		}
	}
}

func listNames(db *bolt.DB) func(*testing.T) {
	return func(t *testing.T) {
		entries, err := ListNames(db)
		if err != nil {
			t.Fatalf("List() failed: %v", err)
		}

		expected := "test"
		got := entries[0].Name

		if got != expected {
			t.Errorf("Expected %s, got %s", expected, got)
		}

		if len(entries) == 0 {
			t.Error("Expected one or more entries, got 0")
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

func TestLsExpired(t *testing.T) {
	db := setContext(t)
	defer db.Close()

	entry := &pb.Entry{
		Name:    "test expired",
		Expires: "Mon, 10 Jan 2020 15:04:05 -0700",
	}

	t.Run("Create", create(db, entry))

	_, err := List(db)
	if err != nil {
		t.Errorf("List() failed: %v", err)
	}
}

func TestCreateErrors(t *testing.T) {
	db := setContext(t)
	defer db.Close()

	entry := &pb.Entry{Name: "test create errors"}
	// Create the entry to receive 'already exists' error
	t.Run("Create", create(db, entry))

	if err := Create(db, &pb.Entry{}); err == nil {
		t.Error("Expected 'save entry' error, got nil")
	}

	// Remove the entry to not receive 'already exists' error
	viper.Set("user.password", nil)
	if err := Create(db, &pb.Entry{}); err == nil {
		t.Error("Expected List 'decrypt entry' error, got nil")
	}
}

func TestGetErrors(t *testing.T) {
	db := setContext(t)
	defer db.Close()

	cases := []*pb.Entry{
		{
			Name:    "test expired",
			Expires: "Mon, 10 Jan 2020 15:04:05 -0700",
		},
		{
			Name: "non-existent",
		},
	}

	for _, tc := range cases {
		t.Run("Create", create(db, tc))

		_, err := Get(db, tc.Name)
		if err == nil {
			t.Error("Expected an error, got nil")
		}
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

	entry := &pb.Entry{Name: "nil bucket"}

	err := db.Update(func(tx *bolt.Tx) error {
		tx.DeleteBucket([]byte("kure_entry"))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := Edit(db, entry.Name, entry); err == nil {
		t.Error("Edit() didn't return 'invalid bucket'")
	}
	_, err = Get(db, entry.Name)
	if err == nil {
		t.Error("Get() didn't return 'invalid bucket'")
	}
	_, err = List(db)
	if err == nil {
		t.Error("List() didn't return 'invalid bucket'")
	}
	_, err = ListByName(db, entry.Name)
	if err == nil {
		t.Error("ListByName() didn't return 'invalid bucket'")
	}
	_, err = ListNames(db)
	if err == nil {
		t.Error("ListAllNames() didn't return 'invalid bucket'")
	}
	if err := Remove(db, entry.Name); err == nil {
		t.Error("Remove() didn't return 'invalid bucket'")
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
