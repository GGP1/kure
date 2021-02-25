package rm

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/GGP1/kure/auth"
	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/card"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var example = `
* Remove a card
kure card rm Sample

* Remove a directory
kure card rm SampleDir -d`

type rmOptions struct {
	dir bool
}

// NewCmd returns the a new command.
func NewCmd(db *bolt.DB, r io.Reader) *cobra.Command {
	opts := rmOptions{}

	cmd := &cobra.Command{
		Use:     "rm <name>",
		Short:   "Remove a card or directory",
		Example: example,
		Args:    cmdutil.MustExist(db, cmdutil.Card),
		PreRunE: auth.Login(db),
		RunE:    runRm(db, r, &opts),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset variables (session)
			opts = rmOptions{}
		},
	}

	cmd.Flags().BoolVarP(&opts.dir, "dir", "d", false, "remove a directory")

	return cmd
}

func runRm(db *bolt.DB, r io.Reader, opts *rmOptions) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")

		if opts.dir && filepath.Ext(name) != "" {
			return errors.Errorf("%q isn't a directory", name)
		}
		name = cmdutil.NormalizeName(name)

		if !cmdutil.Confirm(r, "Are you sure you want to proceed?") {
			return nil
		}

		// Remove single file
		if !opts.dir {
			if err := card.Remove(db, name); err != nil {
				return err
			}

			fmt.Printf("\n%q removed\n", name)
			return nil
		}

		fmt.Printf("Removing %q directory...\n", name)

		cards, err := card.ListNames(db)
		if err != nil {
			return err
		}

		for _, c := range cards {
			// Make sure to delete items under that folder only by appending "/" to the name
			if strings.HasPrefix(c, name+"/") {
				if err := card.Remove(db, c); err != nil {
					return err
				}
				fmt.Println("Remove:", c)
			}
		}

		return nil
	}
}
