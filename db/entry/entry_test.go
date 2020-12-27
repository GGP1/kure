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

	name := "test"
	lockedBuf, e := pb.SecureEntry()
	e.Name = name
	e.Username = "testing"
	e.URL = "golang.org"
	e.Expires = "Never"
	e.Notes = ""

	t.Run("Create", create(db, lockedBuf, e))
	t.Run("Get", get(db, name))
	t.Run("List", list(db))
	t.Run("List fastest", listFastest(db))
	t.Run("List names", listNames(db))
	t.Run("Remove", remove(db, name))
}

func create(db *bolt.DB, buf *memguard.LockedBuffer, entry *pb.Entry) func(*testing.T) {
	return func(t *testing.T) {
		if err := Create(db, buf, entry); err != nil {
			t.Fatalf("Create() failed: %v", err)
		}
	}
}

func get(db *bolt.DB, name string) func(*testing.T) {
	return func(t *testing.T) {
		_, got, err := Get(db, name)
		if err != nil {
			t.Fatalf("Get() failed: %v", err)
		}

		// They aren't DeepEqual
		if got.Name != name {
			t.Errorf("Expected %s, got %s", name, got.Name)
		}
	}
}

func list(db *bolt.DB) func(*testing.T) {
	return func(t *testing.T) {
		_, entries, err := List(db)
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

	lockedBuf, e := pb.SecureEntry()
	e.Name = "test expired"
	e.Expires = "Mon, 10 Jan 2020 15:04:05 -0700"

	t.Run("Create", create(db, lockedBuf, e))

	_, _, err := List(db)
	if err != nil {
		t.Errorf("List() failed: %v", err)
	}

	// Recreate as it was deleted by List()
	lockedBuf2, e2 := pb.SecureEntry()
	e2.Name = "test expired"
	e2.Expires = "Mon, 10 Jan 2020 15:04:05 -0700"

	t.Run("Create", create(db, lockedBuf2, e2))

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

	lockedBuf, e := pb.SecureEntry()
	viper.Set("user.password", nil)
	if err := Create(db, lockedBuf, e); err == nil {
		t.Error("Expected 'encrypt entry' error, got nil")
	}
}

func TestGetErrors(t *testing.T) {
	db := setContext(t)
	defer db.Close()

	lockedBuf, e := pb.SecureEntry()
	e.Name = "test expired"
	e.Expires = "Mon, 10 Jan 2020 15:04:05 -0700"

	t.Run("Create", create(db, lockedBuf, e))

	names := []string{"test expired", "non-existent"}

	for _, name := range names {
		_, _, err := Get(db, name)
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

	name := "nil bucket"
	lockedBuf, e := pb.SecureEntry()
	e.Name = name
	e.Expires = "Never"

	err := db.Update(func(tx *bolt.Tx) error {
		tx.DeleteBucket([]byte("kure_entry"))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := Create(db, lockedBuf, e); err == nil {
		t.Error("Edit() didn't return 'invalid bucket'")
	}
	_, _, err = Get(db, name)
	if err == nil {
		t.Error("Get() didn't return 'invalid bucket'")
	}
	_, _, err = List(db)
	if err == nil {
		t.Error("List() didn't return 'invalid bucket'")
	}
	_, err = ListNames(db)
	if err == nil {
		t.Error("ListAllNames() didn't return 'invalid bucket'")
	}
	if err := Remove(db, name); err == nil {
		t.Error("Remove() didn't return 'invalid bucket'")
	}
}

func TestDecryptError(t *testing.T) {
	db := setContext(t)
	defer db.Close()

	name := "test decrypt error"
	lockedBuf, e := pb.SecureEntry()
	e.Name = name
	e.Expires = "Never"

	if err := Create(db, lockedBuf, e); err != nil {
		t.Fatal(err)
	}

	viper.Set("user.password", nil)

	_, _, err := Get(db, name)
	if err == nil {
		t.Error("Get() didn't return 'decrypt entry' error")
	}
	_, _, err = List(db)
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

	lockedBuf, e := pb.SecureEntry()
	e.Name = ""

	if err := Create(db, lockedBuf, e); err == nil {
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
