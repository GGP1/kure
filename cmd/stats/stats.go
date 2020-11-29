package stats

import (
	"fmt"
	"strings"

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
		RunE:    runStats(db),
	}

	return cmd
}

func runStats(db *bolt.DB) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		if err := cmdutil.RequirePassword(db); err != nil {
			return err
		}

		tx, err := db.Begin(false)
		if err != nil {
			return errors.Wrap(err, "transaction failed")
		}

		nCards := tx.Bucket([]byte("kure_card")).Stats().KeyN
		nEntries := tx.Bucket([]byte("kure_entry")).Stats().KeyN
		nFiles := tx.Bucket([]byte("kure_file")).Stats().KeyN
		nWallets := tx.Bucket([]byte("kure_wallet")).Stats().KeyN
		totalRecords := nCards + nEntries + nFiles + nWallets
		tx.Rollback()

		list, err := listOfBuckets(db)
		if err != nil {
			return err
		}

		totalBuckets := len(list)
		buckets := strings.Join(list, ", ")

		fmt.Printf(`
     STATISTICS
────────────────────
Number of cards: %d
Number of entries: %d
Number of files: %d
Number of wallets: %d

Total records: %d
Number of buckets: %d
Buckets list: %s
`, nCards, nEntries, nFiles, nWallets, totalRecords, totalBuckets, buckets)

		return nil
	}
}

// listOfBuckets returns a list of the existing buckets.
func listOfBuckets(db *bolt.DB) ([]string, error) {
	var buckets []string
	err := db.View(func(tx *bolt.Tx) error {
		return tx.ForEach(func(name []byte, _ *bolt.Bucket) error {
			buckets = append(buckets, string(name))
			return nil
		})
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed listing buckets")
	}

	return buckets, nil
}
