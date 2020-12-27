package file

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	cmdutil "github.com/GGP1/kure/cmd"
	"github.com/GGP1/kure/db/file"
	"github.com/GGP1/kure/pb"

	"github.com/awnumar/memguard"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var (
	ignore    bool
	semaphore uint32
)

var addExample = `
* Add a new file
kure file add -p path/to/file

* Add a folder and all its subfolders, limiting goroutine number to 40
kure file add -p path/to/folder -s 40

* Add files from a folder, ignoring subfolders
kure file add -p path/to/folder -i`

func addSubCmd(db *bolt.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Add files to the database",
		Long: `Add files to the database. As they are stored in a database, the whole file is read into memory, please have this into account when adding new ones.

Path to a file must include its extension (in case it has).

The user can specify a path to a folder as well, on this occasion, Kure will iterate over all the files in the folder and potential subfolders (if the -i flag is false) and store them into the database with the name "name/subfolders/filename".`,
		Aliases: []string{"a"},
		Example: addExample,
		PreRunE: cmdutil.RequirePassword(db),
		RunE:    runAdd(db),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset flags (session)
			ignore = false
			path = ""
			semaphore = 20
		},
	}

	f := cmd.Flags()
	f.BoolVarP(&ignore, "ignore", "i", false, "ignore subfolders")
	f.StringVarP(&path, "path", "p", "", "path to the file/folder")
	f.Uint32VarP(&semaphore, "semaphore", "s", 20, "maximum number of goroutines running concurrently")

	return cmd
}

func runAdd(db *bolt.DB) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")
		if name == "" {
			return errInvalidName
		}

		if path == "" {
			return errInvalidPath
		}

		if semaphore == 0 {
			return errors.New("invalid semaphore quantity")
		}

		if err := cmdutil.Exists(db, name, "file"); err != nil {
			return err
		}

		absolute, err := filepath.Abs(path)
		if err != nil {
			return errInvalidPath
		}
		path = absolute

		f, err := os.Open(path)
		if err != nil {
			return errors.Wrap(err, "failed reading file")
		}
		defer f.Close()

		dir, _ := f.Readdir(0)
		// 0 means it's a file
		if len(dir) == 0 {
			if err := storeFile(db, path, name); err != nil {
				return err
			}
			return nil
		}

		var wg sync.WaitGroup
		sem := make(chan struct{}, semaphore)

		walkDir(db, dir, path, name, &wg, sem)

		return nil
	}
}

// walkDir iterates over the items of a folder and calls checkFile.
func walkDir(db *bolt.DB, dir []os.FileInfo, path, name string, wg *sync.WaitGroup, sem chan struct{}) {
	wg.Add(len(dir))
	for _, f := range dir {
		go checkFile(db, f, path, name, wg, sem)
	}
	wg.Wait()
}

// checkFile checks if the item is a file or a folder.
//
// If the item is a file, it stores it.
//
// If it's a folder it repeats the process until there are no left files to store.
//
// Errors are not returned but logged.
func checkFile(db *bolt.DB, file os.FileInfo, path, name string, wg *sync.WaitGroup, sem chan struct{}) {
	defer func() {
		wg.Done()
		<-sem
	}()

	sem <- struct{}{}

	path = filepath.Join(path, file.Name())
	// Join the name provided by the user (in this case a folder) with the file name,
	// prefer filepath.Join only when working with paths
	name = fmt.Sprintf("%s/%s", name, file.Name())

	if !file.IsDir() {
		if err := storeFile(db, path, name); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
		}
		return
	}

	// Ignore subfolders
	if ignore {
		return
	}

	subdir, err := ioutil.ReadDir(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed reading directory: %v\n", err)
		return
	}

	if len(subdir) != 0 {
		go walkDir(db, subdir, path, name, wg, sem)
	}
}

// storeFile reads, compress and saves a file into the database.
func storeFile(db *bolt.DB, path, name string) error {
	// TODO: find a better way to store large files without exhausting memory
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return errors.Wrap(err, "failed reading file")
	}

	// Format fields
	filename := filepath.Base(path)
	name = strings.ToLower(name)
	if filepath.Ext(name) == "" {
		name += filepath.Ext(filename)
	}

	f := &pb.File{
		Name:      name,
		Content:   bytes.TrimSpace(content),
		Size:      int64(len(content)),
		CreatedAt: time.Now().Unix(),
	}

	if err := file.Create(db, f); err != nil {
		return err
	}
	memguard.WipeBytes(content)

	fmt.Printf("Add: %s\n", filepath.Base(path))
	return nil
}
