package totp

import (
	"crypto/rand"
	"testing"
	"time"

	"github.com/GGP1/kure/crypt"
	"github.com/GGP1/kure/pb"

	"github.com/awnumar/memguard"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	bolt "go.etcd.io/bbolt"
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
	t.Run("Get", get(db, totp.Name))
	t.Run("List", list(db))
	t.Run("List names", listNames(db))
	t.Run("Remove", remove(db, totp.Name))
}

func create(db *bolt.DB, totp *pb.TOTP) func(*testing.T) {
	return func(t *testing.T) {
		if err := Create(db, totp); err != nil {
			t.Fatalf("Create() failed: %v", err)
		}
	}
}

func get(db *bolt.DB, name string) func(*testing.T) {
	return func(t *testing.T) {
		got, err := Get(db, name)
		if err != nil {
			t.Fatalf("Get() failed: %v", err)
		}

		// They aren't DeepEqual
		if got.Name != name {
			t.Errorf("Expected %s, got %s", name, got.Name)
		}
	}
}

func list(db *bolt.DB) func(*testing.T) {
	return func(t *testing.T) {
		totps, err := List(db)
		if err != nil {
			t.Fatalf("List() failed: %v", err)
		}

		if len(totps) == 0 {
			t.Error("Expected one or more totps, got 0")
		}
	}
}

func listNames(db *bolt.DB) func(*testing.T) {
	return func(t *testing.T) {
		totps, err := ListNames(db)
		if err != nil {
			t.Error(err)
		}
		if len(totps) == 0 {
			t.Error("Expected one or more totps, got 0")
		}

		expected := "test"
		got := totps[0]

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
			if err := Create(db, &pb.TOTP{Name: tc.name}); err == nil {
				t.Error("Expected an error and got nil")
			}
		})
	}
}

func TestGetError(t *testing.T) {
	db := setContext(t)

	if _, err := Get(db, "non-existent"); err == nil {
		t.Error("Expected 'does not exist' error, got nil")
	}
}

func TestCryptErrors(t *testing.T) {
	db := setContext(t)

	// Create the one used by Get and List
	name := "test"
	if err := Create(db, &pb.TOTP{Name: name}); err != nil {
		t.Fatal(err)
	}

	// Try to get the TOTP with another password
	viper.Set("auth.password", memguard.NewEnclave([]byte("invalid")))

	if _, err := Get(db, name); err == nil {
		t.Error("Expected Get() to fail but it didn't")
	}
	if _, err := List(db); err == nil {
		t.Error("Expected List() to fail but it didn't")
	}
}

func TestProtoErrors(t *testing.T) {
	db := setContext(t)

	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(totpBucket))
		buf := make([]byte, 64)
		rand.Read(buf)
		encBuf, _ := crypt.Encrypt(buf)
		return b.Put([]byte("unformatted"), encBuf)
	})
	if err != nil {
		t.Fatalf("Failed writing invalid type: %v", err)
	}

	if _, err := Get(db, "unformatted"); err == nil {
		t.Error("Expected Get() to fail but it didn't")
	}
	if _, err := List(db); err == nil {
		t.Error("Expected List() to fail but it didn't")
	}
}

func TestKeyError(t *testing.T) {
	db := setContext(t)

	if err := Create(db, &pb.TOTP{Name: ""}); err == nil {
		t.Error("Expected Create() to fail but it didn't")
	}
}

func TestListNameNil(t *testing.T) {
	db := setContext(t)
	err := db.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket(totpBucket)
	})
	if err != nil {
		t.Fatalf("Failed deleting the file bucket: %v", err)
	}

	list, err := ListNames(db)
	if err != nil || list != nil {
		t.Errorf("Expected to receive a nil list and error, got: %v list, %v error", list, err)
	}
}

func setContext(t *testing.T) *bolt.DB {
	db, err := bolt.Open("../testdata/database", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		t.Fatalf("Failed connecting to the database: %v", err)
	}

	viper.Reset()
	// Reduce argon2 parameters to speed up tests
	auth := map[string]interface{}{
		"password":   memguard.NewEnclave([]byte("1")),
		"iterations": 1,
		"memory":     1,
		"threads":    1,
	}
	viper.Set("auth", auth)

	err = db.Update(func(tx *bolt.Tx) error {
		bucket := "kure_totp"
		tx.DeleteBucket([]byte(bucket))
		if _, err := tx.CreateBucketIfNotExists([]byte(bucket)); err != nil {
			return errors.Wrapf(err, "couldn't create %q bucket", bucket)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("Failed closing the database: %v", err)
		}
	})

	return db
}
