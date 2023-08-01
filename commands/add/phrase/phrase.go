package phrase

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"strings"

	"github.com/GGP1/kure/auth"
	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/entry"
	"github.com/GGP1/kure/pb"
	"github.com/GGP1/kure/terminal"

	"github.com/GGP1/atoll"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

const example = `
* Add an entry generating a random passphrase
kure add phrase Sample -l 6 -s $ -i atoll -e admin,login --list nolist`

type phraseOptions struct {
	list, separator string
	incl, excl      []string
	length          uint64
}

// NewCmd returns a new command.
func NewCmd(db *bolt.DB, r io.Reader) *cobra.Command {
	opts := phraseOptions{}
	cmd := &cobra.Command{
		Use:     "phrase <name>",
		Short:   "Create an entry using a passphrase",
		Aliases: []string{"passphrase"},
		Example: example,
		Args:    cmdutil.MustNotExist(db, cmdutil.Entry),
		PreRunE: auth.Login(db),
		RunE:    runPhrase(db, r, &opts),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset variables (session)
			opts = phraseOptions{
				separator: " ",
			}
		},
	}

	f := cmd.Flags()
	f.Uint64VarP(&opts.length, "length", "l", 0, "number of words")
	f.StringVarP(&opts.separator, "separator", "s", " ", "character that separates each word")
	f.StringSliceVarP(&opts.incl, "include", "i", nil, "words to include in the passphrase")
	f.StringSliceVarP(&opts.excl, "exclude", "e", nil, "words to exclude from the passphrase")
	f.StringVarP(&opts.list, "list", "L", "WordList", "passphrase list used {NoList|WordList|SyllableList}")

	return cmd
}

func runPhrase(db *bolt.DB, r io.Reader, opts *phraseOptions) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")
		name = cmdutil.NormalizeName(name)

		if opts.length < 1 || opts.length > math.MaxUint64 {
			return cmdutil.ErrInvalidLength
		}

		e, err := entryInput(r, name)
		if err != nil {
			return err
		}

		e.Password, err = genPassphrase(opts)
		if err != nil {
			return err
		}

		if err := entry.Create(db, e); err != nil {
			return err
		}

		fmt.Printf("\n%q added\n", name)
		return nil
	}
}

func entryInput(r io.Reader, name string) (*pb.Entry, error) {
	reader := bufio.NewReader(r)

	username := terminal.Scanln(reader, "Username")
	url := terminal.Scanln(reader, "URL")
	expires := terminal.Scanln(reader, "Expires [dd/mm/yy]")
	notes := terminal.Scanlns(reader, "Notes")

	exp, err := cmdutil.FmtExpires(expires)
	if err != nil {
		return nil, err
	}

	entry := &pb.Entry{
		Name:     name,
		Username: username,
		URL:      url,
		Expires:  exp,
		Notes:    notes,
	}
	return entry, nil
}

// genPassphrase returns a customized random passphrase.
func genPassphrase(opts *phraseOptions) (string, error) {
	l := atoll.WordList

	if opts.list != "" {
		opts.list = strings.ReplaceAll(opts.list, " ", "")

		switch strings.ToLower(opts.list) {
		case "nolist", "no":
			l = atoll.NoList

		case "wordlist", "word":
			// Do nothing as it's the default

		case "syllablelist", "syllable":
			l = atoll.SyllableList

		default:
			return "", errors.Errorf("invalid list: %q", opts.list)
		}
	}

	p := &atoll.Passphrase{
		Length:    opts.length,
		Separator: opts.separator,
		Include:   opts.incl,
		Exclude:   opts.excl,
		List:      l,
	}

	passphrase, err := atoll.NewSecret(p)
	if err != nil {
		return "", err
	}

	return string(passphrase), nil
}
