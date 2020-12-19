package restore

import (
	"sync"

	"github.com/GGP1/kure/db/card"
	"github.com/GGP1/kure/db/entry"
	"github.com/GGP1/kure/db/file"
	"github.com/GGP1/kure/db/note"
	"github.com/GGP1/kure/pb"

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

func createCards(db *bolt.DB, cards []*pb.Card, wg *sync.WaitGroup) {
	for _, c := range cards {
		go func(c *pb.Card) {
			card.Create(db, c)
			wg.Done()
		}(c)
	}
}

func createEntries(db *bolt.DB, entries []*pb.Entry, wg *sync.WaitGroup) {
	for _, e := range entries {
		go func(e *pb.Entry) {
			entry.Create(db, e)
			wg.Done()
		}(e)
	}
}

func createFiles(db *bolt.DB, files []*pb.File, wg *sync.WaitGroup) {
	for _, f := range files {
		go func(f *pb.File) {
			file.Restore(db, f)
			wg.Done()
		}(f)
	}
}

func createNotes(db *bolt.DB, notes []*pb.Note, wg *sync.WaitGroup) {
	for _, n := range notes {
		go func(f *pb.Note) {
			note.Create(db, f)
			wg.Done()
		}(n)
	}
}
