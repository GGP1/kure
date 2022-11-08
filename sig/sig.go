// Package sig implements signal events handling.
package sig

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/awnumar/memguard"
	bolt "go.etcd.io/bbolt"
)

// Signal is the element used to handle interruptions.
var Signal = sig{}

type sig struct {
	mu sync.RWMutex
	// channel listening to signals
	interrupt chan os.Signal
	// list of functions to be executed after a signal
	cleanups []func() error
	// keepAlive prevents the program from exiting on a signal,
	// it should be always accessed atomically
	//
	// 1 -> do not exit
	//
	// 0 -> exit
	keepAlive int32
}

// AddCleanup adds a function to be executed on a signal.
// First added, first called order.
func (s *sig) AddCleanup(f func() error) {
	s.mu.Lock()
	s.cleanups = append(s.cleanups, f)
	s.mu.Unlock()
}

// Interrupt stops the process but keeps it alive, to force exit use Kill().
func (s *sig) Interrupt() {
	atomic.StoreInt32(&s.keepAlive, 1)
	s.interrupt <- os.Interrupt
}

// KeepAlive prevents the program from exiting after a signal is received.
//
// Example: shutdown a running server without terminating a running session.
func (s *sig) KeepAlive() {
	atomic.StoreInt32(&s.keepAlive, 1)
}

// Kill forces releasing resources, deleting sensitive information and exiting.
func (s *sig) Kill() {
	atomic.StoreInt32(&s.keepAlive, 0)
	s.interrupt <- os.Kill

	// Block and wait for the program to exit
	block := make(chan struct{})
	<-block
}

// Listen listens for a signal to release resources, delete any sensitive information
// and exit safely. It should be called only once and in the main file.
//
// If keepAlive is true it won't exit. Cleanup functions are executed in any case.
//
// db.Close() will block waiting for open transactions to finish before closing.
func (s *sig) Listen(db *bolt.DB) {
	// interrupt gets updated on each call to Listen
	s.interrupt = make(chan os.Signal, 1)
	signal.Notify(s.interrupt, os.Interrupt, syscall.SIGHUP, syscall.SIGTERM)

	go func() {
		<-s.interrupt

		s.mu.RLock()
		for _, f := range s.cleanups {
			if err := f(); err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
			}
		}
		s.mu.RUnlock()

		if atomic.LoadInt32(&s.keepAlive) == 1 {
			// Reset keep alive state
			atomic.StoreInt32(&s.keepAlive, 0)
			s.Listen(db)
			return
		}

		db.Close()
		fmt.Println("\nExiting...")
		memguard.SafeExit(0)
	}()
}

// ResetCleanups empties cleanup functions.
func (s *sig) ResetCleanups() {
	s.mu.Lock()
	s.cleanups = nil
	s.mu.Unlock()
}
