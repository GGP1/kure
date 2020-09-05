package cmd

import (
	"fmt"
	"time"

	"github.com/GGP1/kure/db"

	"github.com/spf13/cobra"
)

var (
	hide    bool
	listCmd = &cobra.Command{
		Use:   "list",
		Short: "List entries.",
		Run: func(cmd *cobra.Command, args []string) {
			if title != "" {
				e, err := db.GetEntry(title)
				if err != nil {
					fmt.Println("Entry not found.")
					return
				}

				if hide {
					e.Password = ""
				}

				printResult(e.Title, e.Username, e.URL, e.Password, e.Expires)
				return
			}

			entries, err := db.ListEntries()
			if err != nil {
				fmt.Println(err)
			}

			for _, e := range entries {
				if hide {
					e.Password = ""
				}
				printResult(e.Title, e.Username, e.URL, e.Password, e.Expires)
			}
		},
	}
)

func init() {
	RootCmd.AddCommand(listCmd)
	listCmd.Flags().StringVarP(&title, "title", "t", "", "entry title")
	listCmd.Flags().BoolVarP(&hide, "hide", "H", false, "hide entries passwords")
}

func printResult(title, username, password, url string, expires time.Time) {
	s := fmt.Sprintf(
		`%s:                                                         
	Username: %s                                            
	URL: %s                                                 
	Password: %s                                               
	Expires: %v                                             
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`,
		title, username, url, password, expires)
	fmt.Println(s)
}
