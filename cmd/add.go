package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/GGP1/kure/db"
	"github.com/GGP1/kure/pb"

	"github.com/GGP1/atoll"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var addCmd = &cobra.Command{
	Use:   "add <name> [-c custom] [-l length] [-f format] [-i include] [-e exclude] [-r repeat]",
	Short: `Add a new entry to the database`,
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

		if length < 1 {
			fatalf("password length must be higher than 0")
		}

		entry, err := entryInput(name, custom)
		if err != nil {
			fatal(err)
		}

		message := fmt.Sprintf("\nSuccessfully created \"%s\" entry.", name)

		if !custom {
			password := &atoll.Password{
				Length:  length,
				Format:  format,
				Include: include,
			}

			if err := password.Generate(); err != nil {
				fatal(err)
			}

			entry.Password = password.Secret
			message = fmt.Sprintf("%s\nBits of entropy: %.2f", message, password.Entropy)
		}

		if err := db.CreateEntry(entry); err != nil {
			fatal(err)
		}

		fmt.Println(message)
	},
}

var addPhrase = &cobra.Command{
	Use: "phrase <name> [-l length] [-s separator] [-i include] [-e exclude] [list]",
	Run: func(cmd *cobra.Command, args []string) {
		name := strings.Join(args, " ")
		e, err := entryInput(name, custom)
		if err != nil {
			fatal(err)
		}

		var f func(*atoll.Passphrase)
		list = strings.ReplaceAll(list, " ", "")
		list = strings.ToLower(list)

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

		fmt.Printf("\nSuccessfully created \"%s\" entry.\nBits of entropy: %.2f\n", name, passphrase.Entropy)
	},
}

func init() {
	rootCmd.AddCommand(addCmd)

	addCmd.AddCommand(addPhrase)
	addCmd.Flags().BoolVarP(&custom, "custom", "c", false, "use a custom password")
	addCmd.Flags().Uint64VarP(&length, "length", "l", 0, "password length")
	addCmd.Flags().IntSliceVarP(&format, "format", "f", nil, "password format")
	addCmd.Flags().StringVarP(&include, "include", "i", "", "characters to include in the password")
	addCmd.Flags().StringVarP(&exclude, "exclude", "e", "", "characters to exclude from the password")
	addCmd.Flags().BoolVarP(&repeat, "repeat", "r", true, "allow duplicated characters or not")

	addPhrase.Flags().Uint64VarP(&length, "length", "l", 1, "passphrase length")
	addPhrase.Flags().StringVarP(&separator, "separator", "s", " ", "set the character that separates each word")
	addPhrase.Flags().StringSliceVarP(&incl, "include", "i", nil, "list of words to include")
	addPhrase.Flags().StringSliceVarP(&excl, "exclude", "e", nil, "list of words to exclude")
	addPhrase.Flags().StringVar(&list, "list", "NoList", "choose passphrase generating method (NoList, WordList, SyllableList)")
}

func entryInput(name string, pwd bool) (*pb.Entry, error) {
	if name == "" {
		return nil, errInvalidName
	}
	if !strings.Contains(name, "/") && len(name) > 43 {
		return nil, errors.New("entry name must contain 43 letters or less")
	}

	scanner := bufio.NewScanner(os.Stdin)

	var username, password, url, notes, expires string
	scan(scanner, "Username", &username)
	if pwd {
		scan(scanner, "Password", &password)
	}
	scan(scanner, "URL", &url)
	scanlns(scanner, "Notes", &notes)
	scan(scanner, "Expires", &expires)

	switch expires {
	case "0s", "0", "":
		expires = "Never"
	default:
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

	name = strings.ToLower(name)

	entry := &pb.Entry{
		Name:     name,
		Username: username,
		Password: password,
		URL:      url,
		Notes:    notes,
		Expires:  expires,
	}

	return entry, nil
}
