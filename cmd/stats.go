package cmd

import (
	"fmt"
	"strings"

	"github.com/GGP1/kure/db"
	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show database statistics",
	Run: func(cmd *cobra.Command, args []string) {
		stats, err := db.Stats()
		if err != nil {
			fatal(err)
		}

		nCards := stats["cards"]
		nEntries := stats["entries"]
		nFiles := stats["files"]
		nWallets := stats["wallets"]

		totalRecords := nCards + nEntries + nFiles + nWallets
		totalBuckets := len(stats)

		bucketsList, err := db.ListOfBuckets()
		if err != nil {
			fatal(err)
		}

		buckets := strings.Join(bucketsList, ", ")

		fmt.Printf(`     STATISTICS
────────────────────
Number of cards: %d
Number of entries: %d
Number of files: %d
Number of wallets: %d

Total records: %d
Number of buckets: %d
List of buckets: %s
		`, nCards, nEntries, nFiles, nWallets, totalRecords, totalBuckets, buckets)
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)
}
