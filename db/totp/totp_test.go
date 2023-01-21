package totp

import (
	"crypto/rand"
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

func TestTOTP(t *testing.T) {
	db := setContext(t)

	totp := &pb.TOTP{
		Name:   "test",
		Raw:    "IFGEWRKSIFJUMR2R",
		Digits: 6,
	}

	// Create destroys the buffer, hence we cannot use their fields anymore
	t.Run("Create", create(db, totp))
	t.Run("Get", get(db, totp))
	t.Run("List", list(db, totp))
	t.Run("List names", listNames(db, totp))
	t.Run("Remove", remove(db, totp.Name))
}

func create(db *bolt.DB, totp *pb.TOTP) func(*testing.T) {
	return func(t *testing.T) {
		err := Create(db, totp)
		assert.NoError(t, err)
	}
}

func get(db *bolt.DB, expected *pb.TOTP) func(*testing.T) {
	return func(t *testing.T) {
		got, err := Get(db, expected.Name)
		assert.NoError(t, err)

		equal := proto.Equal(expected, got)
		assert.True(t, equal)
	}
}

func list(db *bolt.DB, expected *pb.TOTP) func(*testing.T) {
	return func(t *testing.T) {
		totps, err := List(db)
		assert.NoError(t, err)

		assert.NotZero(t, len(totps), "Expected one or more totps")

		got := totps[0]
		equal := proto.Equal(expected, got)
		assert.True(t, equal)
	}
}

func listNames(db *bolt.DB, expected *pb.TOTP) func(*testing.T) {
	return func(t *testing.T) {
		totps, err := ListNames(db)
		assert.NoError(t, err)

		assert.NotZero(t, len(totps), "Expected one or more totps")

		got := totps[0]
		assert.Equal(t, expected.Name, got)
	}
}

func remove(db *bolt.DB, name string) func(*testing.T) {
	return func(t *testing.T) {
		err := Remove(db, name)
		assert.NoError(t, err)
	}
}

func TestCreateErrors(t *testing.T) {
	db := setContext(t)

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

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			err := Create(db, &pb.TOTP{Name: tc.name})
			assert.Error(t, err)
		})
	}
}

func TestGetError(t *testing.T) {
	db := setContext(t)

	_, err := Get(db, "non-existent")
	assert.Error(t, err)
}

func TestCryptErrors(t *testing.T) {
	db := setContext(t)

	// Create the one used by Get and List
	name := "test"
	err := Create(db, &pb.TOTP{Name: name})
	assert.NoError(t, err)

	// Try to get the TOTP with another password
	config.Set("auth.password", memguard.NewEnclave([]byte("invalid")))

	_, err = Get(db, name)
	assert.Error(t, err)
	_, err = List(db)
	assert.Error(t, err)
}

func TestProtoErrors(t *testing.T) {
	db := setContext(t)

	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket.TOTP.GetName())
		buf := make([]byte, 64)
		rand.Read(buf)
		encBuf, _ := crypt.Encrypt(buf)
		return b.Put([]byte("unformatted"), encBuf)
	})
	assert.NoError(t, err, "Failed writing invalid type")

	_, err = Get(db, "unformatted")
	assert.Error(t, err)

	_, err = List(db)
	assert.Error(t, err)
}

func TestKeyError(t *testing.T) {
	db := setContext(t)

	err := Create(db, &pb.TOTP{Name: ""})
	assert.Error(t, err)
}

func setContext(t testing.TB) *bolt.DB {
	return dbutil.SetContext(t, bucket.TOTP.GetName())
}
