package file

import (
	"fmt"
	"io"
	"path/filepath"

	cmdutil "github.com/GGP1/kure/cmd"
	"github.com/GGP1/kure/db/file"
	"github.com/pkg/errors"

	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var renameExample = `
* Rename a file
kure file rename oldName newName`

func renameSubCmd(db *bolt.DB, r io.Reader) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rename <old-name> <new-name>",
		Short: "Rename a file",
		Long: `Rename a file.

In case any of the paths contains spaces within it, it must be enclosed by double quotes.`,
		Aliases: []string{"rn"},
		Args:    cobra.ExactValidArgs(2),
		Example: renameExample,
		PreRunE: cmdutil.RequirePassword(db),
		RunE:    runRename(db),
	}

	return cmd
}

func runRename(db *bolt.DB) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		oldName := args[0]
		newName := args[1]

		if oldName == "" || newName == "" {
			return errors.New("invalid format, use: kure file rename <oldName> <newName>")
		}

		if filepath.Ext(newName) == "" {
			newName += filepath.Ext(oldName)
		}

		if err := file.Rename(db, oldName, newName); err != nil {
			return err
		}

		fmt.Printf("\n%q renamed as %q\n", oldName, newName)
		return nil
	}
}
