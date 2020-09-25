package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/GGP1/kure/db"
	"github.com/GGP1/kure/model/entry"
	"github.com/GGP1/kure/passgen"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	custom bool
	length uint64
	format []int
)

var addCmd = &cobra.Command{
	Use:   "add [-c custom | -p phrase] [-s separator] [-l length] [-f format] [-i include]",
	Short: "Adds a new entry to the database",
	Run: func(cmd *cobra.Command, args []string) {
		var entropy float64
		var err error

		// If the user didn't specify format levels, use default
		if format == nil {
			if passFormat := viper.GetIntSlice("entry.format"); len(passFormat) > 0 {
				format = passFormat
			}
		}

		// Take entry input from the user
		e, err := entryInput()
		if err != nil {
			must(err)
		}

		// If the user does not use a custom password.
		if !custom {
			if phrase {
				passphrase := &passgen.Passphrase{
					Length:    length,
					Separator: separator,
				}

				// Passphrase generate and entropy always return nil
				e.Password, _ = passphrase.Generate()
				entropy = passphrase.Entropy()
			} else {
				password := &passgen.Password{
					Length:  length,
					Format:  format,
					Include: include,
				}

				e.Password, err = password.Generate()
				if err != nil {
					must(err)
				}

				entropy = password.Entropy()
			}
		}

		if err := db.CreateEntry(e); err != nil {
			must(err)
		}

		fmt.Printf("\nSucessfully created the entry.\nBits of entropy: %.2f", entropy)
	},
}

func init() {
	RootCmd.AddCommand(addCmd)
	addCmd.Flags().BoolVarP(&custom, "custom", "c", false, "use a custom password")
	addCmd.Flags().BoolVarP(&phrase, "phrase", "p", false, "generate a passphrase")
	addCmd.Flags().StringVarP(&separator, "separator", "s", " ", "set the character that separates each word")
	addCmd.Flags().Uint64VarP(&length, "length", "l", 1, "password length")
	addCmd.Flags().IntSliceVarP(&format, "format", "f", nil, "password format")
	addCmd.Flags().StringVarP(&include, "include", "i", "", "characters to include in the password")
}

func entryInput() (*entry.Entry, error) {
	var title, username, password, url, notes, expires string

	scanner := bufio.NewScanner(os.Stdin)

	title = scan(scanner, "Title", title)
	username = scan(scanner, "Username", username)
	url = scan(scanner, "URL", url)
	notes = scan(scanner, "Notes", notes)
	expires = scan(scanner, "Expires", expires)

	if custom {
		password = scan(scanner, "Password", password)
	}

	if expires == "0s" || expires == "0" || expires == "" {
		expires = "Never"
	} else {
		expTime, err := time.ParseDuration(expires)
		if err != nil {
			return nil, errors.Wrap(err, "duration parse")
		}
		// Add duration and format
		expires = time.Now().Add(expTime).Format(time.RFC3339)
	}

	title = strings.ToLower(title)

	entry := entry.New(title, username, password, url, notes, expires)

	return entry, nil
}
