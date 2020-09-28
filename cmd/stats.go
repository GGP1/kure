package cmd

import (
	"fmt"

	"github.com/GGP1/kure/db"
	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show database statistics",
	Run: func(cmd *cobra.Command, args []string) {
		stats, err := db.Stats()
		if err != nil {
			must(err)
		}

		nEntries := stats["entries"]
		nCards := stats["cards"]
		nWallets := stats["wallets"]
		totalRecords := nEntries + nCards + nWallets

		totalBuckets := len(stats)
		bucketsList, err := db.ListOfBuckets()
		if err != nil {
			must(err)
		}

		fmt.Printf(`     STATISTICS
────────────────────
Number of entries: %d
Number of cards: %d
Number of wallets: %d

Total records: %d
Number of buckets: %d
List of buckets: %v
		`, nEntries, nCards, nWallets, totalRecords, totalBuckets, bucketsList)
	},
}

func init() {
	RootCmd.AddCommand(statsCmd)
}
