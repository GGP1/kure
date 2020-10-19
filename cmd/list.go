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
	Use:   "list <name> [-H hide]",
	Short: "List an entry or all the entries",
	Run: func(cmd *cobra.Command, args []string) {
		name := strings.Join(args, " ")

		if name != "" {
			entry, err := db.GetEntry(name)
			if err != nil {
				fatal(err)
			}

			if qr {
				if err := QRCode(entry.Password); err != nil {
					fatal(err)
				}
			}

			fmt.Print("\n")
			printEntry(entry)
			return
		}

		entries, err := db.ListEntries()
		if err != nil {
			fatal(err)
		}

		for _, entry := range entries {
			fmt.Print("\n")
			printEntry(entry)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().BoolVarP(&hide, "hide", "H", false, "hide entries passwords")
	listCmd.Flags().BoolVarP(&qr, "qr", "q", false, "create an image with the password QR code on the user home directory (non-available when listing all entries)")
}

func printEntry(e *entry.Entry) {
	e.Name = strings.Title(e.Name)

	if hide {
		e.Password = "•••••••••••••••"
	}

	dashes := 43
	halfBar := ((dashes - len(e.Name)) / 2) - 1

	fmt.Print("+")
	for i := 0; i < halfBar; i++ {
		fmt.Print("─")
	}
	fmt.Printf(" %s ", e.Name)

	if (len(e.Name) % 2) == 0 {
		halfBar++
	}

	for i := 0; i < halfBar; i++ {
		fmt.Print("─")
	}
	fmt.Print(">\n")

	if e.Username != "" {
		fmt.Printf("│ Username   │ %s\n", e.Username)
	}

	if e.Password != "" {
		fmt.Printf("│ Password   │ %s\n", e.Password)
	}

	if e.URL != "" {
		fmt.Printf("│ URL        │ %s\n", e.URL)
	}

	if e.Notes != "" {
		fmt.Printf("│ Notes      │ %s\n", e.Notes)
	}

	if e.Expires != "" {
		fmt.Printf("│ Expires    │ %s\n", e.Expires)
	}

	fmt.Println("+────────────+──────────────────────────────>")
}
