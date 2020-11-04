package cmd

import (
	"os"
	"strings"
	"time"

	"github.com/GGP1/kure/db"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
)

var copyCmd = &cobra.Command{
	Use:   "copy <name> [-t timeout] [-u username]",
	Short: "Copy entry credentials to clipboard",
	Run: func(cmd *cobra.Command, args []string) {
		name := strings.Join(args, " ")

		entry, err := db.GetEntry(name)
		if err != nil {
			fatal(err)
		}

		copy := entry.Password

		if username {
			copy = entry.Username
		}

		if err := clipboard.WriteAll(copy); err != nil {
			fatalf("couldn't copy the password to the clipboard: %v", err)
		}

		if timeout > 0 {
			<-time.After(timeout)
			clipboard.WriteAll("")
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(copyCmd)
	copyCmd.Flags().DurationVarP(&timeout, "timeout", "t", 0, "clipboard cleaning timeout")
	copyCmd.Flags().BoolVarP(&username, "username", "u", false, "copy entry username")
}
