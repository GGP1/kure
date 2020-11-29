package file

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	cmdutil "github.com/GGP1/kure/cmd"
	"github.com/GGP1/kure/db/file"

	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var renameExample = `
* Rename a file
kure file rename fileName`

func renameSubCmd(db *bolt.DB, r io.Reader) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rename <name>",
		Short:   "Rename a file",
		Aliases: []string{"rn"},
		Example: renameExample,
		RunE:    runRename(db, r),
	}

	return cmd
}

func runRename(db *bolt.DB, r io.Reader) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		oldName := strings.Join(args, " ")

		scanner := bufio.NewScanner(r)
		newName := cmdutil.Scan(scanner, "New name")

		if oldName == "" || newName == "" {
			return errInvalidName
		}

		if err := file.Rename(db, oldName, newName); err != nil {
			return err
		}

		fmt.Printf("\nSuccessfully renamed %q as %q.\n", oldName, newName)
		return nil
	}
}
