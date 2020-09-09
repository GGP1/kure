package cmd

import (
	"fmt"

	"github.com/GGP1/kure/entry"

	"github.com/spf13/cobra"
)

var (
	phrase    bool
	separator string
	include   string
	genCmd    = &cobra.Command{
		Use:   "gen [-l length] [-f format] [-p phrase] [-s separator] [-i include]",
		Short: "Generate a random password",
		Run: func(cmd *cobra.Command, args []string) {
			if phrase {
				passphrase, entropy := entry.GeneratePassphrase(int(length), separator)
				fmt.Printf("Passphrase: %s\nBits of entropy: %.2f\n", passphrase, entropy)
				return
			}

			levels := make(map[uint]struct{})

			for _, v := range format {
				levels[v] = struct{}{}
			}

			password, entropy, err := entry.GeneratePassword(length, levels, include)
			if err != nil {
				fmt.Println("error:", err)
				return
			}

			fmt.Printf("Password: %s\nBits of entropy: %.2f\n", password, entropy)
		},
	}
)

func init() {
	RootCmd.AddCommand(genCmd)
	genCmd.Flags().Uint16VarP(&length, "length", "l", 1, "password length")
	genCmd.Flags().UintSliceVarP(&format, "format", "f", []uint{1, 2, 3}, "password format")
	genCmd.Flags().BoolVarP(&phrase, "phrase", "p", false, "generate a passphrase")
	genCmd.Flags().StringVarP(&separator, "separator", "s", " ", "set the character that separates each word")
	genCmd.Flags().StringVarP(&include, "include", "i", "", "characters to include in pool of the password")
}
