package file

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	cmdutil "github.com/GGP1/kure/cmd"
	"github.com/GGP1/kure/db/file"
	"github.com/GGP1/kure/pb"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var (
	overwrite bool
	multiple  []string
)

var createExample = `
* Create a file and overwrite if it already exists
kure touch fileName -p path/to/folder -o

* Create multiple files
kure touch -t file1,file2,file3 -p path/to/folder

* Create all the files (includes folders and subfolders)
kure touch -p path/to/folder`

func touchSubCmd(db *bolt.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "touch <name> [-m multiple] [-o overwrite] [-p path]",
		Short:   "Create stored files",
		Long:    "Create one, multiple or all the files from the database. In case a path is passed, kure will create any missing folders for you",
		Aliases: []string{"t", "th"},
		Example: createExample,
		RunE:    runTouch(db),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset flags defaults (session)
			multiple = nil
			overwrite = false
			path = ""
		},
	}

	f := cmd.Flags()
	f.StringSliceVarP(&multiple, "multiple", "m", nil, "a list of the files to create")
	f.BoolVarP(&overwrite, "overwrite", "o", false, "overwrite files if they already exist")
	f.StringVarP(&path, "path", "p", "", "destination path (where the file will be created)")

	return cmd
}

func runTouch(db *bolt.DB) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		var wg sync.WaitGroup
		name := strings.Join(args, " ")

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

		// Single file
		if name != "" {
			f, err := file.Get(db, name)
			if err != nil {
				return err
			}

			if err := createFile(f, overwrite); err != nil {
				return err
			}

			fmt.Printf("Created %q on path %s\n", f.Name, path)
			return nil
		}

		// Multiple files
		if len(multiple) > 0 {
			wg.Add(len(multiple))
			for _, m := range multiple {
				// Log errors, do not exit
				go func() {
					defer wg.Done()

					m = strings.ToLower(m)
					f, err := file.Get(db, m)
					if err != nil {
						fmt.Println("error:", err)
					}

					if err := createFile(f, overwrite); err != nil {
						fmt.Println("error:", err)
					}
				}()
				wg.Wait()
			}

			fmt.Printf("Created files on path %s\n", path)
			return nil
		}

		// All files
		files, err := file.List(db)
		if err != nil {
			return err
		}

		fmt.Printf("Creating files into %s\n", path)

		wg.Add(len(files))
		for _, f := range files {
			go func(f *pb.File, path string) {
				defer wg.Done()
				if err := createFiles(f, path); err != nil {
					fmt.Println("error:", err)
				}
			}(f, path)
		}
		wg.Wait()

		return nil
	}
}

// createFile creates a file with the filename and content provided.
func createFile(file *pb.File, overwrite bool) error {
	filename := file.Name

	// If the file name does not contain an extension itself, add it
	if filepath.Ext(filename) == "" {
		filename += filepath.Ext(file.Filename)
	}

	filename = strings.ReplaceAll(filename, "/", "-")

	// Create if it doesn't exist or if we are allowed to overwrite it
	_, err := os.Stat(filename)
	if os.IsNotExist(err) || overwrite {
		if err := ioutil.WriteFile(filename, file.Content, 0644); err != nil {
			return errors.Wrapf(err, "failed writing %q file", filename)
		}
		return nil
	}

	return errors.Errorf("%q already exists, use -o to overwrite already existing files\n", filename)
}

// createFiles takes care of recreating folders and files as they were stored in the database.
func createFiles(file *pb.File, path string) error {
	parts := strings.Split(file.Name, "/")

	for i, p := range parts {
		// If it's the last element, create the file
		if i == len(parts)-1 {
			filename := p + filepath.Ext(file.Filename)

			if err := ioutil.WriteFile(filename, file.Content, 0644); err != nil {
				return errors.Wrap(err, "failed writing file")
			}
			// Go back to the top folder
			os.Chdir(path)
			break
		}

		// If it's not the last element, create a folder
		if err := os.MkdirAll(p, os.ModeDir); err != nil {
			return errors.Wrap(err, "failed making directory")
		}
		os.Chdir(p)
	}

	return nil
}
