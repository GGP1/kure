package cmd

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/GGP1/kure/db"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
)

var timeout time.Duration
var bCard string

var copyCmd = &cobra.Command{
	Use:   "copy <title> [-t timeout]",
	Short: "Copy password to clipboard",
	Run: func(cmd *cobra.Command, args []string) {
		title := strings.Join(args, " ")

		entry, err := db.GetEntry(title)
		if err != nil {
			log.Fatal("error: ", err)
		}

		if err := clipboard.WriteAll(entry.Password); err != nil {
			log.Fatal("error: ", err)
		}

		if timeout > 0 {
			<-time.After(timeout)
			clipboard.WriteAll("")
			os.Exit(1)
		}
	},
}

func init() {
	RootCmd.AddCommand(copyCmd)
	copyCmd.Flags().DurationVarP(&timeout, "timeout", "t", 0, "clipboard cleaning timeout")
}

func copyEntry(title string) error {

	return nil
}
