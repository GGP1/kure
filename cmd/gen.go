package cmd

import (
	"fmt"

	"github.com/GGP1/kure/entry"

	"github.com/spf13/cobra"
)

var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate a random password.",
	Run: func(cmd *cobra.Command, args []string) {
		levels := make(map[uint]struct{})

		for _, v := range format {
			levels[v] = struct{}{}
		}

		password, entropy, err := entry.GeneratePassword(length, levels)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Printf("Password: %s\nBits of entropy: %.2f", password, entropy)
	},
}

func init() {
	RootCmd.AddCommand(genCmd)
	genCmd.Flags().Uint16VarP(&length, "length", "l", 1, "password length")
	genCmd.Flags().UintSliceVarP(&format, "format", "f", []uint{1, 2, 3}, "password format")
}
