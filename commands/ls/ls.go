package ls

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/entry"
	"github.com/GGP1/kure/orderedmap"
	"github.com/GGP1/kure/pb"
	"github.com/GGP1/kure/terminal"
	"github.com/GGP1/kure/tree"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

const example = `
* List one and show sensitive information
kure ls Sample -s

* List one and show the password QR code
kure ls Sample -q

* Filter by name
kure ls Sample -f 

* List all
kure ls`

type lsOptions struct {
	filter, qr, show bool
}

// NewCmd returns a new command.
func NewCmd(db *bolt.DB) *cobra.Command {
	opts := lsOptions{}
	cmd := &cobra.Command{
		Use:   "ls <name>",
		Short: "List entries",
		Long: `List entries.

Listing all the entries does not check for expired entries, this decision was taken to prevent high loads when the number of entries is elevated. Listing a single entry does notifies if it is expired.`,
		Aliases: []string{"entries", "list"},
		Example: example,
		Args:    cmdutil.MustExistLs(db, cmdutil.Entry),
		RunE:    runLs(db, &opts),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset variables (session)
			opts = lsOptions{}
		},
	}

	f := cmd.Flags()
	f.BoolVarP(&opts.filter, "filter", "f", false, "filter by name")
	f.BoolVarP(&opts.qr, "qr", "q", false, "show the password QR code on the terminal")
	f.BoolVarP(&opts.show, "show", "s", false, "show entry password")

	return cmd
}

func runLs(db *bolt.DB, opts *lsOptions) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")
		name = cmdutil.NormalizeName(name)

		// List all
		if name == "" {
			entries, err := entry.ListNames(db)
			if err != nil {
				return err
			}

			tree.Print(entries)
			return nil
		}

		// Filter by name
		if opts.filter {
			entries, err := entry.ListNames(db)
			if err != nil {
				return err
			}

			var matches []string
			for _, entry := range entries {
				matched, err := filepath.Match(name, entry)
				if err != nil {
					return err
				}

				if matched {
					matches = append(matches, entry)
				}
			}

			if len(matches) == 0 {
				return errors.New("no entries were found")
			}

			tree.Print(matches)
			return nil
		}

		// List one
		e, err := entry.Get(db, name)
		if err != nil {
			return err
		}

		if opts.qr {
			if err := terminal.DisplayQRCode(e.Password); err != nil {
				return err
			}
		}

		printEntry(name, e, opts.show)
		return nil
	}
}

func printEntry(name string, e *pb.Entry, show bool) {
	if !show {
		e.Password = "•••••••••••••••"
	}

	if expired(e.Expires) {
		e.Expires = "EXPIRED"
	}

	mp := orderedmap.New()
	mp.Set("Username", e.Username)
	mp.Set("Password", e.Password)
	mp.Set("URL", e.URL)
	mp.Set("Expires", e.Expires)
	mp.Set("Notes", e.Notes)

	fmt.Println(cmdutil.BuildBox(name, mp))
}

// expired returns if the entry is expired or not.
func expired(expires string) bool {
	if expires == "Never" {
		return false
	}

	// Error is always nil as "expires" field was already formatted before being saved
	expiration, _ := time.Parse(time.RFC1123Z, expires)
	return time.Now().After(expiration)
}
