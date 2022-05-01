package dbutil_test

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/GGP1/kure/config"
	"github.com/GGP1/kure/crypt"
	dbutil "github.com/GGP1/kure/db"
	"github.com/GGP1/kure/pb"

	"github.com/awnumar/memguard"
	bolt "go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"
)

var (
	bucketName = []byte("test")
	record     = &pb.Card{
		Name:         "test",
		Number:       "12313121",
		SecurityCode: "007",
	}
)

func TestGet(t *testing.T) {
	db := dbutil.SetContext(t, "./testdata/database", bucketName)

	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(dbutil.GetBucketName(record))
		return dbutil.Put(b, record)
	})
	if err != nil {
		t.Fatal(err)
	}

	got := &pb.Card{}
	if err := dbutil.Get(db, record.Name, got); err != nil {
		t.Error(err)
	}

	if !proto.Equal(record, got) {
		t.Errorf("Expected %#v, got %#v", record, got)
	}
}

func TestGetBucketName(t *testing.T) {
	cases := []struct {
		desc     string
		record   dbutil.Record
		expected []byte
	}{
		{
			desc:     "Entry",
			record:   &pb.Entry{},
			expected: []byte("kure_entry"),
		},
		{
			desc:     "Card",
			record:   &pb.Card{},
			expected: []byte("kure_card"),
		},
		{
			desc:     "File",
			record:   &pb.File{},
			expected: []byte("kure_file"),
		},
		{
			desc:     "TOTP",
			record:   &pb.TOTP{},
			expected: []byte("kure_totp"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			got := dbutil.GetBucketName(tc.record)
			if !bytes.Equal(tc.expected, got) {
				t.Errorf("Expected %q, got %q", tc.expected, got)
			}
		})
	}
}

func TestList(t *testing.T) {
	db := dbutil.SetContext(t, "./testdata/database", bucketName)

	createRecord(t, db, record)
	record2 := &pb.Card{
		Name: "west",
	}
	createRecord(t, db, record2)
	expected := []*pb.Card{record, record2}

	got, err := dbutil.List(db, &pb.Card{})
	if err != nil {
		t.Error(err)
	}

	for _, e := range expected {
		for _, g := range got {
			if e.Name == g.Name && !proto.Equal(e, g) {
				t.Errorf("Expected %#v, got %#v", e, g)
			}
		}
	}

}

func TestListNames(t *testing.T) {
	db := dbutil.SetContext(t, "./testdata/database", bucketName)

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

	got, err := dbutil.ListNames(db, bucketName)
	if err != nil {
		t.Error(err)
	}

	// We expect them to be ordered
	expected := []string{recordA, recordB}
	if !reflect.DeepEqual(expected, got) {
		t.Errorf("Expected %v, got %v", expected, got)
	}
}

func TestListNamesNil(t *testing.T) {
	db := dbutil.SetContext(t, "./testdata/database", bucketName)

	err := db.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket(bucketName)
	})
	if err != nil {
		t.Fatalf("Failed deleting the file bucket: %v", err)
	}

	list, err := dbutil.ListNames(db, bucketName)
	if err != nil || list != nil {
		t.Errorf("Expected to receive a nil list and error, got: %v list, %v error", list, err)
	}
}

func TestPut(t *testing.T) {
	db := dbutil.SetContext(t, "./testdata/database", bucketName)

	createRecord(t, db, record)
}

func TestRemove(t *testing.T) {
	db := dbutil.SetContext(t, "./testdata/database", bucketName)

	recordA := "a"
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		return b.Put([]byte(recordA), nil)
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := dbutil.Remove(db, bucketName, recordA); err != nil {
		t.Fatal(err)
	}

	expected := make([]string, 0)
	got, err := dbutil.ListNames(db, bucketName)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(expected, got) {
		t.Errorf("Expected %v, got %v", expected, got)
	}
}

func TestRemoveNone(t *testing.T) {
	if err := dbutil.Remove(nil, nil); err != nil {
		t.Error(err)
	}
}

func TestCryptErrors(t *testing.T) {
	db := dbutil.SetContext(t, "./testdata/database", bucketName)

	name := "test decrypt error"

	e := &pb.Entry{Name: name, Expires: "Never"}
	createRecord(t, db, e)

	// Try to get the entry with other password
	config.Set("auth.password", memguard.NewEnclave([]byte("invalid")))

	if err := dbutil.Get(db, name, &pb.Entry{}); err == nil {
		t.Error("Expected Get() to fail but it didn't")
	}
	// For some reason it does not fail if a card struct is used
	if _, err := dbutil.List(db, &pb.Entry{}); err == nil {
		t.Error("Expected List() to fail but it didn't")
	}
}

func TestGetErrors(t *testing.T) {
	db := dbutil.SetContext(t, "./testdata/database", bucketName)

	if err := dbutil.Get(db, "non-existent", &pb.Entry{}); err == nil {
		t.Error("Expected an error, got nil")
	}
}

func TestKeyError(t *testing.T) {
	db := dbutil.SetContext(t, "./testdata/database", bucketName)

	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		if err := dbutil.Put(b, &pb.Entry{Name: ""}); err == nil {
			t.Error("Put() didn't fail")
		}
		return nil
	})
}

func TestProtoErrors(t *testing.T) {
	db := dbutil.SetContext(t, "./testdata/database", bucketName)

	name := "unformatted"
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(dbutil.EntryBucket))
		buf := make([]byte, 32)
		encBuf, _ := crypt.Encrypt(buf)
		return b.Put([]byte(name), encBuf)
	})
	if err != nil {
		t.Fatalf("Failed writing invalid type: %v", err)
	}

	if err := dbutil.Get(db, name, &pb.Entry{}); err == nil {
		t.Error("Expected Get() to fail but it didn't")
	}
	if _, err := dbutil.List(db, &pb.Entry{}); err == nil {
		t.Error("Expected List() to fail but it didn't")
	}
}

func TestPutErrors(t *testing.T) {
	db := dbutil.SetContext(t, "./testdata/database", bucketName)

	cases := []struct {
		desc string
		name string
	}{
		{
			desc: "Invalid name",
			name: "",
		},
		{
			desc: "Null characters",
			name: string('\x00'),
		},
	}

	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		for _, tc := range cases {
			t.Run(tc.desc, func(t *testing.T) {
				if err := dbutil.Put(b, &pb.Entry{Name: tc.name}); err == nil {
					t.Error("Expected an error and got nil")
				}
			})
		}
		return nil
	})
}

func createRecord(t *testing.T, db *bolt.DB, record dbutil.Record) {
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(dbutil.GetBucketName(record))
		return dbutil.Put(b, record)
	})
	if err != nil {
		t.Fatal(err)
	}
}
