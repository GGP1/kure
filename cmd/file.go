package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/GGP1/kure/crypt"
	"github.com/GGP1/kure/db"
	"github.com/GGP1/kure/pb"
	"github.com/GGP1/kure/tree"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

const (
	_ = 1 << (10 * iota)
	// KB - 1024 bytes
	KB
	// MB - 1048576 bytes
	MB
	// GB - 1073741824 bytes
	GB
)

var fileCmd = &cobra.Command{
	Use:   "file",
	Short: "File operations",
	Long: `Save the encrypted bytes of the file in the database, create, list or remove wherever you want. The user must specify the complete path, including the file type, for example: path/to/file/sample.png.

In case you don't specify where to list the file, it will be created in your current directory, for listing all the files kure will create a folder with the name "kure_files" in your directory and save all the files there.`}

var addFile = &cobra.Command{
	Use:   "add <name> [-i ignore] [-p path] [-s semaphore]",
	Short: "Add files to the database",
	Long: `Add files to the database.

Path to a file must include its extension (in case it has).

The user can specify a path to a folder also, on this occasion, Kure will iterate over all the files in the folder and potential subfolders and store them into the database with the name "name/subfolders/filename".
Use the -i flag to ignore subfolders and focus only on the folder's files.`,
	Run: func(cmd *cobra.Command, args []string) {
		name := strings.Join(args, " ")
		if name == "" {
			fatal(errInvalidName)
		}

		if path == "" {
			fatal(errInvalidPath)
		}

		name = strings.TrimSpace(strings.ToLower(name))
		if !strings.Contains(name, "/") && len(name) > 43 {
			fatalf("file name must contain 43 letters or less")
		}

		absolute, err := filepath.Abs(path)
		if err != nil {
			fatal(errInvalidPath)
		}
		path = absolute

		password, err := crypt.GetMasterPassword()
		if err != nil {
			fatal(err)
		}

		file, err := os.Open(path)
		if err != nil {
			fatalf("failed reading %s: %v", path, err)
		}
		defer file.Close()

		dir, _ := file.Readdir(0)
		// len(dir) = 0 means it's a file
		if len(dir) == 0 {
			storeFile(path, name, password)
			return
		}

		var wg sync.WaitGroup
		if semaphore == 0 {
			semaphore = 25
		}
		sem := make(chan struct{}, semaphore)
		walkDir(dir, path, name, password, &wg, sem)
	},
}

var creatFile = &cobra.Command{
	Use:   "create <name> [-m many] [-p path] [-o overwrite]",
	Short: "Create stored files",
	Long:  "Create one, many or all the files from the database. In case a path is passed, kure will create any missing folders for you",
	Run: func(cmd *cobra.Command, args []string) {
		var wg sync.WaitGroup
		name := strings.Join(args, " ")

		absolute, err := filepath.Abs(path)
		if err != nil {
			fatal(errInvalidPath)
		}
		path = absolute

		if err := os.MkdirAll(path, os.ModeDir); err != nil {
			fatal(errors.Wrap(err, "failed making directory"))
		}

		if err := os.Chdir(path); err != nil {
			fatal(errors.Wrap(err, "failed changing directory"))
		}

		// Single file
		if name != "" {
			file, err := db.GetFile(name)
			if err != nil {
				fatal(err)
			}

			if err := createFile(file, overwrite); err != nil {
				fatal(err)
			}

			fmt.Printf("Created \"%s\" on path %s\n", file.Name, path)
			return
		}

		// Many files
		if len(many) > 0 {
			fmt.Printf("\nCreating files into %s\n", path)

			wg.Add(len(many))
			for _, m := range many {
				go func() {
					defer wg.Done()

					m = strings.ToLower(m)
					file, err := db.GetFile(m)
					if err != nil {
						// Do not exit, just notify
						fmt.Println("error:", err)
					}

					if err := createFile(file, overwrite); err != nil {
						fmt.Println("error:", err)
					}
				}()
				wg.Wait()
			}
			return
		}

		// All files
		files, err := db.ListFiles()
		if err != nil {
			fatal(err)
		}

		paths := make([]string, len(files))
		contents := make(map[string][]byte, len(files))

		for i, f := range files {
			paths[i] = f.Name + filepath.Ext(f.Filename)
			contents[f.Name] = f.Content
		}

		// Build tree and get its root
		root := tree.Root(paths)

		fmt.Printf("Creating files into %s\n", path)

		if err := createAllFiles(root, contents, path); err != nil {
			fatal(err)
		}
	},
}

