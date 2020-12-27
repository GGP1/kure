package note

import (
	"testing"
	"time"

	"github.com/GGP1/kure/pb"
	"github.com/awnumar/memguard"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	bolt "go.etcd.io/bbolt"
)

func TestNote(t *testing.T) {
	db := setContext(t)
	defer db.Close()

	name := "test"
	text := "text"

	buf, n := pb.SecureNote()
	n.Name = name
	n.Text = text

	// Create destroys the buffer, hence we cannot use their fields anymore
	t.Run("Create", create(db, buf, n))
	t.Run("Get", get(db, name, text))
	t.Run("List", list(db))
	t.Run("List fastest", listFastest(db))
	t.Run("List names", listNames(db))
	t.Run("Remove", remove(db, name))
}

func create(db *bolt.DB, buf *memguard.LockedBuffer, note *pb.Note) func(*testing.T) {
	return func(t *testing.T) {
		if err := Create(db, buf, note); err != nil {
			t.Fatalf("Create() failed: %v", err)
		}
	}
}

func get(db *bolt.DB, name, text string) func(*testing.T) {
	return func(t *testing.T) {
		_, got, err := Get(db, name)
		if err != nil {
			t.Fatalf("Get() failed: %v", err)
		}

		if got.Name != name {
			t.Errorf("Expected %s, got %s", name, got.Name)
		}

		if got.Text != text {
			t.Errorf("Expected %s, got %s", text, got.Text)
		}
	}
}

func list(db *bolt.DB) func(*testing.T) {
	return func(t *testing.T) {
		_, notes, err := List(db)
		if err != nil {
			t.Fatalf("List() failed: %v", err)
		}

		if len(notes) == 0 {
			t.Error("Expected one or more notes, got 0")
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
		notes, err := ListNames(db)
		if err != nil {
			t.Fatalf("List() failed: %v", err)
		}

		if len(notes) == 0 {
			t.Fatal("Expected one or more notes, got 0")
		}

		expected := "test"
		got := notes[0]

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

	lockedBuf, n := pb.SecureNote()
	if err := Create(db, lockedBuf, n); err == nil {
		t.Error("Expected 'save note' error, got nil")
	}
}

func TestGetError(t *testing.T) {
	db := setContext(t)
	defer db.Close()

	_, _, err := Get(db, "non-existent")
	if err == nil {
		t.Error("Expected 'does not exist' error, got nil")
	}
}

func TestBucketError(t *testing.T) {
	db := setContext(t)
	defer db.Close()

	name := "nil bucket"
	buf, note := pb.SecureNote()
	note.Name = name

	db.Update(func(tx *bolt.Tx) error {
		tx.DeleteBucket([]byte("kure_note"))
		return nil
	})

	if err := Create(db, buf, note); err == nil {
		t.Error("Create() didn't return 'invalid bucket' error")
	}
	_, _, err := Get(db, name)
	if err == nil {
		t.Error("Get() didn't return 'invalid bucket' error")
	}
	_, _, err = List(db)
	if err == nil {
		t.Error("List() didn't return 'invalid bucket' error")
	}
	_, err = ListNames(db)
	if err == nil {
		t.Error("ListNames() didn't return 'invalid bucket' error")
	}
	if err := Remove(db, name); err == nil {
		t.Error("Remove() didn't return 'invalid bucket' error")
	}
}

func TestDecryptError(t *testing.T) {
	db := setContext(t)
	defer db.Close()

	name := "test decrypt error"
	buf, note := pb.SecureNote()
	note.Name = name

	if err := Create(db, buf, note); err != nil {
		t.Fatal(err)
	}

	viper.Set("user.password", nil)

	_, _, err := Get(db, name)
	if err == nil {
		t.Error("Get() didn't return 'decrypt note' error")
	}
	_, _, err = List(db)
	if err == nil {
		t.Error("List() didn't return 'decrypt note' error")
	}
	if ListFastest(db) {
		t.Error("Expected ListFastest() to return false and returned true")
	}
}

func TestKeyError(t *testing.T) {
	db := setContext(t)
	defer db.Close()

	buf, note := pb.SecureNote()
	note.Name = ""

	if err := Create(db, buf, note); err == nil {
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
		bucket := "kure_note"
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
