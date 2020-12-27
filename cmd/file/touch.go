package file

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	cmdutil "github.com/GGP1/kure/cmd"
	"github.com/GGP1/kure/db/file"
	"github.com/GGP1/kure/pb"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var (
	overwrite bool
)

var createExample = `
* Create a file and overwrite if it already exists
kure file touch fileName -p path/to/folder -o

* Create multiple files in the current directory
kure file touch file1 file2 file3

* Create all the files (includes folders and subfolders)
kure file touch -p path/to/folder`

func touchSubCmd(db *bolt.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "touch <name>",
		Short: "Create stored files",
		Long: `Create one, multiple, all the files or an specific directory.

For creating an specific file the extension must be included in the arguments, if not, Kure will consider that the user is trying to create a directory and will create all the files in it.

In case any of the paths contains spaces within it, it must be enclosed by double quotes.

In case a path is passed, Kure will create any missing folders for you.`,
		Aliases: []string{"t", "th"},
		Example: createExample,
		PreRunE: cmdutil.RequirePassword(db),
		RunE:    runTouch(db),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset flags (session)
			overwrite = false
			path = ""
		},
	}

	f := cmd.Flags()
	f.BoolVarP(&overwrite, "overwrite", "o", false, "overwrite files if they already exist")
	f.StringVarP(&path, "path", "p", "", "destination folder path")

	return cmd
}

func runTouch(db *bolt.DB) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		absolute, err := filepath.Abs(path)
		if err != nil {
			return errInvalidPath
		}
		path = absolute

		if err := os.MkdirAll(path, os.ModeDir); err != nil {
			return errors.Wrap(err, "failed making directory")
		}

		if err := os.Chdir(path); err != nil {
			return errors.Wrap(err, "failed changing directory")
		}

		// Create all
		if len(args) == 0 {
			filesBuf, files, err := file.List(db)
			if err != nil {
				return err
			}
			defer filesBuf.Destroy()

			fmt.Printf("Creating files at %s\n", path)

			for _, f := range files {
				// Log errors, do not return
				if err := createFiles(f, path); err != nil {
					fmt.Fprintln(os.Stderr, "error:", err)
				}
			}

			return nil
		}

		// Create one or more, files or directories
		fmt.Println("Creating files at", path)
		for _, name := range args {
			name = strings.ToLower(strings.TrimSpace(name))

			// Assume the user wants to recreate an entire directory
			if filepath.Ext(name) == "" {
				filesBuf, files, err := file.List(db)
				if err != nil {
					fmt.Fprintln(os.Stderr, "error:", err)
					continue
				}

				// If the last rune of name is not a slash,
				// add it to make sure to create items under that folder only
				if name[len(name)-1] != '/' {
					name += "/"
				}

				dirBuf, dir := pb.SecureFileSlice()

				for _, f := range files {
					if strings.Contains(f.Name, name) {
						dir = append(dir, f)
					}
				}
				filesBuf.Destroy()

				if len(dir) == 0 {
					fmt.Fprintf(os.Stderr, "error: there is no folder named %q\n", name)
					dirBuf.Destroy()
					continue
				}

				for _, f := range dir {
					if err := createFiles(f, path); err != nil {
						fmt.Fprintln(os.Stderr, "error:", err)
					}
				}

				dirBuf.Destroy()
				continue
			}

			// Create single file
			fileBuf, f, err := file.Get(db, name)
			if err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				continue
			}

			if err := createFile(f, overwrite); err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				continue
			}
			fileBuf.Destroy()
		}

		return nil
	}
}

// createFile creates a file with the filename and content provided.
func createFile(file *pb.File, overwrite bool) error {
	filename := filepath.Base(file.Name)

	// Create if it doesn't exist or if we are allowed to overwrite it
	_, err := os.Stat(filename)
	if os.IsNotExist(err) || overwrite {
		if err := ioutil.WriteFile(filename, file.Content, 0644); err != nil {
			return errors.Wrapf(err, "failed writing %q", filename)
		}
		fmt.Println("Create:", file.Name)
		return nil
	}

	return errors.Errorf("%q already exists, use -o to overwrite already existing files\n", filename)
}

// createFiles takes care of recreating folders and files as they were stored in the database.
//
// This function works synchronously only, running it concurrently will mess up os.Chdir().
//
// The path is used only to return to the root folder.
func createFiles(file *pb.File, path string) error {
	// "the shire/frodo/ring.png" would be [the shire, frodo, ring.png]
	parts := strings.Split(file.Name, "/")

	for i, p := range parts {
		// If it's the last element, create the file
		if i == len(parts)-1 {
			if err := createFile(file, overwrite); err != nil {
				return err
			}
			// Go back to the root folder
			os.Chdir(path)
			break
		}

		// If it's not the last element, create a folder
		if err := os.MkdirAll(p, os.ModeDir); err != nil {
			return errors.Wrap(err, "failed making inner directory")
		}
		os.Chdir(p)
	}

	return nil
}