var lsFile = &cobra.Command{
	Use:   "ls <name> [-f filter]",
	Short: "List files",
	Run: func(cmd *cobra.Command, args []string) {
		name := strings.Join(args, " ")
		if name == "" {
			files, err := db.ListFiles()
			if err != nil {
				fatal(err)
			}

			paths := make([]string, len(files))

			for i, file := range files {
				paths[i] = file.Name
			}

			tree.Print(paths)
			return
		}

		if filter {
			files, err := db.FilesByName(name)
			if err != nil {
				fatal(err)
			}

			for _, file := range files {
				printFile(file)
			}
			return
		}

		file, err := db.GetFile(name)
		if err != nil {
			fatal(err)
		}

		printFile(file)
	},
}

var renameFile = &cobra.Command{
	Use:   "rename <name>",
	Short: "Rename a file",
	Run: func(cmd *cobra.Command, args []string) {
		var newName string
		oldName := strings.Join(args, " ")

		scanner := bufio.NewScanner(os.Stdin)
		scan(scanner, "New name", &newName)

		if oldName == "" || newName == "" {
			fatal(errInvalidName)
		}

		if err := db.RenameFile(oldName, newName); err != nil {
			fatal(err)
		}

		fmt.Printf("\nSuccessfully renamed \"%s\" as \"%s\".\n", oldName, newName)
	},
}

var rmFile = &cobra.Command{
	Use:   "rm <name> [-d directory]",
	Short: "Remove files from the database",
	Run: func(cmd *cobra.Command, args []string) {
		name := strings.Join(args, " ")
		if name == "" {
			fatal(errInvalidName)
		}

		if directory {
			if string(name[len(name)-1]) != "/" {
				name += "/"
			}

			files, err := db.FilesByName(name)
			if err != nil {
				fatal(err)
			}

			if proceed() {
				var wg sync.WaitGroup
				wg.Add(len(files))
				for _, f := range files {
					go func(fName string) {
						defer wg.Done()
						if err := db.RemoveFile(fName); err != nil {
							fatal(err)
						}
					}(f.Name)
				}
				wg.Wait()
			}

			fmt.Printf("Successfully removed \"%s\" directory and all its files\n", name)
			return
		}

		if proceed() {
			if err := db.RemoveFile(name); err != nil {
				fatal(err)
			}

			fmt.Printf("\nSuccessfully removed \"%s\" file.\n", name)
		}
	},
}

func init() {
	rootCmd.AddCommand(fileCmd)

	fileCmd.AddCommand(addFile)
	fileCmd.AddCommand(creatFile)
	fileCmd.AddCommand(lsFile)
	fileCmd.AddCommand(renameFile)
	fileCmd.AddCommand(rmFile)

	addFile.Flags().StringVarP(&path, "path", "p", "", "file/folder path")
	addFile.Flags().Uint32VarP(&semaphore, "semaphore", "s", 25, "maximum number of goroutines to run when adding files to the database")
	addFile.Flags().BoolVarP(&ignore, "ignore", "i", false, "ignore subfolders")

	creatFile.Flags().StringSliceVarP(&many, "many", "m", nil, "a list of the files to create")
	creatFile.Flags().StringVarP(&path, "path", "p", "", "destination path (where the file will be created)")
	creatFile.Flags().BoolVarP(&overwrite, "overwrite", "o", false, "set to overwrite files if they already exist in the path provided")

	lsFile.Flags().BoolVarP(&filter, "filter", "f", false, "filter files by name")

	rmFile.Flags().BoolVarP(&directory, "dir", "d", false, "remove a directory and all the files stored into it")
}

