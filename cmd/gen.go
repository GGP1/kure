package cmd

import (
	"fmt"
	"strings"

	"github.com/GGP1/atoll"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	repeat                              bool
	separator, include, exclude, method string
	incl, excl                          []string
)

var genCmd = &cobra.Command{
	Use:   "gen [-l length] [-f format] [-i include] [-e exclude] [-r repeat]",
	Short: "Generate a random password",
	Run: func(cmd *cobra.Command, args []string) {
		if format == nil {
			if passFormat := viper.GetIntSlice("entry.format"); len(passFormat) > 0 {
				format = passFormat
			} else {
				fatal(errors.New("please specify a format"))
			}
		}

		if entryRepeat := viper.GetBool("entry.repeat"); entryRepeat != false {
			repeat = entryRepeat
		}

		password := &atoll.Password{
			Length:  length,
			Format:  format,
			Include: include,
			Exclude: exclude,
			Repeat:  repeat,
		}

		if err := password.Generate(); err != nil {
			fatal(err)
		}

		fmt.Printf("Password: %s\nBits of entropy: %.2f\n", password.Secret, password.Entropy)
	},
}

var genPhrase = &cobra.Command{
	Use:   "phrase [-l length] [-s separator] [-i include] [-e exclude] [list]",
	Short: "Generate a random passphrase",
	Run: func(cmd *cobra.Command, args []string) {
		var f func(*atoll.Passphrase)

		if method != "" {
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

		fmt.Printf("Passphrase: %s\nBits of entropy: %.2f\n", passphrase.Secret, passphrase.Entropy)
	},
}

func init() {
	rootCmd.AddCommand(genCmd)

	genCmd.AddCommand(genPhrase)
	genCmd.Flags().Uint64VarP(&length, "length", "l", 1, "password length")
	genCmd.Flags().IntSliceVarP(&format, "format", "f", nil, "password format")
	genCmd.Flags().StringVarP(&include, "include", "i", "", "characters to include in the password")
	genCmd.Flags().StringVarP(&exclude, "exclude", "e", "", "characters to exclude from the password")
	genCmd.Flags().BoolVarP(&repeat, "repeat", "r", false, "allow duplicated characters or not (default false)")

	genPhrase.Flags().Uint64VarP(&length, "length", "l", 1, "passphrase length")
	genPhrase.Flags().StringVarP(&separator, "separator", "s", " ", "set the character that separates each word")
	genPhrase.Flags().StringSliceVarP(&incl, "include", "i", nil, "characters to include in the passphrase")
	genPhrase.Flags().StringSliceVarP(&excl, "exclude", "e", nil, "characters to exclude from the passphrase")
	genPhrase.Flags().StringVar(&method, "list", "NoList", "choose passphrase generating method (NoList, WordList, SyllableList)")

	genCmd.MarkFlagRequired("length")
}
