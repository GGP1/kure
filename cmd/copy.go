package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/GGP1/kure/db"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
)

var (
	timeout time.Duration
	copyCmd = &cobra.Command{
		Use:   "copy",
		Short: "Copy password to clipboard.",
		Run: func(cmd *cobra.Command, args []string) {
			entry, err := db.GetEntry(title)
			if err != nil {
				fmt.Println(err)
			}

			err = clipboard.WriteAll(string(entry.Password))
			if err != nil {
				fmt.Println(err)
			}

			if timeout > 0 {
				<-time.After(timeout)
				clipboard.WriteAll("")
				os.Exit(1)
			}
		},
	}
)

func init() {
	RootCmd.AddCommand(copyCmd)
	copyCmd.Flags().StringVarP(&title, "title", "t", "", "entry title")
	copyCmd.Flags().DurationVarP(&timeout, "clean", "c", 0, "clipboard cleaning timeout")
}
