package argon2

import (
	"encoding/binary"
	"testing"

	cmdutil "github.com/GGP1/kure/commands"

	"github.com/stretchr/testify/assert"
	bolt "go.etcd.io/bbolt"
)

func TestArgon2(t *testing.T) {
	db := cmdutil.SetContext(t, "../../../db/testdata/database")

	err := db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("kure_auth"))
		if err != nil {
			return err
		}

		keys := []string{"iterations", "memory", "threads"}
		value := make([]byte, 4)
		binary.BigEndian.PutUint32(value, 1)

		for _, key := range keys {
			if err := b.Put([]byte(key), value); err != nil {
				return err
			}
		}
		return nil
	})
	assert.NoError(t, err)

	err = NewCmd(db).Execute()
	assert.NoError(t, err, "Failed printing argon2 parameters")
	// Output:
	// Iterations: 1
	// Memory: 1
	// Threads: 1
}
