package note

import (
	"fmt"
	"io"
	"strings"

	cmdutil "github.com/GGP1/kure/cmd"
	"github.com/GGP1/kure/db/note"
	"github.com/pkg/errors"

	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var dir bool

var rmExample = `
* Remove a note
kure note rm noteName

* Remove a directory
kure card rm dirName -d`

// rmSubCmd returns the copy subcommand
func rmSubCmd(db *bolt.DB, r io.Reader) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rm <name>",
		Short:   "Remove a note or directory",
		Example: rmExample,
		PreRunE: cmdutil.RequirePassword(db),
		RunE:    runRm(db, r),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset flags (session)
			dir = false
		},
	}

	cmd.Flags().BoolVarP(&dir, "dir", "d", false, "remove a directory")

	return cmd
}

func runRm(db *bolt.DB, r io.Reader) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")
		if name == "" {
			return errInvalidName
		}

		if err := cmdutil.Exists(db, name, "note"); err == nil {
			return errors.Errorf("%q does not exist", name)
		}

		if !cmdutil.Proceed(r) {
			return nil
		}

		// Remove single file
		if !dir {
			if err := note.Remove(db, name); err != nil {
				return err
			}

			fmt.Printf("\n%q deleted\n", name)
			return nil
		}

		fmt.Printf("Removing %q directory...\n", name)

		notes, err := note.ListNames(db)
		if err != nil {
			return err
		}

		// If the last rune of name is not a slash,
		// add it to make sure to not delete items under that folder only
		if name[len(name)-1] != '/' {
			name += "/"
		}

		var list []string
		for _, n := range notes {
			if strings.HasPrefix(n, name) {
				list = append(list, n)
			}
		}

		for _, name := range list {
			fmt.Println("Remove:", name)

			if err := note.Remove(db, name); err != nil {
				return err
			}
		}

		return nil
	}
}
