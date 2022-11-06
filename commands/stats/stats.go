package stats

import (
	"fmt"

	"github.com/GGP1/kure/auth"
	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/bucket"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

const example = `
kure stats`

// NewCmd returns a new command.
func NewCmd(db *bolt.DB) *cobra.Command {
	return &cobra.Command{
		Use:     "stats",
		Short:   "Show database statistics",
		Example: example,
		PreRunE: auth.Login(db),
		RunE:    runStats(db),
	}
}

func runStats(db *bolt.DB) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		tx, err := db.Begin(false)
		if err != nil {
			return errors.Wrap(err, "opening transaction")
		}
		defer tx.Rollback()

		nCards := tx.Bucket(bucket.Card.GetName()).Stats().KeyN
		nEntries := tx.Bucket(bucket.Entry.GetName()).Stats().KeyN
		nFiles := tx.Bucket(bucket.File.GetName()).Stats().KeyN
		nTOTPs := tx.Bucket(bucket.TOTP.GetName()).Stats().KeyN
		total := nCards + nEntries + nFiles + nTOTPs

		fmt.Printf(`
     STATISTICS
────────────────────
Number of cards: %d
Number of entries: %d
Number of files: %d
Number of TOTPs: %d

Total elements: %d
`, nCards, nEntries, nFiles, nTOTPs, total)

		return nil
	}
}
