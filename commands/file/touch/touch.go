package touch

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/file"
	"github.com/GGP1/kure/pb"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

const example = `
* Create a file and overwrite if it already exists
kure file touch fileName -p path/to/folder -o

* Create multiple files in the current directory
kure file touch file1 file2 file3

* Create all the files (includes folders and subfolders)
kure file touch -p path/to/folder`

type touchOptions struct {
	path      string
	overwrite bool
}

// NewCmd returns a new command.
func NewCmd(db *bolt.DB) *cobra.Command {
	opts := touchOptions{}
	cmd := &cobra.Command{
		Use:   "touch <name>",
		Short: "Create stored files",
		Long: `Create one, multiple, all the files or an specific directory.

For creating an specific file the extension must be included in the arguments, if not, kure will consider that the user is trying to create a directory and will create all the files in it.

In case any of the paths contains spaces within it, it must be enclosed by double quotes.

In case a path is passed, kure will create any missing folders for you.`,
		Aliases: []string{"th"},
		Example: example,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return nil
			}
			return cmdutil.MustExist(db, cmdutil.File, true)(cmd, args)
		},
		RunE: runTouch(db, &opts),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset variables (session)
			opts = touchOptions{}
		},
	}

	f := cmd.Flags()
	f.BoolVarP(&opts.overwrite, "overwrite", "o", false, "overwrite files if they already exist")
	f.StringVarP(&opts.path, "path", "p", "", "destination folder path")

	return cmd
}

func runTouch(db *bolt.DB, opts *touchOptions) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		absolute, err := filepath.Abs(opts.path)
		if err != nil {
			return cmdutil.ErrInvalidPath
		}
		opts.path = absolute

		if err := os.MkdirAll(opts.path, 0o700); err != nil {
			return errors.Wrap(err, "making directory")
		}

		if err := os.Chdir(opts.path); err != nil {
			return errors.Wrap(err, "changing directory")
		}

		// Create all
		if len(args) == 0 {
			files, err := file.List(db)
			if err != nil {
				return err
			}

			fmt.Println("Creating files at", opts.path)

			for _, f := range files {
				// Log errors, do not return
				if err := createFiles(f, opts.path, opts.overwrite); err != nil {
					fmt.Fprintln(os.Stderr, "error:", err)
				}
			}

			return nil
		}

		// Create one or more, files or directories
		// Log errors to avoid stopping the whole operation
		fmt.Println("Creating files at", opts.path)
		for _, name := range args {
			name = strings.ToLower(strings.TrimSpace(name))

			// Assume the user wants to recreate an entire directory
			if strings.HasSuffix(name, "/") {
				if err := createDirectory(db, name, opts.path, opts.overwrite); err != nil {
					fmt.Fprintln(os.Stderr, "error:", err)
				}
				continue
			}

			// Create single file
			f, err := file.Get(db, name)
			if err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				continue
			}

			if err := createFile(f, opts.overwrite); err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				continue
			}
		}

		return nil
	}
}

func createDirectory(db *bolt.DB, name, path string, overwrite bool) error {
	files, err := file.List(db)
	if err != nil {
		return err
	}

	// If the last rune of name is not a slash,
	// add it to make sure to create items under that folder only
	if name[len(name)-1] != '/' {
		name += "/"
	}

	var dir []*pb.File
	for _, f := range files {
		if strings.HasPrefix(f.Name, name) {
			dir = append(dir, f)
		}
	}

	if len(dir) == 0 {
		return errors.Errorf("there is no folder named %q", name)
	}

	for _, f := range dir {
		if err := createFiles(f, path, overwrite); err != nil {
			return err
		}
	}

	return nil
}

func createFile(file *pb.File, overwrite bool) error {
	filename := filepath.Base(file.Name)

	// Create if it doesn't exist or if we are allowed to overwrite it
	if _, err := os.Stat(filename); os.IsExist(err) && !overwrite {
		return errors.Errorf("%q already exists, use -o to overwrite files", filename)
	}

	if err := os.WriteFile(filename, file.Content, 0o600); err != nil {
		return errors.Wrapf(err, "writing %q", filename)
	}
	fmt.Println("Create:", file.Name)
	return nil
}

// createFiles takes care of recreating folders and files as they were stored in the database.
//
// This function works synchronously only, running it concurrently messes up os.Chdir().
//
// The path is used only to return to the root folder.
func createFiles(file *pb.File, path string, overwrite bool) error {
	// "the shire/frodo/ring.png" would be [the shire, frodo, ring.png]
	parts := strings.Split(file.Name, "/")

	for i, p := range parts {
		// If it's the last element, create the file
		if i == len(parts)-1 {
			if err := createFile(file, overwrite); err != nil {
				return err
			}
			// Go back to the root folder
			if err := os.Chdir(path); err != nil {
				return errors.Wrap(err, "changing to root directory")
			}
			break
		}

		// If it's not the last element, create a folder
		// Use MkdirAll to avoid errors when the folder already exists
		if err := os.MkdirAll(p, 0o700); err != nil {
			return errors.Wrapf(err, "making %q directory", p)
		}
		if err := os.Chdir(p); err != nil {
			return errors.Wrapf(err, "changing to %q directory", p)
		}
	}

	return nil
}
