package rm

import (
	"fmt"
	"io"
	"strings"

	cmdutil "github.com/GGP1/kure/cmd"
	"github.com/GGP1/kure/db/entry"

	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var example = `
* Remove an entry
kure rm entryName`

// NewCmd returns a new command.
func NewCmd(db *bolt.DB, r io.Reader) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rm <name>",
		Short:   "Remove an entry",
		Example: example,
		RunE:    runRm(db, r),
	}

	return cmd
}

func runRm(db *bolt.DB, r io.Reader) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")

		if cmdutil.Proceed(r) {
			if err := entry.Remove(db, name); err != nil {
				return err
			}

			fmt.Printf("\nSuccessfully removed %q entry.\n", name)
		}
		return nil
	}
}
