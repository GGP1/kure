package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/GGP1/kure/db"
	"github.com/GGP1/kure/model/entry"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit <name>",
	Short: "Edit an entry",
	Run: func(cmd *cobra.Command, args []string) {
		name := strings.Join(args, " ")

		oldEntry, err := db.GetEntry(name)
		if err != nil {
			fatal(err)
		}

		username, url, notes, expiration, err := editEntryInput()
		if err != nil {
			fatal(err)
		}

		e := entry.New(name, username, oldEntry.Password, url, notes, expiration)

		if err := db.EditEntry(e); err != nil {
			fatal(err)
		}

		fmt.Printf("\nSuccessfully edited %s entry.", name)
	},
}

func init() {
	rootCmd.AddCommand(editCmd)
}

func editEntryInput() (username, url, notes, expiration string, err error) {
	scanner := bufio.NewScanner(os.Stdin)

	scan(scanner, "Username", &username)
	scan(scanner, "URL", &url)
	scan(scanner, "Notes", &notes)
	scan(scanner, "Expiration", &expiration)

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
