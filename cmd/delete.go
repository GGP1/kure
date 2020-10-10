package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/GGP1/kure/db"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete an entry",
	Run: func(cmd *cobra.Command, args []string) {
		name := strings.Join(args, " ")

		if _, err := db.GetEntry(name); err != nil {
			fatal(err)
		}

		scanner := bufio.NewScanner(os.Stdin)
		fmt.Print("Are you sure you want to proceed? [y/n]: ")

		scanner.Scan()
		input := strings.ToLower(scanner.Text())

		if strings.Contains(input, "y") || strings.Contains(input, "yes") {
			if err := db.DeleteEntry(name); err != nil {
				fatal(err)
			}

			fmt.Printf("\nSuccessfully deleted %s entry.", name)
		}
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
