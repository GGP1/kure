package cmd

import (
	"fmt"
	"time"

	"github.com/GGP1/kure/db"
	"github.com/GGP1/kure/entry"

	"github.com/spf13/cobra"
)

var (
	title      string
	username   string
	password   string
	url        string
	notes      []string
	expiration string
	length     uint16
	format     []uint

	addCmd = &cobra.Command{
		Use:   "add",
		Short: "Adds a new entry to the database.",
		Run: func(cmd *cobra.Command, args []string) {
			var entropy float64
			levels := make(map[uint]struct{})

			for _, v := range format {
				levels[v] = struct{}{}
			}

			if password == "" {
				var err error
				password, entropy, err = entry.GeneratePassword(length, levels)
				if err != nil {
					fmt.Println(err)
					return
				}
			}

			// Parse time and add it to time.Now
			expTime, err := time.ParseDuration(expiration)
			if err != nil {
				fmt.Println(err)
				return
			}
			exp := time.Now().Add(expTime)

			entry := entry.New(title, username, url, password, exp)

			err = db.CreateEntry(entry)
			if err != nil {
				fmt.Println(err)
				return
			}

			fmt.Printf("Sucessfully created the entry.\nBits of entropy: %.2f", entropy)
		},
	}
)

func init() {
	RootCmd.AddCommand(addCmd)
	addCmd.Flags().StringVarP(&title, "title", "t", "", "entry title")
	addCmd.Flags().StringVarP(&username, "username", "u", "", "entry username")
	addCmd.Flags().StringVarP(&password, "password", "p", "", "custom password")
	addCmd.Flags().StringVarP(&url, "url", "U", "", "entry url")
	addCmd.Flags().StringVarP(&expiration, "expiration", "e", "0s", "entry expiration")
	addCmd.Flags().Uint16VarP(&length, "length", "l", 1, "password length")
	addCmd.Flags().UintSliceVarP(&format, "format", "f", []uint{1, 2, 3}, "password format")
}
