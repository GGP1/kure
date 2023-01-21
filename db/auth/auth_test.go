package auth

import (
	"fmt"
	"testing"

	dbutil "github.com/GGP1/kure/db"
	"github.com/GGP1/kure/db/bucket"

	"github.com/stretchr/testify/assert"
	bolt "go.etcd.io/bbolt"
)

func TestParameters(t *testing.T) {
	db := setContext(t)

	expected := Params{
		AuthKey: []byte("test"),
		Argon2: Argon2{
			Iterations: 1,
			Memory:     1500000,
			Threads:    4,
		},
		UseKeyfile: true,
	}

	err := Register(db, expected)
	assert.NoError(t, err, "Registration failed")

	got, err := GetParams(db)
	assert.NoError(t, err)

	// Force auth key to be test as it's randomly generated and it won't match
	got.AuthKey = []byte("test")
	assert.Equal(t, expected, got)
}

func TestEmptyParameters(t *testing.T) {
	db := setContext(t)
	tx, _ := db.Begin(true)
	tx.DeleteBucket(bucket.Auth.GetName())
	tx.Commit()

	expected := Params{}
	got, err := GetParams(db)
	assert.NoError(t, err)

	assert.Equal(t, expected, got)
}

func TestSetParametersInvalidKeys(t *testing.T) {
	db := setContext(t)

	params := Params{
		AuthKey: []byte("invalid"),
		Argon2: Argon2{
			Iterations: 1,
			Memory:     1,
			Threads:    1,
		},
		UseKeyfile: true,
	}

	cases := []struct {
		key  *[]byte
		desc string
	}{
		{
			desc: "iterations",
			key:  &iterKey,
		},
		{
			desc: "memory",
			key:  &memKey,
		},
		{
			desc: "threads",
			key:  &thKey,
		},
		{
			desc: "keyfile",
			key:  &keyfileKey,
		},
		{
			desc: "auth",
			key:  &authKey,
		},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("Invalid %s key", tc.desc), func(t *testing.T) {
			*tc.key = nil

			tx, err := db.Begin(true)
			assert.NoError(t, err, "Failed opening transaction")

			err = storeParams(tx, params)
			assert.Error(t, err)
			tx.Commit()

			// Fill the variable so we can test the others
			*tc.key = []byte("1")
		})
	}
}

func setContext(t testing.TB) *bolt.DB {
	return dbutil.SetContext(t, bucket.Auth.GetName())
}
