package sig

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"

	bolt "go.etcd.io/bbolt"
)

func TestAddCleanup(t *testing.T) {
	f := func() error { return errors.New("test") }
	Signal.AddCleanup(f)

	for _, cf := range Signal.cleanups {
		got := cf().Error()
		expected := f().Error()

		if got != expected {
			t.Errorf("Expected %q, got %q", expected, got)
		}
	}
}

func TestKeepAlive(t *testing.T) {
	Signal.KeepAlive()
	if Signal.keepAlive != 1 {
		t.Error("Expected keepAlive variable to be false and got true")
	}
	// Reset
	atomic.StoreInt32(&Signal.keepAlive, 0)
}

func TestListenKeepAlive(t *testing.T) {
	db, err := bolt.Open("../db/testdata/database", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		t.Fatalf("Failed connecting to the database: %v", err)
	}
	defer db.Close()

	Signal.Listen(db)
	Signal.Interrupt()
}
