package file

import (
	"bytes"
	"compress/gzip"
	"testing"

	"github.com/GGP1/kure/config"
	"github.com/GGP1/kure/crypt"
	dbutil "github.com/GGP1/kure/db"
	"github.com/GGP1/kure/db/bucket"
	"github.com/GGP1/kure/pb"

	"github.com/awnumar/memguard"
	"github.com/stretchr/testify/assert"
	bolt "go.etcd.io/bbolt"
)

func TestFile(t *testing.T) {
	db := setContext(t)

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)

	_, err := gw.Write([]byte("content"))
	assert.NoError(t, err, "Failed compressing content")

	err = gw.Close()
	assert.NoError(t, err, "Failed closing gzip writer")

	f := &pb.File{
		Name:      "test",
		Content:   buf.Bytes(),
		CreatedAt: 0,
	}
	updatedName := "tested"

	t.Run("Create", create(db, f))
	t.Run("Get", get(db, f))
	t.Run("Rename", rename(db, f.Name, updatedName))
	t.Run("Get cheap", getCheap(db, updatedName))
	f.Name = updatedName
	t.Run("List", list(db))
	t.Run("List names", listNames(db, updatedName))
	t.Run("Remove", remove(db, updatedName))
}

func create(db *bolt.DB, f *pb.File) func(*testing.T) {
	return func(t *testing.T) {
		err := Create(db, f)
		assert.NoError(t, err)
	}
}

func get(db *bolt.DB, expected *pb.File) func(*testing.T) {
	return func(t *testing.T) {
		got, err := Get(db, expected.Name)
		assert.NoError(t, err)

		// Using proto.Equal fails because of differing content buffers
		assert.Equal(t, expected.Name, got.Name)
	}
}

func rename(db *bolt.DB, name, updatedName string) func(*testing.T) {
	return func(t *testing.T) {
		err := Rename(db, name, updatedName)
		assert.NoError(t, err)

		_, err = GetCheap(db, name)
		assert.Error(t, err)
	}
}

func getCheap(db *bolt.DB, expectedName string) func(*testing.T) {
	return func(t *testing.T) {
		gotName, err := GetCheap(db, expectedName)
		assert.NoError(t, err)

		assert.Equal(t, expectedName, gotName.Name)
	}
}

func list(db *bolt.DB) func(*testing.T) {
	return func(t *testing.T) {
		files, err := List(db)
		assert.NoError(t, err)

		assert.NotZero(t, len(files), "Expected one or more files")
	}
}

func listNames(db *bolt.DB, expectedName string) func(*testing.T) {
	return func(t *testing.T) {
		files, err := ListNames(db)
		assert.NoError(t, err)

		assert.NotZero(t, len(files), "Expected one or more files")

		gotName := files[0]
		assert.Equal(t, expectedName, gotName)
	}
}

func remove(db *bolt.DB, name string) func(*testing.T) {
	return func(t *testing.T) {
		err := Remove(db, name)
		assert.NoError(t, err)
	}
}

func TestRemoveNone(t *testing.T) {
	db := dbutil.SetContext(t, bucket.File.GetName())

	err := Remove(db)
	assert.NoError(t, err)
}

func TestCreateErrors(t *testing.T) {
	db := dbutil.SetContext(t, bucket.File.GetName())

	err := Create(db, &pb.File{})
	assert.Error(t, err)
}

func TestGetError(t *testing.T) {
	db := setContext(t)

	_, err := Get(db, "non-existent")
	assert.Error(t, err)
}

func TestGetCheapError(t *testing.T) {
	db := setContext(t)

	_, err := GetCheap(db, "non-existent")
	assert.Error(t, err)
}

func TestRenameError(t *testing.T) {
	db := setContext(t)

	err := Rename(db, "non-existent", "")
	assert.Error(t, err)
}

func TestCryptErrors(t *testing.T) {
	db := setContext(t)

	name := "crypt-errors"
	err := Create(db, &pb.File{Name: name})
	assert.NoError(t, err)

	// Try to get the file with other password
	config.Set("auth.password", memguard.NewEnclave([]byte("invalid")))

	_, err = Get(db, name)
	assert.Error(t, err)

	_, err = GetCheap(db, name)
	assert.Error(t, err)

	_, err = List(db)
	assert.Error(t, err)

	err = Rename(db, name, "fail")
	assert.Error(t, err)
}

func TestProtoErrors(t *testing.T) {
	db := setContext(t)

	name := "unformatted"
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket.File.GetName())
		buf := make([]byte, 64)
		encBuf, _ := crypt.Encrypt(buf)
		return b.Put([]byte(name), encBuf)
	})
	assert.NoError(t, err, "Failed writing invalid type")

	_, err = Get(db, name)
	assert.Error(t, err)

	_, err = GetCheap(db, name)
	assert.Error(t, err)

	_, err = List(db)
	assert.Error(t, err)

	err = Rename(db, name, "fail")
	assert.Error(t, err)
}

func TestKeyError(t *testing.T) {
	db := setContext(t)

	err := Create(db, &pb.File{})
	assert.Error(t, err)

	err = Rename(db, "", "")
	assert.Error(t, err)
}

func setContext(t testing.TB) *bolt.DB {
	return dbutil.SetContext(t, bucket.File.GetName())
}
