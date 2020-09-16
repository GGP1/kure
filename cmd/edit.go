package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/GGP1/kure/db"
	"github.com/GGP1/kure/entry"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit <title>",
	Short: "Edit an entry",
	Run: func(cmd *cobra.Command, args []string) {
		title := strings.Join(args, " ")

		oldEntry, err := db.GetEntry(title)
		if err != nil {
			log.Fatal("error: ", err)
		}

		username, url, notes, expiration, err := editEntryInput()
		if err != nil {
			log.Fatal("error: ", err)
		}

		e := entry.New(title, username, oldEntry.Password, url, notes, expiration)

		err = db.EditEntry(e)
		if err != nil {
			log.Fatal("error: ", err)
		}

		fmt.Printf("\nSuccessfully edited %s entry.", title)
	},
}

func init() {
	RootCmd.AddCommand(editCmd)
}

func editEntryInput() (username, url, notes, expiration string, err error) {
	scanner := bufio.NewScanner(os.Stdin)

	username = scan(scanner, "Username", username)
	url = scan(scanner, "URL", url)
	notes = scan(scanner, "Notes", notes)
	expiration = scan(scanner, "Expiration", expiration)

	if expiration == "0s" || expiration == "0" || expiration == "" {
		expiration = "Never"
	} else {
		expTime, err := time.ParseDuration(expiration)
		if err != nil {
			return "", "", "", "", errors.Wrap(err, "duration parse")
		}
		// Add duration and format
		expiration = time.Now().Add(expTime).Format(time.RFC3339)
	}

	return username, url, notes, expiration, nil
}
