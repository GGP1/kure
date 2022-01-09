package gen

import (
	"fmt"
	"math"

	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/commands/gen/phrase"

	"github.com/GGP1/atoll"

	"github.com/awnumar/memguard"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

const example = `
* Generate a random password
kure gen -l 18 -L 1,2,3 -i %&/ -e ? -r

* Generate and show the QR code image
kure gen -l 20 -q

* Generate, copy and mute standard output
kure gen -l 25 -cm`

type genOptions struct {
	include string
	exclude string
	levels  []int
	length  uint64
	qr      bool
	repeat  bool
	copy    bool
	mute    bool
}

// NewCmd returns a new command.
func NewCmd() *cobra.Command {
	opts := genOptions{}

	cmd := &cobra.Command{
		Use:   "gen",
		Short: "Generate a random password",
		Long: `Generate a random password.
		
Keyspace is the number of possible combinations of the password.
Average time taken to crack is based on a brute force attack scenario where the guesses per second is 1 trillion.`,
		Example: example,
		RunE:    runGen(&opts),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset variables (session)
			opts = genOptions{}
		},
	}

	cmd.AddCommand(phrase.NewCmd())

	f := cmd.Flags()
	f.BoolVarP(&opts.copy, "copy", "c", false, "copy the password to the clipboard")
	f.Uint64VarP(&opts.length, "length", "l", 0, "password length")
	f.IntSliceVarP(&opts.levels, "levels", "L", []int{1, 2, 3, 4, 5}, "password levels")
	f.StringVarP(&opts.include, "include", "i", "", "characters to include in the password")
	f.StringVarP(&opts.exclude, "exclude", "e", "", "characters to exclude from the password")
	f.BoolVarP(&opts.repeat, "repeat", "r", false, "allow character repetition")
	f.BoolVarP(&opts.qr, "qr", "q", false, "show the QR code image on the terminal")
	f.BoolVarP(&opts.mute, "mute", "m", false, "mute standard output when the password is copied")

	return cmd
}

func runGen(opts *genOptions) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		if opts.length < 1 || opts.length > math.MaxUint64 {
			return cmdutil.ErrInvalidLength
		}

		if len(opts.levels) == 0 {
			return errors.New("please specify levels")
		}

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
				return errors.Errorf("invalid level [%d]", lvl)
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
			return err
		}

		pwdBuf := memguard.NewBufferFromBytes([]byte(password))
		memguard.WipeBytes([]byte(password))
		defer pwdBuf.Destroy()

		if opts.qr {
			if err := cmdutil.DisplayQRCode(pwdBuf.String()); err != nil {
				return err
			}
		}

		entropy := p.Entropy()
		keyspace, timeToCrack := phrase.FormatSecretSecurity(atoll.Keyspace(p), atoll.SecondsToCrack(p)/2)

		if !opts.mute || !opts.copy {
			fmt.Printf(`Password: %s

Entropy: %.2f bits
Keyspace: %s
Average time taken to crack: %s
`, pwdBuf.String(), entropy, keyspace, timeToCrack)
		}

		if opts.copy {
			if err := cmdutil.WriteClipboard(cmd, 0, "Password", pwdBuf.String()); err != nil {
				return err
			}
		}
		return nil
	}
}
