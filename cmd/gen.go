package cmd

import (
	"fmt"
	"strings"

	"github.com/GGP1/kure/img"

	"github.com/GGP1/atoll"
	"github.com/pkg/errors"
	"github.com/skip2/go-qrcode"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var genCmd = &cobra.Command{
	Use:   "gen [-l length] [-f format] [-i include] [-e exclude] [-r repeat] [-q qr]",
	Short: "Generate a random password",
	Run: func(cmd *cobra.Command, args []string) {
		if length < 1 || length > 18446744073709551615 {
			fatal(errInvalidLength)
		}

		if format == nil {
			if passFormat := viper.GetIntSlice("entry.format"); len(passFormat) > 0 {
				format = passFormat
			} else {
				fatalf("please specify a format")
			}
		}

		if entryRepeat := viper.GetBool("entry.repeat"); entryRepeat != false {
			repeat = entryRepeat
		}

		p := &atoll.Password{
			Length:  length,
			Format:  format,
			Include: include,
			Exclude: exclude,
			Repeat:  repeat,
		}

		if err := p.Generate(); err != nil {
			fatal(err)
		}

		if qr {
			if err := DisplayQRCode(p.Secret); err != nil {
				fatal(err)
			}
		}

		fmt.Printf("Password: %s\nBits of entropy: %.2f\n", p.Secret, p.Entropy)
	},
}

var genPhrase = &cobra.Command{
	Use:   "phrase [-l length] [-s separator] [-i include] [-e exclude] [list] [-q qr]",
	Short: "Generate a random passphrase",
	Run: func(cmd *cobra.Command, args []string) {
		if length < 1 || length > 18446744073709551615 {
			fatal(errInvalidLength)
		}

		f := atoll.NoList

		if list != "" {
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
		}

		p := &atoll.Passphrase{
			Length:    length,
			Separator: separator,
			Include:   incl,
			Exclude:   excl,
		}

		if err := p.Generate(f); err != nil {
			fatal(err)
		}

		if qr {
			if err := DisplayQRCode(p.Secret); err != nil {
				fatal(err)
			}
		}

		fmt.Printf("Passphrase: %s\nBits of entropy: %.2f\n", p.Secret, p.Entropy)
	},
}

func init() {
	rootCmd.AddCommand(genCmd)

	genCmd.AddCommand(genPhrase)
	genCmd.Flags().Uint64VarP(&length, "length", "l", 0, "password length")
	genCmd.Flags().IntSliceVarP(&format, "format", "f", nil, "password format")
	genCmd.Flags().StringVarP(&include, "include", "i", "", "characters to include in the password")
	genCmd.Flags().StringVarP(&exclude, "exclude", "e", "", "characters to exclude from the password")
	genCmd.Flags().BoolVarP(&repeat, "repeat", "r", false, "allow duplicated characters or not (default false)")
	genCmd.Flags().BoolVarP(&qr, "qr", "q", false, "create an image with the password QR code on the user home directory")

	genPhrase.Flags().Uint64VarP(&length, "length", "l", 0, "passphrase length")
	genPhrase.Flags().StringVarP(&separator, "separator", "s", " ", "set the character that separates each word")
	genPhrase.Flags().StringSliceVarP(&incl, "include", "i", nil, "characters to include in the passphrase")
	genPhrase.Flags().StringSliceVarP(&excl, "exclude", "e", nil, "characters to exclude from the passphrase")
	genPhrase.Flags().StringVar(&list, "list", "NoList", "choose passphrase generating method (NoList, WordList, SyllableList)")
	genPhrase.Flags().BoolVarP(&qr, "qr", "q", false, "create an image with the password QR code on the user home directory")
}

// DisplayQRCode creates a qr code with the password provided.
func DisplayQRCode(secret string) error {
	qr, err := qrcode.Encode(secret, qrcode.Highest, 1)
	if err != nil {
		return errors.Wrap(err, "creating QR code")
	}

	if err := img.Display(secret, qr); err != nil {
		return err
	}

	return nil
}
