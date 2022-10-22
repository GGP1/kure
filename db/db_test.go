package dbutil_test

import (
	"testing"

	"github.com/GGP1/kure/config"
	"github.com/GGP1/kure/crypt"
	dbutil "github.com/GGP1/kure/db"
	"github.com/GGP1/kure/pb"

	"github.com/awnumar/memguard"
	"github.com/stretchr/testify/assert"
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
	assert.NoError(t, err)

	got := &pb.Card{}
	err = dbutil.Get(db, record.Name, got)
	assert.NoError(t, err)

	equal := proto.Equal(record, got)
	assert.True(t, equal)
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
			assert.Equal(t, tc.expected, got)
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
	assert.NoError(t, err)

	for _, e := range expected {
		for _, g := range got {
			if e.Name == g.Name {
				equal := proto.Equal(e, g)
				assert.True(t, equal)
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
	assert.NoError(t, err)

	got, err := dbutil.ListNames(db, bucketName)
	assert.NoError(t, err)

	// We expect them to be ordered
	expected := []string{recordA, recordB}
	assert.Equal(t, expected, got)
}

func TestListNamesNil(t *testing.T) {
	db := dbutil.SetContext(t, "./testdata/database", bucketName)

	err := db.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket(bucketName)
	})
	assert.NoError(t, err, "Failed deleting the file bucket")

	list, err := dbutil.ListNames(db, bucketName)
	assert.NoError(t, err)
	assert.Nil(t, list)
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
	assert.NoError(t, err)

	err = dbutil.Remove(db, bucketName, recordA)
	assert.NoError(t, err)

	expected := make([]string, 0)
	got, err := dbutil.ListNames(db, bucketName)
	assert.NoError(t, err)

	assert.Equal(t, expected, got)
}

func TestRemoveNone(t *testing.T) {
	err := dbutil.Remove(nil, nil)
	assert.NoError(t, err)
}

func TestCryptErrors(t *testing.T) {
	db := dbutil.SetContext(t, "./testdata/database", bucketName)

	name := "test decrypt error"

	e := &pb.Entry{Name: name, Expires: "Never"}
	createRecord(t, db, e)

	// Try to get the entry with other password
	config.Set("auth.password", memguard.NewEnclave([]byte("invalid")))

	err := dbutil.Get(db, name, &pb.Entry{})
	assert.Error(t, err)
	// For some reason it does not fail if a card struct is used
	_, err = dbutil.List(db, &pb.Entry{})
	assert.Error(t, err)
}

func TestGetErrors(t *testing.T) {
	db := dbutil.SetContext(t, "./testdata/database", bucketName)

	err := dbutil.Get(db, "non-existent", &pb.Entry{})
	assert.Error(t, err)
}

func TestKeyError(t *testing.T) {
	db := dbutil.SetContext(t, "./testdata/database", bucketName)
	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		return dbutil.Put(b, &pb.Entry{Name: ""})
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
	assert.NoError(t, err, "Failed writing invalid type")

	err = dbutil.Get(db, name, &pb.Entry{})
	assert.Error(t, err)
	_, err = dbutil.List(db, &pb.Entry{})
	assert.Error(t, err)
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
				err := dbutil.Put(b, &pb.Entry{Name: tc.name})
				assert.Error(t, err)
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
	assert.NoError(t, err)
}
