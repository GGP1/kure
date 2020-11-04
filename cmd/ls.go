package cmd

import (
	"fmt"
	"strings"

	"github.com/GGP1/kure/db"
	"github.com/GGP1/kure/pb"
	"github.com/GGP1/kure/tree"

	"github.com/spf13/cobra"
)

var lsCmd = &cobra.Command{
	Use:   "ls <name> [-H hide] [-f filter]",
	Short: "List entries",
	Run: func(cmd *cobra.Command, args []string) {
		name := strings.Join(args, " ")

		if name == "" {
			entries, err := db.ListEntries()
			if err != nil {
				fatal(err)
			}

			paths := make([]string, len(entries))

			for i, entry := range entries {
				paths[i] = entry.Name
			}

			tree.Print(paths)
			return
		}

		if filter {
			entries, err := db.EntriesByName(name)
			if err != nil {
				fatal(err)
			}

			for _, entry := range entries {
				fmt.Print("\n")
				printEntry(entry)
			}
			return
		}

		entry, err := db.GetEntry(name)
		if err != nil {
			fatal(err)
		}

		if qr {
			if err := DisplayQRCode(entry.Password); err != nil {
				fatal(err)
			}
		}

		fmt.Print("\n")
		printEntry(entry)
	},
}

func init() {
	rootCmd.AddCommand(lsCmd)

	lsCmd.Flags().BoolVarP(&hide, "hide", "H", false, "hide entries passwords")
	lsCmd.Flags().BoolVarP(&qr, "qr", "q", false, "show the password QR code on the terminal (non-available when listing all entries)")
	lsCmd.Flags().BoolVarP(&filter, "filter", "f", false, "filter entries by name")
}

func printEntry(e *pb.Entry) {
	printObjectName(e.Name)

	if hide {
		e.Password = "•••••••••••••••"
	}

	fmt.Printf(`│ Username   │ %s
│ Password   │ %s
│ URL        │ %s
│ Expires    │ %s
`, e.Username, e.Password, e.URL, e.Expires)

	n := strings.Split(e.Notes, "\n")
	fmt.Printf("│ Notes      │ %s\n", n[0])

	for _, s := range n[1:] {
		fmt.Printf("│            │ %s\n", s)
	}

	fmt.Println("+────────────+──────────────────────────────>")
}
