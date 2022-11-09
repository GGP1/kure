package add

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/GGP1/kure/auth"
	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/file"
	"github.com/GGP1/kure/pb"
	"github.com/GGP1/kure/terminal"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

const example = `
* Add a new file
kure file add Sample -p path/to/file

* Add a note
kure file add Sample -n

* Add a folder and all its subfolders, limiting goroutine number to 40
kure file add Sample -p path/to/folder -s 40

* Add files from a folder, ignoring subfolders
kure file add Sample -p path/to/folder -i`

type addOptions struct {
	path      string
	note      bool
	ignore    bool
	semaphore uint32
}

// NewCmd returns a new command.
func NewCmd(db *bolt.DB, r io.Reader) *cobra.Command {
	opts := addOptions{}

	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Add files to the database",
		Long: `Add files to the database. As they are stored in a database, the whole file is read into memory, please have this into account when adding new ones.

Path to a file must include its extension (in case it has one).

The user can specify a path to a folder as well, on this occasion, Kure will iterate over all the files in the folder and potential subfolders (if the -i flag is false) and store them into the database with the name "name/subfolders/filename". Empty folders will be skipped.`,
		Aliases: []string{"new"},
		Example: example,
		Args:    cmdutil.MustNotExist(db, cmdutil.File),
		PreRunE: auth.Login(db),
		RunE:    runAdd(db, r, &opts),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset variables (session)
			opts = addOptions{
				semaphore: 50,
			}
		},
	}

	f := cmd.Flags()
	f.BoolVarP(&opts.ignore, "ignore", "i", false, "ignore subfolders")
	f.StringVarP(&opts.path, "path", "p", "", "path to the file/folder")
	f.BoolVarP(&opts.note, "note", "n", false, "add a note")
	f.Uint32VarP(&opts.semaphore, "semaphore", "s", 50, "maximum number of goroutines running concurrently")

	return cmd
}

func runAdd(db *bolt.DB, r io.Reader, opts *addOptions) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")
		name = cmdutil.NormalizeName(name)

		if opts.note {
			return addNote(db, r, name)
		}

		if opts.semaphore < 1 {
			return errors.New("invalid semaphore quantity")
		}

		// Add the extension to differentiate between files and folders
		if filepath.Ext(name) == "" && opts.path != "" {
			name += filepath.Ext(opts.path)
		}

		dir, err := os.ReadDir(opts.path)
		if err != nil {
			// If it's not a directory, attempt storing a file
			return storeFile(db, opts.path, name)
		}

		if len(dir) == 0 {
			return errors.Errorf("%q directory is empty", filepath.Base(opts.path))
		}

		// With filepath.Walk we can't associate each file
		// with its folder, also, this may perform poorly on large directories.
		// Allocations are reduced and performance improved.
		var wg sync.WaitGroup
		sem := make(chan struct{}, opts.semaphore)
		wg.Add(len(dir))
		walkDir(db, dir, opts.path, name, opts.ignore, &wg, sem)
		wg.Wait()
		return nil
	}
}

// walkDir iterates over the items of a folder and calls checkFile.
func walkDir(db *bolt.DB, dir []os.DirEntry, path, name string, ignore bool, wg *sync.WaitGroup, sem chan struct{}) {
	for _, f := range dir {
		// If it's not a directory or a regular file, skip
		if !f.IsDir() && !f.Type().IsRegular() {
			wg.Done()
			continue
		}

		go checkFile(db, f, path, name, ignore, wg, sem)
	}
}

// checkFile checks if the item is a file or a folder.
//
// If the item is a file, it stores it.
//
// If it's a folder it repeats the process until there are no left files to store.
//
// Errors are not returned but logged.
func checkFile(db *bolt.DB, file os.DirEntry, path, name string, ignore bool, wg *sync.WaitGroup, sem chan struct{}) {
	defer func() {
		wg.Done()
		<-sem
	}()

	sem <- struct{}{}

	// Join name (which is the folder name) with the file name (that could be a folder as well)
	name = fmt.Sprintf("%s/%s", name, file.Name())
	// Prefer filepath.Join when working with paths only
	path = filepath.Join(path, file.Name())

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

	subdir, err := os.ReadDir(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: reading directory: %v\n", err)
		return
	}

	if len(subdir) != 0 {
		wg.Add(len(subdir))
		go walkDir(db, subdir, path, name, ignore, wg, sem)
	}
}

// storeFile reads and saves a file into the database.
func storeFile(db *bolt.DB, path, filename string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return errors.Wrap(err, "reading file")
	}

	f := &pb.File{
		Name:      strings.ToLower(filename),
		Content:   content,
		Size:      int64(len(content)),
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Time{}.Unix(),
	}

	// There is no better way to report as Batch combines
	// all the transactions into a single one
	abs, _ := filepath.Abs(path)

	fmt.Println("Add:", abs)
	return file.Create(db, f)
}

// addNote takes input from the user and creates a file inside the "notes" folder
// and with the .txt extension.
func addNote(db *bolt.DB, r io.Reader, name string) error {
	name = "notes/" + name
	if filepath.Ext(name) == "" {
		name += ".txt"
	}

	if err := cmdutil.Exists(db, name, cmdutil.File); err != nil {
		return err
	}

	text := terminal.Scanlns(bufio.NewReader(r), "Text")

	f := &pb.File{
		Name:      name,
		Content:   []byte(text),
		Size:      int64(len(text)),
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Time{}.Unix(),
	}

	fmt.Println("Add:", name)
	return file.Create(db, f)
}
