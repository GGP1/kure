package rm

import (
	"fmt"
	"io"
	"strings"

	"github.com/GGP1/kure/auth"
	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/file"

	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

const example = `
* Remove a file
kure file rm Sample

* Remove a directory
kure file rm SampleDir/`

// NewCmd returns a new command.
func NewCmd(db *bolt.DB, r io.Reader) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rm <name>",
		Short:   "Remove a file or directory",
		Example: example,
		Args:    cmdutil.MustExist(db, cmdutil.File, true),
		PreRunE: auth.Login(db),
		RunE:    runRm(db, r),
	}

	return cmd
}

func runRm(db *bolt.DB, r io.Reader) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")
		name = cmdutil.NormalizeName(name, true)

		if !cmdutil.Confirm(r, "Are you sure you want to proceed?") {
			return nil
		}

		// Remove single file
		if !strings.HasSuffix(name, "/") {
			fmt.Println("Remove:", name)
			if err := file.Remove(db, name); err != nil {
				return err
			}

			return nil
		}

		// Remove directory
		fmt.Printf("Removing %q directory...\n", name)
		files, err := file.ListNames(db)
		if err != nil {
			return err
		}

		selected := make([]string, 0)
		for _, f := range files {
			if strings.HasPrefix(f, name) {
				selected = append(selected, f)
				fmt.Println("Remove:", f)
			}
		}

		return file.Remove(db, selected...)
	}
}
