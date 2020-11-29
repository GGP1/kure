package file

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	cmdutil "github.com/GGP1/kure/cmd"
	"github.com/GGP1/kure/db/file"
	"github.com/GGP1/kure/pb"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var (
	ignore    bool
	semaphore uint32
	buffer    uint64
)

var addExample = `
* Add a new file
kure file add -p path/to/file

* Add a folder and all its subfolders, limiting goroutine number to 40
kure file add -p path/to/folder -s 40

* Add files from a folder, ignoring subfolders and using a 4096 bytes buffer
kure file add -p path/to/folder -i -b 4096`

func addSubCmd(db *bolt.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <name> [-b buffer] [-i ignore] [-p path] [-s semaphore]",
		Short: "Add files to the database",
		Long: `Add files to the database.
	
Path to a file must include its extension (in case it has).
	
The user can specify a path to a folder also, on this occasion, Kure will iterate over all the files in the folder and potential subfolders and store them into the database with the name "name/subfolders/filename".
Use the -i flag to ignore subfolders and focus only on the folder's files.`,
		Aliases: []string{"a"},
		Example: addExample,
		RunE:    runAdd(db),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset flags defaults (session)
			buffer = 0
			ignore = false
			path = ""
			semaphore = 1
		},
	}

	f := cmd.Flags()
	f.Uint64VarP(&buffer, "buffer", "b", 0, "buffer size when reading files, by default it reads the entire file directly to memory")
	f.BoolVarP(&ignore, "ignore", "i", false, "ignore subfolders")
	f.StringVarP(&path, "path", "p", "", "file/folder path")
	f.Uint32VarP(&semaphore, "semaphore", "s", 1, "maximum number of goroutines running concurrently")

	return cmd
}

func runAdd(db *bolt.DB) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")
		if name == "" {
			return errInvalidName
		}
		if !strings.Contains(name, "/") && len(name) > 57 {
			return errors.New("file name must contain 57 letters or less")
		}

		if path == "" {
			return errInvalidPath
		}

		if semaphore == 0 {
			return errors.New("error: invalid semaphore quantity")
		}

		absolute, err := filepath.Abs(path)
		if err != nil {
			return errInvalidPath
		}
		path = absolute

		if err := cmdutil.RequirePassword(db); err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return errors.Errorf("failed reading %s: %v", path, err)
		}
		defer file.Close()

		dir, _ := file.Readdir(0)
		// len(dir) = 0 means it's a file
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

// storeFile saves the file in the database.
func storeFile(db *bolt.DB, path, name string) error {
	var content []byte

	// Check if the last element of the name is not too long
	fullname := strings.Split(name, "\\")
	basename := fullname[len(fullname)-1]
	if len(basename) > 57 {
		return errors.Errorf("%q is longer than 57 characters", basename)
	}

	switch buffer {
	case 0:
		file, err := ioutil.ReadFile(path)
		if err != nil {
			return errors.Wrap(err, "failed reading file")
		}
		content = file

	default:
		file, err := os.Open(path)
		if err != nil {
			return errors.Wrap(err, "failed opening file")
		}
		defer file.Close()

		buf := make([]byte, buffer)

		for {
			n, err := file.Read(buf)
			if err == io.EOF {
				break
			}
			if err != nil {
				return errors.Wrap(err, "failed reading file")
			}

			content = append(content, buf[:n]...)
		}
	}

	// Format fields
	name = strings.ToLower(strings.ReplaceAll(name, "\\", "/"))
	if filepath.Ext(name) == "" {
		name += filepath.Ext(path)
	}

	filename := filepath.Base(path)
	content = bytes.TrimSpace(content)

	f := &pb.File{
		Name:      name,
		Filename:  filename,
		Content:   content,
		CreatedAt: time.Now().Unix(),
	}

	if err := file.Create(db, f); err != nil {
		return err
	}

	fmt.Printf("Added %q\n", filename)
	return nil
}

// walkDir iterates over the content of a folder and calls checkFile.
//
// Receiving wg and sem as parameters prevents us from initializating them in this function
// and getting an error when called multiple times.
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
	sem <- struct{}{}
	defer func() {
		wg.Done()
		<-sem
	}()

	path = filepath.Join(path, file.Name())
	name = filepath.Join(name, file.Name())

	if !file.IsDir() {
		if err := storeFile(db, path, name); err != nil {
			fmt.Println("error:", err)
		}
		return
	}

	// Ignore subfolders
	if ignore {
		return
	}

	subdir, err := ioutil.ReadDir(path)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}

	if len(subdir) != 0 {
		go walkDir(db, subdir, path, name, wg, sem)
	}
}