// createAllFiles uses recursion to navigate over the file tree creating folders and files.
//
// When calling Chdir the path is already checked so no errors can occur.
func createAllFiles(root *tree.Folder, contents map[string][]byte, path string) error {
	for i, r := range root.Children {
		if len(r.Children) == 0 {
			if err := ioutil.WriteFile(r.Name, contents[r.Name], 0644); err != nil {
				return errors.Wrap(err, "failed writing file")
			}
			delete(contents, r.Name)

			// If it is the last element of a branch, switch to the initial path
			if i == len(root.Children)-1 {
				os.Chdir(path)
			}
			continue
		}

		if err := os.Mkdir(r.Name, os.ModeDir); err != nil {
			return errors.Wrap(err, "failed making directory")
		}
		os.Chdir(r.Name)

		createAllFiles(r, contents, path)
	}

	return nil
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
			return errors.Wrapf(err, "failed writing \"%s\" file", filename)
		}
		return nil
	}

	return errors.Errorf("\"%s\" already exists, use -o to overwrite already existing files\n", filename)
}

func printFile(f *pb.File) {
	printObjectName(f.Name)

	t := time.Unix(f.CreatedAt, 0)
	bytes := len(f.Content)
	size := fmt.Sprintf("%d bytes", bytes)

	if bytes >= KB && bytes < MB {
		size = fmt.Sprintf("%d KB", bytes/KB)
	} else if bytes >= MB && bytes < GB {
		size = fmt.Sprintf("%d MB", bytes/MB)
	} else if bytes >= GB {
		size = fmt.Sprintf("%d GB", bytes/GB)
	}

	fmt.Printf(`│ Name       │ %s
│ Filename   │ %s
│ Size       │ %s
│ Created at │ %s
`, f.Name, f.Filename, size, t)

	fmt.Println("+────────────+──────────────────────────────>")
}

// storeFile saves the file in the database.
func storeFile(path, name string, password []byte) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		fatalf("failed reading file: %v", err)
	}

	// Format fields
	name = strings.ToLower(strings.ReplaceAll(name, "\\", "/"))
	filename := filepath.Base(path)
	content = bytes.TrimSpace(content)
	createdAt := time.Now().Unix()

	file := &pb.File{
		Name:      name,
		Filename:  filename,
		Content:   content,
		CreatedAt: createdAt,
	}

	if err := db.CreateFileX(file, password); err != nil {
		fatal(err)
	}

	fmt.Printf("Added \"%s\"\n", filename)
}

// walkDir iterates over the content of a folder and calls checkFile.
//
// Receiving wg and sem as parameters prevents us from initializating them in this function
// and getting an error when called multiple times.
func walkDir(dir []os.FileInfo, path, name string, password []byte, wg *sync.WaitGroup, sem chan struct{}) {
	wg.Add(len(dir))
	for _, f := range dir {
		go checkFile(f, path, name, password, wg, sem)
	}
	wg.Wait()
}

// checkFile checks if the item is a file or a folder.
//
// If the item is a file, it stores it.
//
// If it's a folder it repeats the process until there are no left files to store.
func checkFile(file os.FileInfo, path, name string, password []byte, wg *sync.WaitGroup, sem chan struct{}) {
	sem <- struct{}{}
	defer func() {
		wg.Done()
		<-sem
	}()

	path = filepath.Join(path, file.Name())
	name = filepath.Join(name, file.Name())

	if !file.IsDir() {
		storeFile(path, name, password)
		return
	}

	if ignore {
		return
	}

	subdir, err := ioutil.ReadDir(path)
	if err != nil {
		// Do not exit, just notify
		fmt.Printf("error: failed reading directory: %v\n", err)
	}

	if len(subdir) != 0 {
		go walkDir(subdir, path, name, password, wg, sem)
	}
}
