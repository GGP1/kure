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

var rmExample = `
* Remove a card
kure card rm cardName`

// rmSubCmd returns the copy subcommand
func rmSubCmd(db *bolt.DB, r io.Reader) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "rm <name>",
		Short:         "Remove a card from the database",
		Example:       rmExample,
		RunE:          runRm(db, r),
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	return cmd
}

func runRm(db *bolt.DB, r io.Reader) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")
		if name == "" {
			return errInvalidName
		}

		if cmdutil.Proceed(r) {
			if err := card.Remove(db, name); err != nil {
				return errors.Wrap(err, "error")
			}
		}

		fmt.Printf("\nSuccessfully removed %q card.\n", name)
		return nil
	}
}
