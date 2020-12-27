package file

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	cmdutil "github.com/GGP1/kure/cmd"
	"github.com/GGP1/kure/db/file"
	"github.com/pkg/errors"

	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var dir bool

var rmExample = `
* Remove a file
kure file rm fileName

* Remove a directory
kure file rm directoryName -d`

func rmSubCmd(db *bolt.DB, r io.Reader) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rm <name>",
		Short:   "Remove files from the database",
		Example: rmExample,
		PreRunE: cmdutil.RequirePassword(db),
		RunE:    runRm(db, r),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset flags (session)
			dir = false
		},
	}

	f := cmd.Flags()
	f.BoolVarP(&dir, "dir", "d", false, "remove a directory and all the files stored into it")

	return cmd
}

func runRm(db *bolt.DB, r io.Reader) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")
		if name == "" {
			return errInvalidName
		}

		name = strings.TrimSpace(strings.ToLower(name))

		if err := cmdutil.Exists(db, name, "file"); err == nil {
			return errors.Errorf("%q does not exist", name)
		}

		if !cmdutil.Proceed(r) {
			return nil
		}

		// Remove single file
		if !dir {
			if err := file.Remove(db, name); err != nil {
				return err
			}

			fmt.Printf("\n%q deleted\n", name)
			return nil
		}

		// Remove directory
		fmt.Printf("Removing %q directory...\n", name)

		files, err := file.ListNames(db)
		if err != nil {
			return err
		}

		// If the last rune of name is not a slash,
		// add it to make sure to delete items under that folder only
		if name[len(name)-1] != '/' {
			name += "/"
		}

		var list []string
		for _, f := range files {
			if strings.HasPrefix(f, name) {
				list = append(list, f)
			}
		}

		var wg sync.WaitGroup
		sem := make(chan struct{}, 25)

		wg.Add(len(list))
		for _, l := range list {
			go removeFile(db, l, &wg, sem)
		}
		wg.Wait()
		return nil
	}
}

// Unlike the others objects' rm command, we use goroutines as a file folder may contain
// subfolders and a lot of files in it, limit is set to 40 to not wake up too many goroutines.
func removeFile(db *bolt.DB, name string, wg *sync.WaitGroup, sem chan struct{}) {
	sem <- struct{}{}

	if err := file.Remove(db, name); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
	}

	fmt.Println("Remove:", name)
	wg.Done()
	<-sem
}
