package file

import (
	"bytes"
	"compress/gzip"
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

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)

	_, err := gw.Write([]byte("content"))
	if err != nil {
		t.Fatalf("Failed compressing content")
	}

	if err := gw.Close(); err != nil {
		t.Fatalf("Failed closing gzip writer")
	}

	file := &pb.File{
		Name:      "test",
		Content:   buf.Bytes(),
		Filename:  "test.txt",
		CreatedAt: 0,
	}

	// Restore is tested separated
	t.Run("Create", create(db, file))
	t.Run("Get", get(db, file))
	t.Run("List", list(db))
	t.Run("List fastest", listFastest(db))
	t.Run("List names", listNames(db))
	t.Run("Rename", rename(db, file.Name, "newtestname"))
	t.Run("Remove", remove(db, "newtestname"))
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

func listFastest(db *bolt.DB) func(*testing.T) {
	return func(t *testing.T) {
		if !ListFastest(db) {
			t.Error("Failed decrypting files")
		}
	}
}

func listNames(db *bolt.DB) func(*testing.T) {
	return func(t *testing.T) {
		files, err := ListNames(db)
		if err != nil {
			t.Fatalf("List() failed: %v", err)
		}

		if len(files) == 0 {
			t.Fatal("Expected one or more files, got 0")
		}

		expected := "test"
		got := files[0]

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

func rename(db *bolt.DB, oldName, newName string) func(*testing.T) {
	return func(t *testing.T) {
		if err := Rename(db, oldName, newName); err != nil {
			t.Fatalf("Rename() failed: %v", err)
		}
	}
}

func TestRestore(t *testing.T) {
	db := setContext(t)
	defer db.Close()

	file := &pb.File{
		Name:      "test",
		Content:   []byte("Minas tirith"),
		Filename:  "test.txt",
		CreatedAt: 0,
	}

	if err := Restore(db, file); err != nil {
		t.Errorf("Restore() failed: %v", err)
	}
}

func TestCreateErrors(t *testing.T) {
	db := setContext(t)
	defer db.Close()

	file := &pb.File{Name: "test create errors"}
	// Create the file to receive 'already exists' error
	t.Run("Create", create(db, file))

	if err := Create(db, &pb.File{}); err == nil {
		t.Error("Expected 'save file' error, got nil")
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

func TestRenameError(t *testing.T) {
	db := setContext(t)
	defer db.Close()

	// Create file to force "New name already used" error
	if err := Create(db, &pb.File{Name: "test rename"}); err != nil {
		t.Fatalf("Failed creating file: %v", err)
	}

	cases := []struct {
		desc    string
		oldName string
		newName string
	}{
		{
			desc:    "File does not exists",
			oldName: "non-existent",
		},
		{
			desc:    "New name already used",
			oldName: "test rename",
			newName: "test rename",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			if err := Rename(db, tc.oldName, tc.newName); err == nil {
				t.Error("Expected Rename() to fail but it didn't")
			}
		})
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
	_, err = ListNames(db)
	if err == nil {
		t.Error("ListNames() didn't return 'invalid bucket'")
	}
	if err := Remove(db, file.Name); err == nil {
		t.Error("Remove() didn't return 'invalid bucket'")
	}
	if err := Restore(db, file); err == nil {
		t.Error("Restore() didn't return 'invalid bucket'")
	}
}

func TestDecryptError(t *testing.T) {
	db := setContext(t)
	defer db.Close()

	file := &pb.File{Name: "test decrypt error"}
	if err := Create(db, file); err != nil {
		t.Fatal(err)
	}

	viper.Set("user.password", nil)

	_, err := Get(db, file.Name)
	if err == nil {
		t.Error("Get() didn't return 'decrypt file' error")
	}
	_, err = List(db)
	if err == nil {
		t.Error("List() didn't return 'decrypt file' error")
	}
	if ListFastest(db) {
		t.Error("Expected ListFastest() to return false and returned true")
	}
}

func TestKeyError(t *testing.T) {
	db := setContext(t)
	defer db.Close()

	file := &pb.File{Name: ""}

	if err := Create(db, file); err == nil {
		t.Error("Create() didn't fail")
	}
	if err := Restore(db, file); err == nil {
		t.Error("Restore() didn't fail")
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
		bucket := "kure_file"
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
