package sig

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	bolt "go.etcd.io/bbolt"
)

func TestAddCleanup(t *testing.T) {
	f := func() error { return errors.New("test") }
	Signal.AddCleanup(f)

	for _, cf := range Signal.cleanups {
		got := cf().Error()
		expected := f().Error()
		assert.Equal(t, expected, got)
	}
}

func TestKeepAlive(t *testing.T) {
	Signal.KeepAlive()
	assert.Equal(t, Signal.keepAlive, int32(1))
	// Reset
	atomic.StoreInt32(&Signal.keepAlive, 0)
}

func TestListenKeepAlive(t *testing.T) {
	db, err := bolt.Open("../db/testdata/database", 0o600, &bolt.Options{Timeout: 1 * time.Second})
	assert.NoError(t, err, "Failed connecting to the database")
	defer db.Close()

	Signal.Listen(db)
	Signal.Interrupt()
}
