package rm

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/GGP1/kure/auth"
	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/file"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var example = `
* Remove a file
kure file rm Sample

* Remove a directory
kure file rm SampleDir -d`

type rmOptions struct {
	dir bool
}

// NewCmd returns a new command.
func NewCmd(db *bolt.DB, r io.Reader) *cobra.Command {
	opts := rmOptions{}

	cmd := &cobra.Command{
		Use:     "rm <name>",
		Short:   "Remove files from the database",
		Example: example,
		Args:    cmdutil.MustExist(db, cmdutil.File),
		PreRunE: auth.Login(db),
		RunE:    runRm(db, r, &opts),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset variables (session)
			opts = rmOptions{}
		},
	}

	cmd.Flags().BoolVarP(&opts.dir, "dir", "d", false, "remove a directory and all the files stored in it")

	return cmd
}

func runRm(db *bolt.DB, r io.Reader, opts *rmOptions) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")
		ext := filepath.Ext(name)

		if opts.dir && ext != "" && ext != "." {
			return errors.Errorf("%q isn't a directory", name)
		}
		name = cmdutil.NormalizeName(name)

		if !cmdutil.Confirm(r, "Are you sure you want to proceed?") {
			return nil
		}

		// Remove single file
		if !opts.dir {
			if err := file.Remove(db, name); err != nil {
				return err
			}

			fmt.Printf("\n%q removed\n", name)
			return nil
		}

		// Remove directory
		fmt.Printf("Removing %q directory...\n", name)
		// Unlike the others objects' rm command, we use goroutines as a file folder may contain
		// subfolders and a lot of files in it
		files, err := file.ListNames(db)
		if err != nil {
			return err
		}

		for _, f := range files {
			// Make sure to delete items under that folder only
			if strings.HasPrefix(f, name+"/") {
				if err := file.Remove(db, f); err != nil {
					return err
				}

				fmt.Println("Remove:", f)
			}
		}

		return nil
	}
}
