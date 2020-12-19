package gen

import (
	"fmt"
	"math"
	"strings"

	cmdutil "github.com/GGP1/kure/cmd"

	"github.com/GGP1/atoll"
	"github.com/atotto/clipboard"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	list, separator string
	incl, excl      []string
)

var phraseExample = `
* Generate a random passphrase and show QR code
kure gen phrase -l 7 -q`

func phraseSubCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "phrase",
		Short:   "Generate a random passphrase",
		Aliases: []string{"passphrase", "p"},
		Example: phraseExample,
		RunE:    runPhrase(),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset flags (session)
			length = 0
			separator, list = " ", "NoList"
			incl, excl = nil, nil
			qr = false
		},
	}

	f := cmd.Flags()
	f.BoolVarP(&copy, "copy", "c", false, "copy the passphrase to the clipboard")
	f.Uint64VarP(&length, "length", "l", 0, "passphrase length")
	f.StringVarP(&separator, "separator", "s", " ", "character that separates each word")
	f.StringVar(&list, "list", "NoList", "passphrase list used {NoList|WordList|SyllableList}")
	f.StringSliceVarP(&incl, "include", "i", nil, "characters to include in the passphrase")
	f.StringSliceVarP(&excl, "exclude", "e", nil, "characters to exclude from the passphrase")
	f.BoolVarP(&qr, "qr", "q", false, "show QR code on terminal")

	return cmd
}

func runPhrase() cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		if length < 1 || length > math.MaxUint64 {
			return errInvalidLength
		}

		// https://github.com/GGP1/atoll#entropy
		var poolLen uint64 = 18325
		l := atoll.WordList

		if list != "" {
			list = strings.ReplaceAll(list, " ", "")
			list = strings.ToLower(list)

			switch list {
			case "nolist", "no":
				l = atoll.NoList
				poolLen = uint64(math.Pow(26, float64(length)))
			case "wordlist", "word":
				l = atoll.WordList
			case "syllablelist", "syllable":
				l = atoll.SyllableList
				poolLen = 10129
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

		if copy {
			if err := clipboard.WriteAll(passphrase); err != nil {
				return errors.Wrap(err, "failed writing to the clipboard")
			}
		}

		if qr {
			if err := cmdutil.DisplayQRCode(passphrase); err != nil {
				return err
			}
		}

		entropy := math.Log2(math.Pow(float64(poolLen), float64(length)))

		fmt.Printf("Passphrase: %s\nBits of entropy: %.2f\n", passphrase, entropy)
		return nil
	}
}
