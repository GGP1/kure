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
	bucketName = dbutil.GetBucketName(record)
)

func TestGet(t *testing.T) {
	db := dbutil.SetContext(t, bucketName)

	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
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
	db := dbutil.SetContext(t, bucketName)

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
	db := dbutil.SetContext(t, bucketName)

	recordA := "a"
	recordB := "b"
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		if err := b.Put(dbutil.XorName([]byte(recordA)), nil); err != nil {
			return err
		}
		return b.Put(dbutil.XorName([]byte(recordB)), nil)
	})
	assert.NoError(t, err)

	got, err := dbutil.ListNames(db, bucketName)
	assert.NoError(t, err)

	// We expect the names xored with the auth key. They should be ordered
	expected := []string{recordA, recordB}
	assert.Equal(t, expected, got)
}

func TestListNamesNil(t *testing.T) {
	db := dbutil.SetContext(t, bucketName)

	err := db.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket(bucketName)
	})
	assert.NoError(t, err, "Failed deleting the file bucket")

	list, err := dbutil.ListNames(db, bucketName)
	assert.NoError(t, err)
	assert.Nil(t, list)
}

func TestPut(t *testing.T) {
	db := dbutil.SetContext(t, bucketName)
	createRecord(t, db, record)
}

func TestRemove(t *testing.T) {
	db := dbutil.SetContext(t, bucketName)

	recordA := "a"
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		return b.Put(dbutil.XorName([]byte(recordA)), nil)
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
	db := dbutil.SetContext(t, bucket.Entry.GetName())

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
	db := dbutil.SetContext(t, bucket.Entry.GetName())

	err := dbutil.Get(db, "non-existent", &pb.Entry{})
	assert.Error(t, err)
}

func TestKeyError(t *testing.T) {
	db := dbutil.SetContext(t, bucketName)
	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		return dbutil.Put(b, &pb.Card{Name: ""})
	})
}

func TestProtoErrors(t *testing.T) {
	db := dbutil.SetContext(t, bucket.Entry.GetName())

	name := "unformatted"
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket.Entry.GetName())
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
	db := dbutil.SetContext(t, bucketName)

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
				err := dbutil.Put(b, &pb.Card{Name: tc.name})
				assert.Error(t, err)
			})
		}
		return nil
	})
}

func TestXorName(t *testing.T) {
	defer config.Reset()

	key := []byte{
		51, 0, 107, 95, 158, 240, 55, 129, 1, 249, 4,
		159, 37, 118, 174, 228, 69, 140, 141, 199, 105,
		124, 4, 120, 253, 220, 202, 0, 199, 47, 164, 134,
	}
	config.Set("auth.key", key)

	cases := []struct {
		name     string
		expected string
	}{
		{
			name: "test",
		},
		{
			name: "adidas",
		},
		{
			name: "github",
		},
		{
			name: "nike",
		},
		{
			name: "folder/test",
		},
		{
			name: "super_extralarge_name",
		},
		{
			name: "123456789",
		},
		{
			name: string([]byte{51, 0, 107, 95, 158, 240}),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			xorName := dbutil.XorName([]byte(tc.name))
			assert.NotEqual(t, tc.name, xorName)

			gotName := dbutil.XorName(xorName)
			assert.Equal(t, tc.name, string(gotName))
		})
	}
}

func createRecord(t *testing.T, db *bolt.DB, record dbutil.Record) {
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(dbutil.GetBucketName(record))
		return dbutil.Put(b, record)
	})
	assert.NoError(t, err)
}
