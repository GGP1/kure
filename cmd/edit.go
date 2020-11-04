package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/GGP1/kure/db"
	"github.com/GGP1/kure/pb"

	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit <name> [-p password]",
	Short: "Edit an entry",
	Long: `Edit entry fields.
	
"-" = Clear field.
"" (nothing) = Do not modify field.`,
	Run: func(cmd *cobra.Command, args []string) {
		name := strings.Join(args, " ")

		oldEntry, err := db.GetEntry(name)
		if err != nil {
			fatal(err)
		}

		newEntry, err := editEntryInput(oldEntry)
		if err != nil {
			fatal(err)
		}

		if err := db.EditEntry(name, newEntry); err != nil {
			fatal(err)
		}

		fmt.Printf("\nSuccessfully edited \"%s\" entry.\n", name)
	},
}

func init() {
	rootCmd.AddCommand(editCmd)

	editCmd.Flags().BoolVarP(&password, "password", "p", false, "edit the entry password")
}

func editEntryInput(entry *pb.Entry) (*pb.Entry, error) {
	scanner := bufio.NewScanner(os.Stdin)

	var name, username, url, notes, expires string
	fmt.Println("Use '-' to clear a field and '' (nothing) to keep it unchanged")
	fmt.Print("\n")
	scan(scanner, "New name", &name)
	scan(scanner, "Username", &username)
	if password {
		var pwd string
		scan(scanner, "Password", &pwd)

		entry.Password = pwd
	}
	scan(scanner, "URL", &url)
	scanlns(scanner, "Notes", &notes)
	scan(scanner, "Expires", &expires)

	if expires == "0s" || expires == "0" || expires == "" {
		expires = "Never"
	} else {
		var (
			expTime time.Time
			err     error
		)

		expires = strings.ReplaceAll(expires, "-", "/")
		expTime, err = time.Parse("2006/01/02", expires)
		if err != nil {
			expTime, err = time.Parse("02/01/2006", expires)
			if err != nil {
				fatalf("invalid time format. Valid formats: d/m/y or y/m/d.")
			}
		}

		expires = expTime.Format(time.RFC1123Z)
	}

	if name == "-" {
		name = ""
	}
	if username == "-" {
		username = ""
	}
	if url == "-" {
		url = ""
	}
	if notes == "-" {
		notes = ""
	}

	entry.Name = name
	entry.Username = username
	entry.URL = url
	entry.Notes = notes
	entry.Expires = expires

	return entry, nil
}
