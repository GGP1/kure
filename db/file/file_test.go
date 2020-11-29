package file

import (
	"testing"
	"time"

	"github.com/GGP1/kure/pb"
	"github.com/awnumar/memguard"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	bolt "go.etcd.io/bbolt"
)

func TestFile(t *testing.T) {
	db := setContext(t)
	defer db.Close()

	file := &pb.File{
		Name:      "test",
		Content:   []byte("content"),
		Filename:  "test.txt",
		CreatedAt: 0,
	}
	newName := "newtestname"

	t.Run("Create", create(db, file))
	t.Run("Get", get(db, file))
	t.Run("List", list(db))
	t.Run("List by name", listByName(db, file.Name))
	t.Run("Rename", rename(db, file.Name, newName))
	t.Run("Remove", remove(db, newName))
}

func create(db *bolt.DB, file *pb.File) func(*testing.T) {
	return func(t *testing.T) {
		if err := Create(db, file); err != nil {
			t.Fatalf("Create() failed: %v", err)
		}
	}
}

func get(db *bolt.DB, file *pb.File) func(*testing.T) {
	return func(t *testing.T) {
		got, err := Get(db, file.Name)
		if err != nil {
			t.Fatalf("Get() failed: %v", err)
		}

		// They aren't DeepEqual
		if got.Name != file.Name {
			t.Errorf("Expected %s, got %s", file.Name, got.Name)
		}
	}
}

func list(db *bolt.DB) func(*testing.T) {
	return func(t *testing.T) {
		files, err := List(db)
		if err != nil {
			t.Fatalf("List() failed: %v", err)
		}

		if len(files) == 0 {
			t.Error("Expected one or more files, got 0")
		}
	}
}

func listByName(db *bolt.DB, name string) func(*testing.T) {
	return func(t *testing.T) {
		files, err := ListByName(db, name)
		if err != nil {
			t.Fatalf("ListByName() failed: %v", err)
		}

		if len(files) == 0 {
			t.Error("Expected one or more files, got 0")
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

func rename(db *bolt.DB, oldName, newName string) func(*testing.T) {
	return func(t *testing.T) {
		if err := Rename(db, oldName, newName); err != nil {
			t.Fatalf("Rename() failed: %v", err)
		}
	}
}

func TestCreateErrors(t *testing.T) {
	db := setContext(t)
	defer db.Close()

	file := &pb.File{Name: "test create errors"}
	// Create the file to receive 'already exists' error
	t.Run("Create", create(db, file))

	if err := Create(db, file); err == nil {
		t.Error("Expected 'already exists' error, got nil")
	}

	if err := Create(db, &pb.File{}); err == nil {
		t.Error("Expected 'save entry' error, got nil")
	}

	// Remove the file to not receive 'already exists' error
	viper.Set("user.password", nil)
	if err := Create(db, &pb.File{}); err == nil {
		t.Error("Expected List 'decrypt file' error, got nil")
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

	file := &pb.File{Name: "nil bucket"}

	db.Update(func(tx *bolt.Tx) error {
		tx.DeleteBucket([]byte("kure_file"))
		return nil
	})

	_, err := Get(db, file.Name)
	if err == nil {
		t.Error("Get() didn't return 'invalid bucket'")
	}
	_, err = List(db)
	if err == nil {
		t.Error("List() didn't return 'invalid bucket'")
	}
	if err := Remove(db, file.Name); err == nil {
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
