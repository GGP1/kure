package add

import (
	"fmt"
	"io"
	"math"
	"strings"

	"github.com/GGP1/atoll"
	cmdutil "github.com/GGP1/kure/cmd"
	"github.com/GGP1/kure/db/entry"

	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var (
	list, separator string
	incl, excl      []string
)

var phraseExample = `
* Add an entry generating a customized random passphrase
kure add phrase -l 6 -s $ -i atoll -e admin,login --list wordlist`

// phraseSubCmd returns the phrase subcommand.
func phraseSubCmd(db *bolt.DB, r io.Reader) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "phrase <name> [-l length] [-s separator] [-i include] [-e exclude] [list]",
		Short:   "Add an entry using a passphrase",
		Aliases: []string{"p"},
		Example: phraseExample,
		RunE:    runPhrase(db, r),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset flags defaults (session)
			list, separator = "", ""
			incl, excl = nil, nil
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	f := cmd.Flags()
	f.Uint64VarP(&length, "length", "l", 1, "passphrase length")
	f.StringVarP(&separator, "separator", "s", " ", "set the character that separates each word")
	f.StringSliceVarP(&incl, "include", "i", nil, "list of words to include")
	f.StringSliceVarP(&excl, "exclude", "e", nil, "list of words to exclude")
	f.StringVar(&list, "list", "NoList", "choose passphrase generating method (NoList, WordList, SyllableList)")

	return cmd
}

func runPhrase(db *bolt.DB, r io.Reader) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")

		if length < 1 || length > math.MaxUint64 {
			return errInvalidLength
		}

		e, err := entryInput(db, name, custom, r)
		if err != nil {
			return err
		}

		var l func(*atoll.Passphrase)

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
			return err
		}

		e.Password = passphrase

		if err := entry.Create(db, e); err != nil {
			return err
		}

		fmt.Printf("\nSuccessfully created %q entry.", name)
		return nil
	}
}
