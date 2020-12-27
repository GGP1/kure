package restore

import (
	"sync"

	"github.com/GGP1/kure/db/card"
	"github.com/GGP1/kure/db/entry"
	"github.com/GGP1/kure/db/file"
	"github.com/GGP1/kure/db/note"
	"github.com/GGP1/kure/pb"
	"github.com/awnumar/memguard"

	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

// NewCmd returns a new command.
func NewCmd(db *bolt.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restore",
		Short: "Restore the database",
		Long: `Restore the records using a new password or different Argon2 parameters.

Warning: interrupting any restoring process may cause irreversible damage in the database information. Use it with caution.`}

	cmd.AddCommand(argon2SubCmd(db), passwordSubCmd(db))

	return cmd
}

// memguard.NewBuffer(1) is used as a mock of each object buffer, the locked buffer passed corresponds
// to the slice and we cannot destroy it until all of the records are created. Underlying objects are also protected
// but there aren't locked buffers for each one of them.
//
// Errors are ignored as there is no user input, they should never fail.
func createCards(db *bolt.DB, buf *memguard.LockedBuffer, cards []*pb.Card, wg *sync.WaitGroup) {
	for _, c := range cards {
		go func(c *pb.Card) {
			card.Create(db, memguard.NewBuffer(1), c)
			wg.Done()
		}(c)
	}
	buf.Destroy()
}

func createEntries(db *bolt.DB, buf *memguard.LockedBuffer, entries []*pb.Entry, wg *sync.WaitGroup) {
	for _, e := range entries {
		go func(e *pb.Entry) {
			entry.Create(db, memguard.NewBuffer(1), e)
			wg.Done()
		}(e)
	}
	buf.Destroy()
}

func createFiles(db *bolt.DB, buf *memguard.LockedBuffer, files []*pb.File, wg *sync.WaitGroup) {
	for _, f := range files {
		go func(f *pb.File) {
			file.Restore(db, f)
			wg.Done()
		}(f)
	}
	buf.Destroy()
}

func createNotes(db *bolt.DB, buf *memguard.LockedBuffer, notes []*pb.Note, wg *sync.WaitGroup) {
	for _, n := range notes {
		go func(f *pb.Note) {
			note.Create(db, memguard.NewBuffer(1), f)
			wg.Done()
		}(n)
	}
	buf.Destroy()
}
