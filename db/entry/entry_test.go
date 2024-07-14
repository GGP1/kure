package entry

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
	t.Run("Get", get(db, e))
	t.Run("List", list(db, e, e2))
	t.Run("List names", listNames(db, names))
	t.Run("Remove", remove(db, e.Name, e2.Name))
	t.Run("Update", update(db))
	t.Run("Update name", updateName(db))
}

func create(db *bolt.DB, entries ...*pb.Entry) func(*testing.T) {
	return func(t *testing.T) {
		err := Create(db, entries...)
		assert.NoError(t, err)
	}
}

func get(db *bolt.DB, expected *pb.Entry) func(*testing.T) {
	return func(t *testing.T) {
		got, err := Get(db, expected.Name)
		assert.NoError(t, err)

		if !proto.Equal(expected, got) {
			t.Errorf("Expected %v, got %v", expected, got)
		}
	}
}

func list(db *bolt.DB, expected ...*pb.Entry) func(*testing.T) {
	return func(t *testing.T) {
		entries, err := List(db)
		assert.NoError(t, err)

		for _, got := range entries {
			for _, exp := range expected {
				if got.Name == exp.Name {
					equal := proto.Equal(exp, got)
					assert.True(t, equal)
				}
			}
		}
	}
}

func listNames(db *bolt.DB, names map[string]struct{}) func(*testing.T) {
	return func(t *testing.T) {
		entryNames, err := ListNames(db)
		assert.NoError(t, err)

		for _, name := range entryNames {
			_, ok := names[name]
			assert.Truef(t, ok, "Expected %q to be in the list but it isn't", name)
		}
	}
}

func remove(db *bolt.DB, names ...string) func(*testing.T) {
	return func(t *testing.T) {
		err := Remove(db, names...)
		assert.NoError(t, err)
	}
}

func update(db *bolt.DB) func(*testing.T) {
	return func(t *testing.T) {
		oldEntry := &pb.Entry{Name: "test"}
		err := Create(db, oldEntry)
		assert.NoError(t, err)

		newEntry := &pb.Entry{Name: "test", Username: "username"}
		err = Update(db, oldEntry.Name, newEntry)
		assert.NoError(t, err)

		_, err = Get(db, newEntry.Name)
		assert.NoError(t, err)
	}
}

func updateName(db *bolt.DB) func(*testing.T) {
	return func(t *testing.T) {
		oldEntry := &pb.Entry{Name: "old"}
		err := Create(db, oldEntry)
		assert.NoError(t, err)

		newEntry := &pb.Entry{Name: "new"}
		err = Update(db, oldEntry.Name, newEntry)
		assert.NoError(t, err)

		_, err = Get(db, newEntry.Name)
		assert.NoError(t, err)

		_, err = Get(db, oldEntry.Name)
		assert.Error(t, err)
	}
}

func TestCreateNone(t *testing.T) {
	db := setContext(t)
	err := Create(db)
	assert.NoError(t, err)

	names, err := ListNames(db)
	assert.NoError(t, err)

	assert.Zero(t, len(names), "Expected no entries")
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
			err := Create(db, &pb.Entry{Name: tc.name})
			assert.Error(t, err)
		})
	}
}

func TestGetErrors(t *testing.T) {
	db := setContext(t)

	_, err := Get(db, "non-existent")
	assert.Error(t, err)
}

func TestUpdateError(t *testing.T) {
	db := setContext(t)

	name := string([]rune{'\x00'})
	err := Update(db, "old", &pb.Entry{Name: name})
	assert.Error(t, err)
}

func TestCryptErrors(t *testing.T) {
	db := setContext(t)

	name := "test decrypt error"

	e := &pb.Entry{Name: name, Expires: "Never"}
	err := Create(db, e)
	assert.NoError(t, err)

	// Try to get the entry with other password
	config.Set("auth.password", memguard.NewEnclave([]byte("invalid")))

	_, err = Get(db, name)
	assert.Error(t, err)
	_, err = List(db)
	assert.Error(t, err)
}

func TestProtoErrors(t *testing.T) {
	db := setContext(t)

	name := "unformatted"
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket.Entry.GetName())
		buf := make([]byte, 64)
		encBuf, _ := crypt.Encrypt(buf)
		return b.Put([]byte(name), encBuf)
	})
	assert.NoError(t, err, "Failed writing invalid type")

	_, err = Get(db, name)
	assert.Error(t, err)
	_, err = List(db)
	assert.Error(t, err)
}

func TestKeyError(t *testing.T) {
	db := setContext(t)

	err := Create(db, &pb.Entry{Name: ""})
	assert.Error(t, err)
}

func setContext(t testing.TB) *bolt.DB {
	return dbutil.SetContext(t, bucket.Entry.GetName())
}
