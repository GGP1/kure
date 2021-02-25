package rm

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/GGP1/kure/auth"
	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/entry"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var example = `
* Remove an entry
kure rm Sample

* Remove a directory
kure rm SampleDir -d`

type rmOptions struct {
	dir bool
}

// NewCmd returns a new command.
func NewCmd(db *bolt.DB, r io.Reader) *cobra.Command {
	opts := rmOptions{}

	cmd := &cobra.Command{
		Use:     "rm <name>",
		Short:   "Remove an entry or a directory",
		Example: example,
		Args:    cmdutil.MustExist(db, cmdutil.Entry),
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
			if err := entry.Remove(db, name); err != nil {
				return err
			}

			fmt.Printf("\n%q removed\n", name)
			return nil
		}

		fmt.Printf("Removing %q directory...\n", name)

		entries, err := entry.ListNames(db)
		if err != nil {
			return err
		}

		for _, e := range entries {
			// Make sure to delete items under that folder only by appending "/" to the name
			if strings.HasPrefix(e, name+"/") {
				if err := entry.Remove(db, e); err != nil {
					return err
				}

				fmt.Println("Remove:", e)
			}
		}

		return nil
	}
}
