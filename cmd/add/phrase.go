package add

import (
	"fmt"
	"io"
	"math"
	"strings"

	"github.com/GGP1/atoll"
	cmdutil "github.com/GGP1/kure/cmd"
	"github.com/GGP1/kure/db/entry"
	"github.com/awnumar/memguard"

	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var (
	list, separator string
	incl, excl      []string
)

var phraseExample = `
* Add an entry generating a random passphrase
kure add phrase entryName -l 6 -s $ -i atoll -e admin,login --list wordlist`

func phraseSubCmd(db *bolt.DB, r io.Reader) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "phrase <name>",
		Short:   "Create an entry using a passphrase",
		Aliases: []string{"passphrase", "p"},
		Example: phraseExample,
		PreRunE: cmdutil.RequirePassword(db),
		RunE:    runPhrase(db, r),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset flags (session)
			list, separator = "", ""
			incl, excl = nil, nil
		},
	}

	f := cmd.Flags()
	f.Uint64VarP(&length, "length", "l", 1, "passphrase length")
	f.StringVarP(&separator, "separator", "s", " ", "character that separates each word")
	f.StringSliceVarP(&incl, "include", "i", nil, "list of words to include")
	f.StringSliceVarP(&excl, "exclude", "e", nil, "list of words to exclude")
	f.StringVar(&list, "list", "NoList", "passphrase generating method {NoList|WordList|SyllableList}")

	return cmd
}

func runPhrase(db *bolt.DB, r io.Reader) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")
		if name == "" {
			return errInvalidName
		}

		if length < 1 || length > math.MaxUint64 {
			return errInvalidLength
		}

		name = strings.TrimSpace(strings.ToLower(name))

		if err := cmdutil.Exists(db, name, "entry"); err != nil {
			return err
		}

		lockedBuf, e, err := input(db, r, name, false)
		if err != nil {
			return err
		}

		passphrase, err := genPassphrase(length, separator, list, incl, excl)
		if err != nil {
			return err
		}

		e.Password = passphrase
		memguard.WipeBytes([]byte(passphrase))

		if err := entry.Create(db, lockedBuf, e); err != nil {
			return err
		}

		fmt.Printf("\n%q added\n", name)
		return nil
	}
}

// genPassphrase returns a customized random passphrase.
func genPassphrase(length uint64, separator, list string, incl, excl []string) (string, error) {
	l := atoll.NoList

	if list != "" {
		list = strings.ReplaceAll(list, " ", "")
		list = strings.ToLower(list)

		switch list {
		case "nolist", "no":
			l = atoll.NoList
		case "wordlist", "word":
			l = atoll.WordList
		case "syllablelist", "syllable":
			l = atoll.SyllableList
		}
	}

	p := &atoll.Passphrase{
		Length:    length,
		Separator: separator,
		Include:   incl,
		Exclude:   excl,
		List:      l,
	}

	passphrase, err := atoll.NewSecret(p)
	if err != nil {
		return "", err
	}

	return passphrase, nil
}
