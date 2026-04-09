package stats

import (
	"encoding/json"
	"fmt"
	"os"

	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/bucket"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

const example = `
* Show statistics
kure stats

* Show statistics in JSON format
kure stats --json`

type statsOptions struct {
	json bool
}

// NewCmd returns a new command.
func NewCmd(db *bolt.DB) *cobra.Command {
	opts := statsOptions{}
	cmd := &cobra.Command{
		Use:     "stats",
		Short:   "Show database statistics",
		Example: example,
		RunE:    runStats(db, &opts),
	}

	f := cmd.Flags()
	f.BoolVar(&opts.json, "json", false, "output statistics in JSON format")

	return cmd
}

func runStats(db *bolt.DB, opts *statsOptions) cmdutil.RunEFunc {
	return func(_ *cobra.Command, _ []string) error {
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

		if opts.json {
			stats := map[string]int{
				"cards":   nCards,
				"entries": nEntries,
				"files":   nFiles,
				"totps":   nTOTPs,
				"total":   total,
			}
			if err := json.NewEncoder(os.Stdout).Encode(stats); err != nil {
				return errors.Wrap(err, "encoding statistics to JSON")
			}
			return nil
		}

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
