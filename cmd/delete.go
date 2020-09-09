package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/GGP1/kure/crypt"
	"github.com/GGP1/kure/db"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <title>",
	Short: "Delete an entry",
	Run: func(cmd *cobra.Command, args []string) {
		title := strings.Join(args, " ")

		entry, err := db.GetEntry(title)
		if err != nil {
			fmt.Println("error:", err)
			return
		}

		// If the password is encrypted, request it to delete the entry
		if entry.Safe {
			pwd, err := passInput()
			if err != nil {
				fmt.Println("error:", err)
				return
			}

			_, err = crypt.Decrypt(entry.Password, pwd)
			if err != nil {
				fmt.Printf("\nerror: %v\n", err)
				return
			}
			fmt.Println("")
		}

		scanner := bufio.NewScanner(os.Stdin)
		fmt.Print("Are you sure you want to proceed? [y/n]: ")

		scanner.Scan()
		text := scanner.Text()
		res := strings.ToLower(text)

		if strings.Contains(res, "y") || strings.Contains(res, "yes") {
			if err := db.DeleteEntry(title); err != nil {
				fmt.Println("error:", err)
			}

			fmt.Printf("\nSuccessfully deleted %s entry.", entry.Title)
		}
	},
}

func init() {
	RootCmd.AddCommand(deleteCmd)
}
