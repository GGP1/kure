package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/GGP1/atoll"
	"github.com/GGP1/kure/db"
	"github.com/GGP1/kure/model/entry"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	custom, phrase bool
	length         uint64
	format         []int
)

var addCmd = &cobra.Command{
	Use:   "add <name> [-c custom] [-l length] [-f format] [-i include] [-e exclude] [-r repeat]",
	Short: "Add a new entry to the database",
	Run: func(cmd *cobra.Command, args []string) {
		name := strings.Join(args, " ")

		// If the user didn't specify format levels, use default
		if format == nil {
			if passFormat := viper.GetIntSlice("entry.format"); len(passFormat) > 0 {
				format = passFormat
			}
		}

		if entryRepeat := viper.GetBool("entry.repeat"); entryRepeat != false {
			repeat = entryRepeat
		}

		// Take entry input from the user and set name
		e, err := entryInput(name)
		if err != nil {
			fatal(err)
		}

		if custom {
			if err := db.CreateEntry(e); err != nil {
				fatal(err)
			}

			fmt.Println("\nSucessfully created the entry")
			return
		}

		password := &atoll.Password{
			Length:  length,
			Format:  format,
			Include: include,
		}

		if err := password.Generate(); err != nil {
			fatal(err)
		}

		e.Password = password.Secret

		if err := db.CreateEntry(e); err != nil {
			fatal(err)
		}

		fmt.Printf("\nSucessfully created the entry.\nBits of entropy: %.2f", password.Entropy)
	},
}

var addPhrase = &cobra.Command{
	Use: "phrase <name> [-l length] [-s separator] [-i include] [-e exclude] [list]",
	Run: func(cmd *cobra.Command, args []string) {
		name := strings.Join(args, " ")
		e, err := entryInput(name)
		if err != nil {
			fatal(err)
		}

		var f func(*atoll.Passphrase)
		m := strings.ReplaceAll(method, " ", "")
		list := strings.ToLower(m)

		switch list {
		case "nolist":
			f = atoll.NoList
		case "wordlist":
			f = atoll.WordList
		case "syllablelist":
			f = atoll.SyllableList
		}

		passphrase := &atoll.Passphrase{
			Length:    length,
			Separator: separator,
			Include:   incl,
			Exclude:   excl,
		}

		if err := passphrase.Generate(f); err != nil {
			fatal(err)
		}
		e.Password = passphrase.Secret

		if err := db.CreateEntry(e); err != nil {
			fatal(err)
		}

		fmt.Printf("\nSucessfully created the entry.\nBits of entropy: %.2f", passphrase.Entropy)
	},
}

func init() {
	rootCmd.AddCommand(addCmd)

	addCmd.AddCommand(addPhrase)
	addCmd.Flags().BoolVarP(&custom, "custom", "c", false, "use a custom password")
	addCmd.Flags().Uint64VarP(&length, "length", "l", 1, "password length")
	addCmd.Flags().IntSliceVarP(&format, "format", "f", nil, "password format")
	addCmd.Flags().StringVarP(&include, "include", "i", "", "characters to include in the password")
	addCmd.Flags().StringVarP(&exclude, "exclude", "e", "", "characters to exclude from the password")
	addCmd.Flags().BoolVarP(&repeat, "repeat", "r", true, "allow duplicated characters or not")

	addPhrase.Flags().Uint64VarP(&length, "length", "l", 1, "passphrase length")
	addPhrase.Flags().StringVarP(&separator, "separator", "s", " ", "set the character that separates each word")
	addPhrase.Flags().StringSliceVarP(&incl, "include", "i", nil, "list of words to include")
	addPhrase.Flags().StringSliceVarP(&excl, "exclude", "e", nil, "list of words to exclude")
	addPhrase.Flags().StringVar(&method, "list", "NoList", "choose passphrase generating method (NoList, WordList, SyllableList)")
}

func entryInput(name string) (*entry.Entry, error) {
	var username, password, url, notes, expires string

	if name == "" {
		return nil, errInvalidName
	}

	if len(name) > 43 {
		return nil, errors.New("entry name must contain 43 letters or less")
	}

	scanner := bufio.NewScanner(os.Stdin)

	scan(scanner, "Username", &username)
	if custom {
		scan(scanner, "Password", &password)
	}
	scan(scanner, "URL", &url)
	scan(scanner, "Notes", &notes)
	scan(scanner, "Expires", &expires)

	if expires == "0s" || expires == "0" || expires == "" {
		expires = "Never"
	} else {
		expTime, err := time.ParseDuration(expires)
		if err != nil {
			return nil, errors.Wrap(err, "expiration parse")
		}
		// Add duration and format
		expires = time.Now().Add(expTime).Format(time.RFC3339)
	}

	name = strings.TrimSpace(strings.ToLower(name))

	entry := entry.New(name, username, password, url, notes, expires)

	return entry, nil
}
