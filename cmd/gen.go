package cmd

import (
	"fmt"

	"github.com/GGP1/kure/passgen"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	phrase             bool
	separator, include string
)

var genCmd = &cobra.Command{
	Use:   "gen [-l length] [-f format] [-p phrase] [-s separator] [-i include]",
	Short: "Generate a random password",
	Run: func(cmd *cobra.Command, args []string) {
		// If the user didn't specify format levels
		if format == nil {
			if passFormat := viper.GetIntSlice("entry.format"); len(passFormat) > 0 {
				format = passFormat
			}
		}

		if phrase {
			passphrase := &passgen.Passphrase{
				Length:    length,
				Separator: separator,
			}

			// Passphrase generate always returns nil
			pass, _ := passphrase.Generate()
			entropy := passphrase.Entropy()

			fmt.Printf("Passphrase: %s\nBits of entropy: %.2f\n", pass, entropy)
			return
		}

		password := &passgen.Password{
			Length:  length,
			Format:  format,
			Include: include,
		}

		pass, err := password.Generate()
		if err != nil {
			must(err)
		}

		entropy := password.Entropy()

		fmt.Printf("Password: %s\nBits of entropy: %.2f\n", pass, entropy)
	},
}

func init() {
	RootCmd.AddCommand(genCmd)
	genCmd.Flags().Uint64VarP(&length, "length", "l", 1, "password length")
	genCmd.Flags().IntSliceVarP(&format, "format", "f", nil, "password format")
	genCmd.Flags().BoolVarP(&phrase, "phrase", "p", false, "generate a passphrase")
	genCmd.Flags().StringVarP(&separator, "separator", "s", " ", "set the character that separates each word")
	genCmd.Flags().StringVarP(&include, "include", "i", "", "characters to include in the password")

	if err := genCmd.MarkFlagRequired("length"); err != nil {
		must(err)
	}
}
