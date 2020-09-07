package cmd

import (
	"fmt"
	"strings"

	"github.com/GGP1/kure/crypt"
	"github.com/GGP1/kure/db"
	"github.com/GGP1/kure/entry"

	"github.com/spf13/cobra"
)

var (
	hide    bool
	listCmd = &cobra.Command{
		Use:   "list [-h hide] [-S secure]",
		Short: "List entries",
		Run: func(cmd *cobra.Command, args []string) {
			title := strings.Join(args, " ")

			if title != "" {
				entry, err := db.GetEntry(title)
				if err != nil {
					fmt.Println("error: entry not found")
					return
				}

				if secure && entry.Secure {
					pwd, err := passInput()
					if err != nil {
						fmt.Println("error:", err)
						return
					}

					decryptedPwd, err := crypt.Decrypt(entry.Password, pwd)
					if err != nil {
						fmt.Printf("\nerror: %v\n", err)
						return
					}

					entry.Password = decryptedPwd
				}

				printResult(entry)
				return
			}

			entries, err := db.ListEntries()
			if err != nil {
				fmt.Println("error:", err)
			}

			for _, e := range entries {
				printResult(e)
			}
		},
	}
)

func init() {
	RootCmd.AddCommand(listCmd)
	listCmd.Flags().BoolVarP(&hide, "hide", "H", false, "hide entries passwords")
	listCmd.Flags().BoolVarP(&secure, "secure", "S", false, "decrypt password before listing")
}

func printResult(e *entry.Entry) {
	password := string(e.Password)
	if hide {
		password = ""
	}

	// If secure flag is false and the password is encrypted, set encrypted label.
	// This is used because encrypted text makes the log messy.
	if !secure {
		if e.Secure {
			password = "- encrypted password -"
		}
	}

	t := strings.Title(string(e.Title))

	s := fmt.Sprintf(
		`
%s:
	Username: %s
	Password: %s
	     URL: %s
	   Notes: %s
	 Expires: %s
	 
▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬`,
		t, e.Username, password, e.URL, e.Notes, e.Expires)
	fmt.Println(s)
}
