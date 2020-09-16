package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/GGP1/kure/db"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <title>",
	Short: "Delete an entry",
	Run: func(cmd *cobra.Command, args []string) {
		title := strings.Join(args, " ")

		_, err := db.GetEntry(title)
		if err != nil {
			log.Fatal("error: this entry does not exist")
		}

		scanner := bufio.NewScanner(os.Stdin)
		fmt.Print("Are you sure you want to proceed? [y/n]: ")

		scanner.Scan()
		text := scanner.Text()
		input := strings.ToLower(text)

		if strings.Contains(input, "y") || strings.Contains(input, "yes") {
			if err := db.DeleteEntry(title); err != nil {
				log.Fatal("error: ", err)
			}

			fmt.Printf("\nSuccessfully deleted %s entry.", title)
		}
	},
}

func init() {
	RootCmd.AddCommand(deleteCmd)
}
