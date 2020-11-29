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
		Use:   "ls <name> [-f filter] [-H hide] [-q qr]",
		Short: "List entries",
		Long: `List entries.
		
When using [-q qr] flag, make sure the terminal is bigger than the image or it will spoil.`,
		Example: example,
		RunE:    runLs(db),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset flags defaults (session)
			filter, hide, qr = false, false, false
		},
		SilenceErrors: true,
		SilenceUsage:  true,
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

		if name == "" {
			entries, err := entry.ListNames(db)
			if err != nil {
				return errors.Wrap(err, "error")
			}

			paths := make([]string, len(entries))

			for i, entry := range entries {
				paths[i] = entry.Name
			}

			tree.Print(paths)
			return nil
		}

		if filter {
			entries, err := entry.ListByName(db, name)
			if err != nil {
				return errors.Wrap(err, "error")
			}

			if len(entries) == 0 {
				return errors.New("error: no wallets were found")
			}

			for _, entry := range entries {
				printEntry(entry)
			}
			return nil
		}

		e, err := entry.Get(db, name)
		if err != nil {
			return errors.Wrap(err, "error")
		}

		if qr {
			if err := cmdutil.DisplayQRCode(e.Password); err != nil {
				return errors.Wrap(err, "error")
			}
			fmt.Print("\n") // used to avoid messing up the entry printing
		}

		printEntry(e)
		return nil
	}
}

func printEntry(e *pb.Entry) {
	cmdutil.PrintObjectName(e.Name)

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

	// If notes occupies more than one line, insert rows
	for _, s := range n[1:] {
		fmt.Printf("│            │ %s\n", s)
	}

	fmt.Println("└────────────+────────────────────────────────────────────>")
}
