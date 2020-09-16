package cmd

import (
	"fmt"
	"log"

	"github.com/GGP1/kure/entry"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	phrase    bool
	separator string
	include   string
)

var genCmd = &cobra.Command{
	Use:   "gen [-l length] [-f format] [-p phrase] [-s separator] [-i include]",
	Short: "Generate a random password",
	Run: func(cmd *cobra.Command, args []string) {
		if passFormat := viper.GetIntSlice("entry.format"); len(passFormat) > 0 {
			format = passFormat
		}

		if phrase {
			passphrase, entropy := entry.GeneratePassphrase(int(length), separator)
			fmt.Printf("Passphrase: %s\nBits of entropy: %.2f\n", passphrase, entropy)
			return
		}

		password, entropy, err := entry.GeneratePassword(length, format, include)
		if err != nil {
			log.Fatal("error: ", err)
		}

		fmt.Printf("Password: %s\nBits of entropy: %.2f\n", password, entropy)
	},
}

func init() {
	RootCmd.AddCommand(genCmd)
	genCmd.Flags().Uint16VarP(&length, "length", "l", 1, "password length")
	genCmd.Flags().IntSliceVarP(&format, "format", "f", []int{1, 2, 3, 4}, "password format")
	genCmd.Flags().BoolVarP(&phrase, "phrase", "p", false, "generate a passphrase")
	genCmd.Flags().StringVarP(&separator, "separator", "s", " ", "set the character that separates each word")
	genCmd.Flags().StringVarP(&include, "include", "i", "", "characters to include in pool of the password")
}
