package card

import (
	"fmt"
	"io"
	"strings"

	cmdutil "github.com/GGP1/kure/cmd"
	"github.com/GGP1/kure/db/card"
	"github.com/pkg/errors"

	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var dir bool

var rmExample = `
* Remove a card
kure card rm cardName

* Remove a directory
kure card rm dirName -d`

// rmSubCmd returns the copy subcommand
func rmSubCmd(db *bolt.DB, r io.Reader) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rm <name>",
		Short:   "Remove a card or directory",
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

		if err := cmdutil.Exists(db, name, "card"); err == nil {
			return errors.Errorf("%q does not exist", name)
		}

		if !cmdutil.Proceed(r) {
			return nil
		}

		// Remove single file
		if !dir {
			if err := card.Remove(db, name); err != nil {
				return err
			}

			fmt.Printf("\n%q deleted\n", name)
			return nil
		}

		fmt.Printf("Removing %q directory...\n", name)

		cards, err := card.ListNames(db)
		if err != nil {
			return err
		}

		// If the last rune of name is not a slash,
		// add it to make sure to delete items under that folder only
		if name[len(name)-1] != '/' {
			name += "/"
		}

		var list []string
		for _, c := range cards {
			if strings.HasPrefix(c, name) {
				list = append(list, c)
			}
		}

		for _, name := range list {
			fmt.Println("Remove:", name)

			if err := card.Remove(db, name); err != nil {
				return err
			}
		}

		return nil
	}
}
