package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/GGP1/kure/db"
	"github.com/GGP1/kure/entry"

	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit <title>",
	Short: "Edit an entry",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		title := strings.Join(args, " ")

		oldEntry, err := db.GetEntry(title)
		if err != nil {
			fmt.Println("error:", err)
			return
		}

		username, url, notes, expiration := editEntryInput()

		title, expiration, err = formatFields(title, expiration)
		if err != nil {
			fmt.Println("error:", err)
			return
		}

		e := entry.New(title, username, string(oldEntry.Password), url, notes, expiration, oldEntry.Safe)

		err = db.EditEntry(e)
		if err != nil {
			fmt.Println("error:", err)
		}

		fmt.Printf("\nSuccessfully edited %s entry.", title)
	},
}

func init() {
	RootCmd.AddCommand(editCmd)
}

func editEntryInput() (username, url, notes, expiration string) {
	scanner := bufio.NewScanner(os.Stdin)

	username = scan(scanner, "Username", username)
	url = scan(scanner, "URL", url)
	notes = scan(scanner, "Notes", notes)
	expiration = scan(scanner, "Expiration", expiration)

	return username, url, notes, expiration
}
