package auth

import (
	"fmt"
	"reflect"
	"testing"

	dbutil "github.com/GGP1/kure/db"

	bolt "go.etcd.io/bbolt"
)

func TestParameters(t *testing.T) {
	db := setContext(t)

	expected := Parameters{
		AuthKey:    []byte("test"),
		Iterations: 1,
		Memory:     1500000,
		Threads:    4,
		UseKeyfile: true,
	}

	if err := Register(db, expected); err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	got, err := GetParameters(db)
	if err != nil {
		t.Fatal(err)
	}
	// Force auth key to be test as it's randomly generated and it won't match
	got.AuthKey = []byte("test")

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Expected %#v, got %#v", expected, got)
	}
}

func TestEmptyParameters(t *testing.T) {
	db := setContext(t)
	tx, _ := db.Begin(true)
	tx.DeleteBucket(authBucket)
	tx.Commit()

	expected := Parameters{}
	got, err := GetParameters(db)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Expected %#v, got %#v", expected, got)
	}
}

func TestSetParametersInvalidKeys(t *testing.T) {
	db := setContext(t)

	params := Parameters{
		AuthKey:    []byte("invalid"),
		Iterations: 1,
		Memory:     1,
		Threads:    1,
		UseKeyfile: true,
	}

	cases := []struct {
		desc string
		key  *[]byte
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
			if err != nil {
				t.Fatalf("Failed opening transaction: %v", err)
			}

			if err := setParameters(tx, params); err == nil {
				t.Error("Expected an error and got nil")
			}
			tx.Commit()

			// Fill the variable so we can test the others
			*tc.key = []byte("1")
		})
	}
}

func setContext(t testing.TB) *bolt.DB {
	return dbutil.SetContext(t, "../testdata/database", authBucket)
}
