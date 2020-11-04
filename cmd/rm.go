package cmd

import (
	"fmt"
	"strings"

	"github.com/GGP1/kure/db"

	"github.com/spf13/cobra"
)

var rmCmd = &cobra.Command{
	Use:   "rm <name>",
	Short: "Remove an entry from the database",
	Run: func(cmd *cobra.Command, args []string) {
		name := strings.Join(args, " ")

		if _, err := db.GetEntry(name); err != nil {
			fatal(err)
		}

		if proceed() {
			if err := db.RemoveEntry(name); err != nil {
				fatal(err)
			}

			fmt.Printf("\nSuccessfully removed \"%s\" entry.\n", name)
		}
	},
}

func init() {
	rootCmd.AddCommand(rmCmd)
}
