package cmd

import (
	"fmt"
	"strings"

	"github.com/GGP1/kure/db"
	"github.com/GGP1/kure/model/entry"

	"github.com/spf13/cobra"
)

var hide bool

var listCmd = &cobra.Command{
	Use:   "list <title> [-H hide]",
	Short: "List entries",
	Run: func(cmd *cobra.Command, args []string) {
		title := strings.Join(args, " ")

		if title != "" {
			entry, err := db.GetEntry(title)
			if err != nil {
				must(err)
			}

			printEntry(entry)
			return
		}

		entries, err := db.ListEntries()
		if err != nil {
			must(err)
		}

		for _, entry := range entries {
			printEntry(entry)
		}
	},
}

func init() {
	RootCmd.AddCommand(listCmd)
	listCmd.Flags().BoolVarP(&hide, "hide", "H", false, "hide entries passwords")
}

func printEntry(e *entry.Entry) {
	password := e.Password
	title := strings.Title(e.Title)

	if hide {
		password = "•••••••••••••••"
	}

	str := fmt.Sprintf(
		`
+──────────────+───────────────────────────>
│ Title        │ %s
│ Username     │ %s
│ Password     │ %s
│ URL          │ %s
│ Notes        │ %s
│ Expires      │ %s
+──────────────+───────────────────────────>`,
		title, e.Username, password, e.URL, e.Notes, e.Expires)
	fmt.Println(str)
}
