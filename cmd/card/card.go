package card

import (
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var errInvalidName = errors.New("error: invalid name")

var example = `
kure card (add/copy/ls/rm)`

// NewCmd returns a new command.
func NewCmd(db *bolt.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "card",
		Short:   "Card operations",
		Example: example,
	}

	cmd.AddCommand(addSubCmd(db, os.Stdin), copySubCmd(db), lsSubCmd(db), rmSubCmd(db, os.Stdin))

	return cmd
}
