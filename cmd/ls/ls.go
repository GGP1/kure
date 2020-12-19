package ls

import (
	"fmt"
	"strings"

	cmdutil "github.com/GGP1/kure/cmd"
	"github.com/GGP1/kure/db/entry"
	"github.com/GGP1/kure/pb"
	"github.com/GGP1/kure/tree"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var filter, hide, qr bool

var example = `
* List all
kure ls

* Filter by name
kure ls entryName -f 

* List one and hide sensible information (optional)
kure ls entryName -H

* List one and show the password QR code
kure ls entryName -q`

// NewCmd returns a new command.
func NewCmd(db *bolt.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ls <name>",
		Short: "List entries",
		Long: `List entries.
		
When using [-q qr] flag, make sure the terminal is bigger than the image or it will spoil.`,
		Aliases: []string{"entries"},
		Example: example,
		PreRunE: cmdutil.RequirePassword(db),
		RunE:    runLs(db),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset flags (session)
			filter, hide, qr = false, false, false
		},
	}

	f := cmd.Flags()
	f.BoolVarP(&hide, "hide", "H", false, "hide entries passwords")
	f.BoolVarP(&qr, "qr", "q", false, "show the password QR code on the terminal (non-available when listing all entries)")
	f.BoolVarP(&filter, "filter", "f", false, "filter entries by name")

	return cmd
}

func runLs(db *bolt.DB) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")

		switch name {
		case "":
			entries, err := entry.ListNames(db)
			if err != nil {
				return err
			}
			tree.Print(entries)

		default:
			if filter {
				entries, err := entry.ListNames(db)
				if err != nil {
					return err
				}

				var list []string
				for _, entry := range entries {
					if strings.Contains(entry, name) {
						list = append(list, entry)
					}
				}

				if len(list) == 0 {
					return errors.New("no entries were found")
				}

				tree.Print(list)
				break
			}

			e, err := entry.Get(db, name)
			if err != nil {
				return err
			}

			if qr {
				if err := cmdutil.DisplayQRCode(e.Password); err != nil {
					return err
				}
				fmt.Print("\n") // used to avoid messing up the entry print
			}

			printEntry(e)
		}

		return nil
	}
}

func printEntry(e *pb.Entry) {
	if hide {
		e.Password = "•••••••••••••••"
	}

	fields := map[string]string{
		"Username": e.Username,
		"Password": e.Password,
		"URL":      e.URL,
		"Expires":  e.Expires,
		"Notes":    e.Notes,
	}

	box := cmdutil.BuildBox(e.Name, fields)
	fmt.Println("\n" + box)
}
