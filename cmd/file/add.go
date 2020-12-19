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
	ignore     bool
	semaphore  uint32
	bufferSize uint64
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
		Use:   "add <name>",
		Short: "Add files to the database",
		Long: `Add files to the database.

Path to a file must include its extension (in case it has).

The user can specify a path to a folder as well, on this occasion, Kure will iterate over all the files in the folder and potential subfolders (if the -i flag is false) and store them into the database with the name "name/subfolders/filename".

Default behavior in case the buffer flag is not used:
   • file <= 200MB: read the entire file directly to memory.
   • file > 200MB: use a 64MB buffer.`,
		Aliases: []string{"a"},
		Example: addExample,
		PreRunE: cmdutil.RequirePassword(db),
		RunE:    runAdd(db),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset flags (session)
			bufferSize = 0
			ignore = false
			path = ""
			semaphore = 1
		},
	}

	f := cmd.Flags()
	f.Uint64VarP(&bufferSize, "buffer", "b", 0, "buffer size when reading files")
	f.BoolVarP(&ignore, "ignore", "i", false, "ignore subfolders")
	f.StringVarP(&path, "path", "p", "", "path to the file/folder")
	f.Uint32VarP(&semaphore, "semaphore", "s", 1, "maximum number of goroutines running concurrently")

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
			fInfo, err := f.Stat()
			if err != nil {
				return errors.Wrap(err, "failed reading file stats")
			}

			if err := storeFile(db, path, name, fInfo.Size()); err != nil {
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

// storeFile reads, compress and saves a file into the database.
//
// In case the user specified a buffer size, use os.Open instead of ioutil.ReadFile.
func storeFile(db *bolt.DB, path, name string, size int64) error {
	var content []byte

	// If the file size is larger than 200MB and the user didn't specified a buffer, set it to 64MB
	if size > (MB*200) && bufferSize == 0 {
		bufferSize = MB * 64
	}

	fmt.Printf("\nAdd: %s", filepath.Base(path))

	switch bufferSize {
	case 0:
		f, err := ioutil.ReadFile(path)
		if err != nil {
			return errors.Wrap(err, "failed reading file")
		}

		content = f

	default:
		f, err := os.Open(path)
		if err != nil {
			return errors.Wrap(err, "failed opening file")
		}
		defer f.Close()

		buf := make([]byte, bufferSize)

		for {
			n, err := f.Read(buf)
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
	content = bytes.TrimSpace(content)
	filename := filepath.Base(path)
	name = strings.ToLower(name)
	if filepath.Ext(name) == "" {
		name += filepath.Ext(filename)
	}

	f := &pb.File{
		Name:      name,
		Filename:  filename,
		Content:   content,
		CreatedAt: time.Now().Unix(),
	}

	if err := file.Create(db, f); err != nil {
		return err
	}

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
	// Join the name provided by the user (in this case a folder) with the file name,
	// prefer filepath.Join only when working with paths
	name = fmt.Sprintf("%s/%s", name, file.Name())

	if !file.IsDir() {
		if err := storeFile(db, path, name, file.Size()); err != nil {
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
