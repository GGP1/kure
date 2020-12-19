package note

import (
	"errors"
	"io"

	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var errInvalidName = errors.New("invalid name")

var example = `
kure note {add|copy|ls|rm}`

// NewCmd returns a new command.
func NewCmd(db *bolt.DB, r io.Reader) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "note",
		Short:   "Note operations",
		Example: example,
	}

	cmd.AddCommand(addSubCmd(db, r), copySubCmd(db), lsSubCmd(db), rmSubCmd(db, r))

	return cmd
}
