package file

import (
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var path string

var (
	errInvalidName = errors.New("error: invalid name")
	errInvalidPath = errors.New("error: invalid path")
)

var example = `
kure file (add/create/ls/rename/rm)`

// NewCmd returns a new command.
func NewCmd(db *bolt.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "file",
		Short: "File operations",
		Long: `Save the encrypted bytes of the file in the database, create, list or remove wherever you want. The user must specify the complete path, including the file type, for example: path/to/file/sample.png.
	
In case you don't specify where to list the file, it will be created in your current directory, for listing all the files kure will create a folder with the name "kure_files" in your directory and save all the files there.`,
		Example: example,
	}

	cmd.AddCommand(addSubCmd(db), lsSubCmd(db), renameSubCmd(db, os.Stdin), rmSubCmd(db, os.Stdin), touchSubCmd(db))

	return cmd
}
