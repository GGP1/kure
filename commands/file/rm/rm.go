package rm

import (
	"fmt"
	"io"
	"strings"

	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/file"
	"github.com/GGP1/kure/terminal"

	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

const example = `
* Remove a file
kure file rm Sample

* Remove a directory
kure file rm SampleDir/

* Remove multiple files
kure file rm Sample Sample2 Sample3`

// NewCmd returns a new command.
func NewCmd(db *bolt.DB, r io.Reader) *cobra.Command {
	return &cobra.Command{
		Use:     "rm <names>",
		Short:   "Remove files or directories",
		Example: example,
		Args:    cmdutil.MustExist(db, cmdutil.File, true),
		RunE:    runRm(db, r),
	}
}

func runRm(db *bolt.DB, r io.Reader) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		if !terminal.Confirm(r, "Are you sure you want to proceed?") {
			return nil
		}

		names := make([]string, 0, len(args))
		for _, name := range args {
			name = cmdutil.NormalizeName(name, true)

			if !strings.HasSuffix(name, "/") {
				names = append(names, name)
				fmt.Println("Remove:", name)
				continue
			}

			files, err := file.ListNames(db)
			if err != nil {
				return err
			}

			for _, f := range files {
				if strings.HasPrefix(f, name) {
					names = append(names, f)
					fmt.Println("Remove:", f)
				}
			}
		}

		return file.Remove(db, names...)
	}
}
