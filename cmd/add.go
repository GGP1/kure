package cmd

import (
	"fmt"
	"strings"
	"syscall"
	"time"

	"github.com/GGP1/kure/crypt"
	"github.com/GGP1/kure/db"
	"github.com/GGP1/kure/entry"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	title      string
	username   string
	password   string
	url        []string
	expiration string
	length     uint16
	format     []uint
	secure     bool

	addCmd = &cobra.Command{
		Use:   "add",
		Short: "Adds a new entry to the database",
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
					fmt.Println("error:", err)
					return
				}
			}

			if secure {
				fmt.Print("Enter Password: ")
				pwd, err := terminal.ReadPassword(int(syscall.Stdin))
				if err != nil {
					fmt.Println("error: reading password:", err)
					return
				}

				encryptedPwd, err := crypt.Encrypt([]byte(password), pwd)
				if err != nil {
					fmt.Println("error:", err)
					return
				}
				password = string(encryptedPwd)
			}

			title, expiration, url, err := formatFields(title, expiration, url)
			if err != nil {
				fmt.Println("error:", err)
				return
			}

			entry := entry.New(title, username, password, url, expiration, secure)

			if err := db.CreateEntry(entry); err != nil {
				fmt.Println("error:", err)
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
	addCmd.Flags().StringSliceVarP(&url, "url", "U", []string{""}, "entry url")
	addCmd.Flags().StringVarP(&expiration, "expiration", "e", "0s", "entry expiration")
	addCmd.Flags().Uint16VarP(&length, "length", "l", 1, "password length")
	addCmd.Flags().UintSliceVarP(&format, "format", "f", []uint{1, 2, 3, 4}, "password format")
	addCmd.Flags().BoolVarP(&secure, "secure", "S", false, "security mode")

	addCmd.MarkFlagRequired("title")
	if password == "" {
		addCmd.MarkFlagRequired("length")
		addCmd.MarkFlagRequired("format")
	}
}

func formatFields(title, expiration string, url []string) (string, string, string, error) {
	t := strings.ToLower(title)

	if expiration == "0s" || expiration == "0" {
		expiration = "Never"
	} else {
		expTime, err := time.ParseDuration(expiration)
		if err != nil {
			return "", "", "", errors.Wrap(err, "duration parse")
		}
		// Add duration and format
		expiration = time.Now().Add(expTime).Format(time.RFC3339)
	}

	urls := strings.Join(url, ",")
	uri := strings.ReplaceAll(urls, ",", ", ")

	return t, expiration, uri, nil
}
