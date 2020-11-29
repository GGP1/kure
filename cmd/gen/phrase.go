package gen

import (
	"fmt"
	"math"
	"strings"

	cmdutil "github.com/GGP1/kure/cmd"

	"github.com/GGP1/atoll"
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
		Use:   "phrase [-l length] [-s separator] [-i include] [-e exclude] [list] [-q qr]",
		Short: "Generate a random passphrase",
		Long: `Generate a random passphrase.
		
When using [-q qr] flag, make sure the terminal is bigger than the image or it will spoil.`,
		Aliases: []string{"p"},
		Example: phraseExample,
		RunE:    runPhrase(),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset flags defaults (session)
			length = 0
			separator, list = " ", "NoList"
			incl, excl = nil, nil
			qr = false
		},
	}

	f := cmd.Flags()
	f.Uint64VarP(&length, "length", "l", 0, "passphrase length")
	f.StringVarP(&separator, "separator", "s", " ", "set the character that separates each word")
	f.StringVar(&list, "list", "NoList", "choose passphrase generating method (NoList, WordList, SyllableList)")
	f.StringSliceVarP(&incl, "include", "i", nil, "characters to include in the passphrase")
	f.StringSliceVarP(&excl, "exclude", "e", nil, "characters to exclude from the passphrase")
	f.BoolVarP(&qr, "qr", "q", false, "create an image with the password QR code on the user home directory")

	return cmd
}

func runPhrase() cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		if length < 1 || length > math.MaxUint64 {
			return errInvalidLength
		}

		var l func(*atoll.Passphrase)
		var poolLength uint64

		if list != "" {
			list = strings.ReplaceAll(list, " ", "")
			list = strings.ToLower(list)

			// https://github.com/GGP1/atoll#entropy
			switch list {
			case "nolist", "no":
				l = atoll.NoList
				poolLength = uint64(math.Pow(26, float64(length)))
			case "wordlist", "word":
				l = atoll.WordList
				poolLength = 18325
			case "syllablelist", "syllable":
				l = atoll.SyllableList
				poolLength = 10129
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

		if qr {
			if err := cmdutil.DisplayQRCode(passphrase); err != nil {
				return err
			}
		}

		entropy := math.Log2(math.Pow(float64(poolLength), float64(length)))

		fmt.Printf("Passphrase: %s\nBits of entropy: %.2f\n", passphrase, entropy)
		return nil
	}
}
