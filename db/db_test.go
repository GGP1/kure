package dbutil_test

import (
	"testing"

	"github.com/GGP1/kure/config"
	"github.com/GGP1/kure/crypt"
	dbutil "github.com/GGP1/kure/db"
	"github.com/GGP1/kure/db/bucket"
	"github.com/GGP1/kure/pb"

	"github.com/awnumar/memguard"
	"github.com/stretchr/testify/assert"
	bolt "go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"
)

var (
	record = &pb.Card{
		Name:         "test",
		Number:       "12313121",
		SecurityCode: "007",
	}
	bucketName      = dbutil.GetBucketName(record)
	namesBucketName = dbutil.GetNamesBucketName(record)
)

func TestGet(t *testing.T) {
	db := dbutil.SetContext(t, bucketName, namesBucketName)

	err := db.Update(func(tx *bolt.Tx) error {
		return dbutil.Put(tx, record)
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
			expected: bucket.Entry.GetName(),
		},
		{
			desc:     "Card",
			record:   &pb.Card{},
			expected: bucket.Card.GetName(),
		},
		{
			desc:     "File",
			record:   &pb.File{},
			expected: bucket.File.GetName(),
		},
		{
			desc:     "TOTP",
			record:   &pb.TOTP{},
			expected: bucket.TOTP.GetName(),
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
	db := dbutil.SetContext(t, bucketName, namesBucketName)

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
	db := dbutil.SetContext(t, bucketName, namesBucketName)

	recordA := "a"
	recordB := "b"
	names := []string{recordA, recordB}
	err := db.Update(func(tx *bolt.Tx) error {
		for _, name := range names {
			if err := dbutil.Put(tx, &pb.Card{Name: name}); err != nil {
				return err
			}
		}
		return nil
	})
	assert.NoError(t, err)

	got, err := dbutil.ListNames(db, namesBucketName)
	assert.NoError(t, err)

	// We expect them to be ordered
	assert.Equal(t, names, got)
}

func TestListNamesNil(t *testing.T) {
	db := dbutil.SetContext(t, namesBucketName)

	list, err := dbutil.ListNames(db, namesBucketName)
	assert.NoError(t, err)
	assert.Empty(t, list)
}

func TestPut(t *testing.T) {
	db := dbutil.SetContext(t, bucketName, namesBucketName)
	createRecord(t, db, record)
}

func TestRemove(t *testing.T) {
	db := dbutil.SetContext(t, bucketName, namesBucketName)

	recordA := "a"
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		return b.Put([]byte(recordA), nil)
	})
	assert.NoError(t, err)

	err = db.Update(func(tx *bolt.Tx) error {
		return dbutil.Remove(tx, record, recordA)
	})
	assert.NoError(t, err)

	expected := make([]string, 0)
	got, err := dbutil.ListNames(db, namesBucketName)
	assert.NoError(t, err)

	assert.Equal(t, expected, got)
}

func TestRemoveNone(t *testing.T) {
	err := dbutil.Remove(nil, nil)
	assert.NoError(t, err)
}

func TestCryptErrors(t *testing.T) {
	db := dbutil.SetContext(t, bucketName, namesBucketName)

	name := "test decrypt error"

	e := &pb.Card{Name: name}
	createRecord(t, db, e)

	// Try to get the entry with other password
	config.Set("auth.password", memguard.NewEnclave([]byte("invalid")))

	err := dbutil.Get(db, name, &pb.Card{})
	assert.Error(t, err)
	// For some reason it does not fail if a card struct is used
	_, err = dbutil.List(db, &pb.Card{})
	assert.Error(t, err)
}

func TestGetErrors(t *testing.T) {
	db := dbutil.SetContext(t, bucketName, namesBucketName)

	err := dbutil.Get(db, "non-existent", &pb.Card{})
	assert.Error(t, err)
}

func TestKeyError(t *testing.T) {
	db := dbutil.SetContext(t, bucketName)
	db.Update(func(tx *bolt.Tx) error {
		return dbutil.Put(tx, &pb.Card{Name: ""})
	})
}

func TestProtoErrors(t *testing.T) {
	db := dbutil.SetContext(t, bucketName, namesBucketName)

	name := "unformatted"
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket.Card.GetName())
		buf := make([]byte, 32)
		encBuf, _ := crypt.Encrypt(buf)
		return b.Put([]byte(name), encBuf)
	})
	assert.NoError(t, err, "Failed writing invalid type")

	err = dbutil.Get(db, name, &pb.Card{})
	assert.Error(t, err)
	_, err = dbutil.List(db, &pb.Card{})
	assert.Error(t, err)
}

func TestPutErrors(t *testing.T) {
	db := dbutil.SetContext(t, bucketName, namesBucketName)

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
		for _, tc := range cases {
			t.Run(tc.desc, func(t *testing.T) {
				err := dbutil.Put(tx, &pb.Card{Name: tc.name})
				assert.Error(t, err)
			})
		}
		return nil
	})
}

func createRecord(t *testing.T, db *bolt.DB, record dbutil.Record) {
	err := db.Update(func(tx *bolt.Tx) error {
		return dbutil.Put(tx, record)
	})
	assert.NoError(t, err)
}
