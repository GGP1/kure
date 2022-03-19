package entry

import (
	"testing"

	"github.com/GGP1/kure/config"
	"github.com/GGP1/kure/crypt"
	dbutil "github.com/GGP1/kure/db"
	"github.com/GGP1/kure/pb"

	"github.com/awnumar/memguard"
	bolt "go.etcd.io/bbolt"
)

func TestEntry(t *testing.T) {
	db := setContext(t)

	e := &pb.Entry{
		Name:     "test",
		Username: "testing",
		URL:      "golang.org",
		Expires:  "Never",
		Notes:    "",
	}
	e2 := &pb.Entry{Name: "test2"}
	names := map[string]struct{}{
		e.Name:  {},
		e2.Name: {},
	}

	t.Run("Create", create(db, e, e2))
	t.Run("Get", get(db, e.Name))
	t.Run("List", list(db, names))
	t.Run("List names", listNames(db, names))
	t.Run("Remove", remove(db, e.Name, e2.Name))
	t.Run("Update", update(db))
}

func create(db *bolt.DB, entries ...*pb.Entry) func(*testing.T) {
	return func(t *testing.T) {
		if err := Create(db, entries...); err != nil {
			t.Error(err)
		}
	}
}

func get(db *bolt.DB, name string) func(*testing.T) {
	return func(t *testing.T) {
		got, err := Get(db, name)
		if err != nil {
			t.Error(err)
		}

		// They aren't DeepEqual
		if got.Name != name {
			t.Errorf("Expected %s, got %s", name, got.Name)
		}
	}
}

func list(db *bolt.DB, names map[string]struct{}) func(*testing.T) {
	return func(t *testing.T) {
		entries, err := List(db)
		if err != nil {
			t.Error(err)
		}

		for _, e := range entries {
			if _, ok := names[e.Name]; !ok {
				t.Errorf("Expected %q to be in the list but it isn't", e.Name)
			}
		}
	}
}

func listNames(db *bolt.DB, names map[string]struct{}) func(*testing.T) {
	return func(t *testing.T) {
		entryNames, err := ListNames(db)
		if err != nil {
			t.Error(err)
		}

		for _, name := range entryNames {
			if _, ok := names[name]; !ok {
				t.Errorf("Expected %q to be in the list but it isn't", name)
			}
		}
	}
}

func remove(db *bolt.DB, names ...string) func(*testing.T) {
	return func(t *testing.T) {
		if err := Remove(db, names...); err != nil {
			t.Error(err)
		}
	}
}

func update(db *bolt.DB) func(*testing.T) {
	return func(t *testing.T) {
		oldEntry := &pb.Entry{Name: "old"}
		if err := Create(db, oldEntry); err != nil {
			t.Fatal(err)
		}

		newEntry := &pb.Entry{Name: "new"}
		if err := Update(db, oldEntry.Name, newEntry); err != nil {
			t.Fatal(err)
		}

		if _, err := Get(db, newEntry.Name); err != nil {
			t.Error(err)
		}
	}
}

func TestCreateNone(t *testing.T) {
	db := setContext(t)
	if err := Create(db); err != nil {
		t.Error(err)
	}

	names, err := ListNames(db)
	if err != nil {
		t.Error(err)
	}

	if len(names) != 0 {
		t.Errorf("Expected no entries and got %d", len(names))
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
			if err := Create(db, &pb.Entry{Name: tc.name}); err == nil {
				t.Error("Expected an error and got nil")
			}
		})
	}
}

func TestGetErrors(t *testing.T) {
	db := setContext(t)

	if _, err := Get(db, "non-existent"); err == nil {
		t.Error("Expected an error, got nil")
	}
}

func TestCryptErrors(t *testing.T) {
	db := setContext(t)

	name := "test decrypt error"

	e := &pb.Entry{Name: name, Expires: "Never"}
	if err := Create(db, e); err != nil {
		t.Fatal(err)
	}

	// Try to get the entry with other password
	config.Set("auth.password", memguard.NewEnclave([]byte("invalid")))

	if _, err := Get(db, name); err == nil {
		t.Error("Expected Get() to fail but it didn't")
	}
	if _, err := List(db); err == nil {
		t.Error("Expected List() to fail but it didn't")
	}
}

func TestProtoErrors(t *testing.T) {
	db := setContext(t)

	name := "unformatted"
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(dbutil.EntryBucket))
		buf := make([]byte, 64)
		encBuf, _ := crypt.Encrypt(buf)
		return b.Put([]byte(name), encBuf)
	})
	if err != nil {
		t.Fatalf("Failed writing invalid type: %v", err)
	}

	if _, err := Get(db, name); err == nil {
		t.Error("Expected Get() to fail but it didn't")
	}
	if _, err := List(db); err == nil {
		t.Error("Expected List() to fail but it didn't")
	}
}

func TestKeyError(t *testing.T) {
	db := setContext(t)

	if err := Create(db, &pb.Entry{Name: ""}); err == nil {
		t.Error("Create() didn't fail")
	}
}

func setContext(t testing.TB) *bolt.DB {
	return dbutil.SetContext(t, "../testdata/database", dbutil.EntryBucket)
}
