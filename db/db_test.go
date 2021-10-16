package dbutil

import (
	"reflect"
	"testing"

	bolt "go.etcd.io/bbolt"
)

var bucketName = []byte("test")

func TestCreateEncoded(t *testing.T) {

}

func TestListEncoded(t *testing.T) {

}

func TestListNames(t *testing.T) {
	db := SetContext(t, "./testdata/database", bucketName)

	recordA := "a"
	recordB := "b"
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		if err := b.Put([]byte(recordB), nil); err != nil {
			return err
		}
		return b.Put([]byte(recordA), nil)
	})
	if err != nil {
		t.Fatal(err)
	}

	got, err := ListNames(db, bucketName)
	if err != nil {
		t.Error(err)
	}

	// We expect them to be ordered
	expected := []string{recordA, recordB}
	if !reflect.DeepEqual(expected, got) {
		t.Errorf("Expected %v, got %v", expected, got)
	}
}

func TestRemove(t *testing.T) {
	db := SetContext(t, "./testdata/database", bucketName)

	recordA := "a"
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		return b.Put([]byte(recordA), nil)
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := Remove(db, bucketName, recordA); err != nil {
		t.Fatal(err)
	}

	expected := make([]string, 0)
	got, err := ListNames(db, bucketName)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(expected, got) {
		t.Errorf("Expected %v, got %v", expected, got)
	}
}

func TestListNamesNil(t *testing.T) {
	db := SetContext(t, "./testdata/database", bucketName)

	err := db.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket(bucketName)
	})
	if err != nil {
		t.Fatalf("Failed deleting the file bucket: %v", err)
	}

	list, err := ListNames(db, bucketName)
	if err != nil || list != nil {
		t.Errorf("Expected to receive a nil list and error, got: %v list, %v error", list, err)
	}
}
