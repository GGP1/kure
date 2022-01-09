package mv

import (
	"fmt"
	"path/filepath"

	"github.com/GGP1/kure/auth"
	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/file"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

const example = `
* Move a file
kure file mv oldSample newSample`

// NewCmd returns a new command.
func NewCmd(db *bolt.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mv <src> <dst>",
		Short: "Rename/move a file",
		Long: `Rename/move a file.

Renaming a directory is not allowed. In case any of the paths contains spaces within it, it must be enclosed by double quotes.`,
		Args:    cobra.ExactArgs(2),
		Example: example,
		PreRunE: auth.Login(db),
		RunE:    runMv(db),
	}

	return cmd
}

func runMv(db *bolt.DB) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		oldName := args[0]
		newName := args[1]
		if oldName == "" || newName == "" {
			return errors.New("invalid format, use: kure file mv <oldName> <newName>")
		}

		oldName = cmdutil.NormalizeName(oldName)
		newName = cmdutil.NormalizeName(newName)
		if filepath.Ext(newName) == "" {
			newName += filepath.Ext(oldName)
		}

		if err := cmdutil.Exists(db, newName, cmdutil.File); err != nil {
			return err
		}

		oldFile, err := file.Get(db, oldName)
		if err != nil {
			return err
		}

		if err := file.Remove(db, oldName); err != nil {
			return err
		}

		oldFile.Name = newName
		if err := file.Create(db, oldFile); err != nil {
			return err
		}

		fmt.Printf("\n%q renamed as %q\n", oldName, newName)
		return nil
	}
}
