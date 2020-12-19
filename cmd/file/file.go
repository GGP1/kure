package file

import (
	"io"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var path string

var (
	errInvalidName = errors.New("invalid name")
	errInvalidPath = errors.New("invalid path")
)

var example = `
kure file (add|cat|ls|rename|rm|touch)`

// NewCmd returns a new command.
func NewCmd(db *bolt.DB, r io.Reader, w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "file",
		Short:   "File operations",
		Example: example,
	}

	cmd.AddCommand(addSubCmd(db), catSubCmd(db, w), lsSubCmd(db), renameSubCmd(db, r), rmSubCmd(db, r), touchSubCmd(db))

	return cmd
}
