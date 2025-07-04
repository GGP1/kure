package add

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/commands/add/phrase"
	"github.com/GGP1/kure/db/entry"
	"github.com/GGP1/kure/pb"
	"github.com/GGP1/kure/terminal"

	"github.com/GGP1/atoll"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

const example = `
* Add an entry using a custom password
kure add Sample -c

* Add an entry generating a random password
kure add Sample -l 27 -L 1,2,3,4,5 -i & -e / -r`

type addOptions struct {
	include, exclude string
	levels           []int
	length           uint64
	custom, repeat   bool
}

// NewCmd returns a new command.
func NewCmd(db *bolt.DB, r io.Reader) *cobra.Command {
	opts := addOptions{}
	cmd := &cobra.Command{
		Use:     "add <name>",
		Short:   "Add an entry",
		Aliases: []string{"create", "new"},
		Example: example,
		Args:    cmdutil.MustNotExist(db, cmdutil.Entry),
		RunE:    runAdd(db, r, &opts),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset variables (session)
			opts = addOptions{}
		},
	}

	cmd.AddCommand(phrase.NewCmd(db, r))

	f := cmd.Flags()
	f.BoolVarP(&opts.custom, "custom", "c", false, "use a custom password")
	f.Uint64VarP(&opts.length, "length", "l", 0, "password length")
	f.IntSliceVarP(&opts.levels, "levels", "L", []int{1, 2, 3, 4, 5}, "password levels")
	f.StringVarP(&opts.include, "include", "i", "", "characters to include in the password")
	f.StringVarP(&opts.exclude, "exclude", "e", "", "characters to exclude from the password")
	f.BoolVarP(&opts.repeat, "repeat", "r", false, "allow character repetition")

	return cmd
}

func runAdd(db *bolt.DB, r io.Reader, opts *addOptions) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")
		name = cmdutil.NormalizeName(name)

		if !opts.custom {
			if opts.length < 1 {
				return cmdutil.ErrInvalidLength
			}
			if len(opts.levels) == 0 {
				return errors.New("please specify levels")
			}
		}

		e, err := entryInput(r, name, opts.custom)
		if err != nil {
			return err
		}

		if !opts.custom {
			// Generate random password
			e.Password, err = genPassword(opts)
			if err != nil {
				return err
			}
		}

		if err := entry.Create(db, e); err != nil {
			return err
		}

		fmt.Printf("\n%q added\n", name)
		return nil
	}
}

// genPassword returns a customized random password or an error.
func genPassword(opts *addOptions) (string, error) {
	levels := make([]atoll.Level, len(opts.levels))
	for i, lvl := range opts.levels {
		switch lvl {
		case 1:
			levels[i] = atoll.Lower
		case 2:
			levels[i] = atoll.Upper
		case 3:
			levels[i] = atoll.Digit
		case 4:
			levels[i] = atoll.Space
		case 5:
			levels[i] = atoll.Special

		default:
			return "", errors.Errorf("invalid level [%d]", lvl)
		}
	}

	p := &atoll.Password{
		Length:  opts.length,
		Levels:  levels,
		Include: opts.include,
		Exclude: opts.exclude,
		Repeat:  opts.repeat,
	}

	password, err := atoll.NewSecret(p)
	if err != nil {
		return "", err
	}

	return string(password), nil
}

func entryInput(r io.Reader, name string, custom bool) (*pb.Entry, error) {
	var password string
	reader := bufio.NewReader(r)

	username := terminal.Scanln(reader, "Username")
	if custom {
		enclave, err := terminal.ScanPassword("Password", true)
		if err != nil {
			return nil, err
		}

		pwd, err := enclave.Open()
		if err != nil {
			return nil, errors.Wrap(err, "opening enclave")
		}

		password = pwd.String()
	}
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
		Password: password,
		URL:      url,
		Expires:  exp,
		Notes:    notes,
	}

	return entry, nil
}
