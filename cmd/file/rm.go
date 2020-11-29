package file

import (
	"fmt"
	"io"
	"strings"
	"sync"

	cmdutil "github.com/GGP1/kure/cmd"
	"github.com/GGP1/kure/db/file"

	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var directory bool

var rmExample = `
* Remove a file
kure file rm fileName

* Remove a directory using a maximum of 25 goroutines
kure file rm directoryName -d -s 25`

func rmSubCmd(db *bolt.DB, r io.Reader) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rm <name> [-d directory] [-s semaphore]",
		Short:   "Remove files from the database",
		Example: rmExample,
		RunE:    runRm(db, r),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset flags defaults (session)
			directory = false
			semaphore = 1
		},
	}

	f := cmd.Flags()
	f.BoolVarP(&directory, "dir", "d", false, "remove a directory and all the files stored into it")
	f.Uint32VarP(&semaphore, "semaphore", "s", 1, "maximum number of goroutines running concurrently")

	return cmd
}

func runRm(db *bolt.DB, r io.Reader) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")
		if name == "" {
			return errInvalidName
		}

		if !cmdutil.Proceed(r) {
			return nil
		}

		if !directory {
			if err := file.Remove(db, name); err != nil {
				return err
			}

			fmt.Printf("\nSuccessfully removed %q file.\n", name)
			return nil
		}

		// If the last rune of name is not a slash, add it
		if name[len(name)-1] != '/' {
			name += "/"
		}

		files, err := file.ListByName(db, name)
		if err != nil {
			return err
		}

		var wg sync.WaitGroup
		sem := make(chan struct{}, semaphore)

		fmt.Printf("\nRemoving %q directory...\n", name)

		wg.Add(len(files))
		for _, f := range files {
			go removeFile(db, f.Name, &wg, sem)
		}
		wg.Wait()

		fmt.Printf("\nSuccessfully removed %q directory and all its files.\n", name)

		return nil
	}
}

func removeFile(db *bolt.DB, name string, wg *sync.WaitGroup, sem chan struct{}) {
	sem <- struct{}{}

	if err := file.Remove(db, name); err != nil {
		fmt.Println("error:", err)
	}

	wg.Done()
	<-sem
}
