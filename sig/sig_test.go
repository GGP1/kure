package sig

import (
	"errors"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	bolt "go.etcd.io/bbolt"
)

func TestAddCleanup(t *testing.T) {
	f := func() error { return errors.New("test") }

	t.Run("Sequentially", func(t *testing.T) {
		Signal.ResetCleanups()
		Signal.AddCleanup(f)

		for _, cf := range Signal.cleanups {
			got := cf().Error()
			expected := f().Error()
			assert.Equal(t, expected, got)
		}
	})

	t.Run("Concurrently", func(t *testing.T) {
		Signal.ResetCleanups()
		count := 5

		var wg sync.WaitGroup
		wg.Add(count)
		for range count {
			go func() {
				Signal.AddCleanup(f)
				wg.Done()
			}()
		}
		wg.Wait()

		assert.Equal(t, count, len(Signal.cleanups))
	})
}

func TestKeepAlive(t *testing.T) {
	Signal.KeepAlive()
	assert.Equal(t, Signal.keepAlive, int32(1))
	// Reset
	atomic.StoreInt32(&Signal.keepAlive, 0)
}

func TestListenKeepAlive(t *testing.T) {
	dbFile, err := os.CreateTemp("", "*")
	assert.NoError(t, err)

	db, err := bolt.Open(dbFile.Name(), 0o600, &bolt.Options{Timeout: 1 * time.Second})
	assert.NoError(t, err, "Failed connecting to the database")
	defer db.Close()

	Signal.Listen(db)
	Signal.Interrupt()
}
