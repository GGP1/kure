package mv

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/GGP1/kure/auth"
	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/file"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

const example = `
* Move a file
kure file mv oldFile newFile

* Move a directory
kure file mv oldDir/ newDir/`

// NewCmd returns a new command.
func NewCmd(db *bolt.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mv <src> <dst>",
		Short: "Move a file or directory",
		Long: `Move a file or directory.

In case any of the paths contains spaces within it, it must be enclosed by double quotes.`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return errors.Errorf("accepts 2 arg(s), received %d", len(args))
			}

			oldName := cmdutil.NormalizeName(args[0], true)
			if err := cmdutil.Exists(db, oldName, cmdutil.File); err == nil {
				return errors.Errorf("there's no file nor directory named %q", strings.TrimSuffix(oldName, "/"))
			}
			return nil
		},
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

		oldName = cmdutil.NormalizeName(oldName, true)
		newName = cmdutil.NormalizeName(newName, true)

		if strings.HasSuffix(oldName, "/") {
			if !strings.HasSuffix(newName, "/") {
				return errors.New("cannot move a directory into a file")
			}
			return mvDir(db, oldName, newName)
		}

		if filepath.Ext(newName) == "" {
			newName += filepath.Ext(oldName)
		}

		if err := cmdutil.Exists(db, newName, cmdutil.File); err != nil {
			return err
		}

		if err := file.Rename(db, oldName, newName); err != nil {
			return err
		}

		fmt.Printf("\n%q renamed as %q\n", oldName, newName)
		return nil
	}
}

func mvDir(db *bolt.DB, oldName, newName string) error {
	names, err := file.ListNames(db)
	if err != nil {
		return err
	}

	fmt.Printf("Moving %q directory into %q...\n", strings.TrimSuffix(oldName, "/"), strings.TrimSuffix(newName, "/"))

	for _, name := range names {
		if strings.HasPrefix(name, oldName) {
			if err := file.Rename(db, name, newName+strings.TrimPrefix(name, oldName)); err != nil {
				return err
			}
		}
	}

	return nil
}
