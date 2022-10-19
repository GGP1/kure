package rm

import (
	"fmt"
	"io"
	"strings"

	"github.com/GGP1/kure/auth"
	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/entry"
	"github.com/GGP1/kure/terminal"

	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

const example = `
* Remove an entry
kure rm Sample

* Remove a directory
kure rm SampleDir/`

// NewCmd returns a new command.
func NewCmd(db *bolt.DB, r io.Reader) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rm <name>",
		Short:   "Remove an entry or a directory",
		Example: example,
		Args:    cmdutil.MustExist(db, cmdutil.Entry, true),
		PreRunE: auth.Login(db),
		RunE:    runRm(db, r),
	}

	return cmd
}

func runRm(db *bolt.DB, r io.Reader) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")
		name = cmdutil.NormalizeName(name, true)

		if !terminal.Confirm(r, "Are you sure you want to proceed?") {
			return nil
		}

		// Remove single file
		if !strings.HasSuffix(name, "/") {
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
			if strings.HasPrefix(e, name) {
				if err := entry.Remove(db, e); err != nil {
					return err
				}

				fmt.Println("Remove:", e)
			}
		}

		return nil
	}
}
