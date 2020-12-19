package stats

import (
	"fmt"

	cmdutil "github.com/GGP1/kure/cmd"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var example = `
* Show statistics on terminal
kure stats`

// NewCmd returns a new command.
func NewCmd(db *bolt.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "stats",
		Short:   "Show database statistics",
		Example: example,
		PreRunE: cmdutil.RequirePassword(db),
		RunE:    runStats(db),
	}

	return cmd
}

func runStats(db *bolt.DB) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		tx, err := db.Begin(false)
		if err != nil {
			return errors.Wrap(err, "transaction failed")
		}

		nCards := tx.Bucket([]byte("kure_card")).Stats().KeyN
		nEntries := tx.Bucket([]byte("kure_entry")).Stats().KeyN
		nFiles := tx.Bucket([]byte("kure_file")).Stats().KeyN
		nNotes := tx.Bucket([]byte("kure_note")).Stats().KeyN
		totalElements := nCards + nEntries + nFiles + nNotes
		tx.Rollback()

		fmt.Printf(`
     STATISTICS
────────────────────
Number of cards: %d
Number of entries: %d
Number of files: %d
Number of notes: %d

Total elements: %d
`, nCards, nEntries, nFiles, nNotes, totalElements)

		return nil
	}
}
