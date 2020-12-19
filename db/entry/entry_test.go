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
	t.Run("List fastest", listFastest(db))
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

func listFastest(db *bolt.DB) func(*testing.T) {
	return func(t *testing.T) {
		if !ListFastest(db) {
			t.Error("Failed decrypting entries")
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
		got := entries[0]

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

func TestListExpired(t *testing.T) {
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

	// Recreate as it was deleted by List()
	t.Run("Create", create(db, entry))
	_, err = ListNames(db)
	if err != nil {
		t.Errorf("List() failed: %v", err)
	}
}

func TestIsExpired(t *testing.T) {
	expired := make(chan bool, 1)
	errCh := make(chan error, 1)
	isExpired("Mon, 02 Jan 2030 15:04:05 -0700", expired, errCh)

	select {
	case e := <-expired:
		if e {
			t.Error("Expected non expired result")
		}

	case err := <-errCh:
		if err != nil {
			t.Errorf("isExpired() failed: %v", err)
		}
	}
}

func TestCreateErrors(t *testing.T) {
	db := setContext(t)
	defer db.Close()

	// Remove the entry to not receive 'already exists' error
	viper.Set("user.password", nil)
	if err := Create(db, &pb.Entry{}); err == nil {
		t.Error("Expected 'encrypt entry' error, got nil")
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

	// Create only the first entry
	t.Run("Create", create(db, cases[0]))

	for _, tc := range cases {
		_, err := Get(db, tc.Name)
		if err == nil {
			t.Error("Expected an error, got nil")
		}
	}
}

func TestIsExpiredError(t *testing.T) {
	expired := make(chan bool, 1)
	errCh := make(chan error, 1)
	isExpired("invalid format", expired, errCh)

	select {
	case e := <-expired:
		if e {
			t.Error("Expected the expired channel to return an error")
		}

	case <-errCh:
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

	if err := Create(db, entry); err == nil {
		t.Error("Edit() didn't return 'invalid bucket'")
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
	_, err = ListNames(db)
	if err == nil {
		t.Error("ListAllNames() didn't return 'invalid bucket'")
	}
	if err := Remove(db, entry.Name); err == nil {
		t.Error("Remove() didn't return 'invalid bucket'")
	}
}

func TestDecryptError(t *testing.T) {
	db := setContext(t)
	defer db.Close()

	entry := &pb.Entry{Name: "test decrypt error"}
	if err := Create(db, entry); err != nil {
		t.Fatal(err)
	}

	viper.Set("user.password", nil)

	_, err := Get(db, entry.Name)
	if err == nil {
		t.Error("Get() didn't return 'decrypt entry' error")
	}
	_, err = List(db)
	if err == nil {
		t.Error("List() didn't return 'decrypt entry' error")
	}
	if ListFastest(db) {
		t.Error("Expected ListFastest() to return false and returned true")
	}
}

func TestKeyError(t *testing.T) {
	db := setContext(t)
	defer db.Close()

	entry := &pb.Entry{Name: ""}

	if err := Create(db, entry); err == nil {
		t.Error("Create() didn't fail")
	}
	if err := Edit(db, entry.Name, entry); err == nil {
		t.Error("Edit() didn't fail")
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
		bucket := "kure_entry"
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
