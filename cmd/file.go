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

	"github.com/GGP1/kure/db"
	"github.com/GGP1/kure/model/file"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	many      []string
	overwrite bool
)
var fileCmd = &cobra.Command{
	Use:   "file",
	Short: "File operations",
	Long: `Save the encrypted bytes of the file in the database, create, delete or list wherever you want. The user must specify the complete path, including the file type, for example: path/to/file/sample.png.

In case you don't specify where to list the file, it will be created in your current directory, for listing all the files kure will create a folder with the name "kure_files" in your directory and save all the files there.`}

var addFile = &cobra.Command{
	Use:   "add <name> [-p path]",
	Short: "Add files to the database",
	Long:  "The user can specify either a path to a file or to a folder, in case it points to a folder, Kure will iterate over all the files in the folder (ignoring sub folders) and store them into the database with the name: '<name>-<file number>'",
	Run: func(cmd *cobra.Command, args []string) {
		var wg sync.WaitGroup
		name := strings.Join(args, " ")
		if name == "" {
			fatal(errInvalidName)
		}
		name = strings.TrimSpace(strings.ToLower(name))

		if path == "" || path == "/" {
			fatalf("invalid path")
		}

		path = filepath.Clean(path)

		file, err := os.Open(path)
		if err != nil {
			fatalf("failed reading file on %s: %v", path, err)
		}
		defer file.Close()

		dir, _ := file.Readdir(0)

		// len(dir) == 0 means it's a file
		if len(dir) == 0 {
			storeFile(path, name)
			return
		}

		wg.Add(len(dir))
		var num int
		for _, f := range dir {
			if f.IsDir() {
				wg.Done()
				continue
			}

			num++
			n := fmt.Sprintf("%s-%d", name, num)
			p := fmt.Sprintf("%s/%s", path, f.Name())
			go func(p, n string) {
				defer wg.Done()
				storeFile(p, n)
			}(p, n)
		}
		wg.Wait()
	},
}

var creatFile = &cobra.Command{
	Use:   "create <name> [-m many] [-p path] [-o overwrite]",
	Short: "Create one, many or all the files from the database. In case a path is passed, kure will create any missing folders for you",
	Run: func(cmd *cobra.Command, args []string) {
		var wg sync.WaitGroup
		name := strings.Join(args, " ")

		if path != "" {
			if path == "/" {
				path = ""
			}

			if err := os.MkdirAll(path, os.ModeDir); err != nil {
				fatalf("failed creating directory: %v", err)
			}

			if err := os.Chdir(path); err != nil {
				fatalf("failed changing directory: %v", err)
			}
		}

		if name != "" {
			file, err := db.GetFile(name)
			if err != nil {
				fatal(err)
			}

			if err := createFile(file, overwrite); err != nil {
				fatal(err)
			}
			return
		}

		if len(many) > 0 {
			wg.Add(len(many))
			for _, m := range many {
				go func() {
					defer wg.Done()

					file, err := db.GetFile(strings.ToLower(m))
					if err != nil {
						fmt.Println(err) // Do not exit if a file doesn't exists, just notify
					}

					if err := createFile(file, overwrite); err != nil {
						fatal(err)
					}
				}()
				wg.Wait()
			}
			return
		}

		files, err := db.ListFiles()
		if err != nil {
			fatal(err)
		}

		wg.Add(len(files))
		for _, f := range files {
			go func(f *file.File) {
				defer wg.Done()
				if err := createFile(f, overwrite); err != nil {
					fatal(err)
				}
			}(f)
		}
		wg.Wait()
	},
}

var deleteFile = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a file from the database",
	Run: func(cmd *cobra.Command, args []string) {
		name := strings.Join(args, " ")
		if name == "" {
			fatal(errInvalidName)
		}

		_, err := db.GetFile(name)
		if err != nil {
			fatalf("%s file does not exist", name)
		}

		scanner := bufio.NewScanner(os.Stdin)
		fmt.Print("Are you sure you want to proceed? [y/n]: ")

		scanner.Scan()
		text := scanner.Text()
		input := strings.ToLower(text)

		if strings.Contains(input, "y") {
			if err := db.DeleteFile(name); err != nil {
				fatal(err)
			}

			fmt.Printf("\nSuccessfully deleted %s file.", name)
		}
	},
}

var listFile = &cobra.Command{
	Use:   "list <name>",
	Short: "Display information about files",
	Run: func(cmd *cobra.Command, args []string) {
		name := strings.Join(args, " ")

		if name != "" {
			file, err := db.GetFile(name)
			if err != nil {
				fatal(err)
			}

			t := time.Unix(file.CreatedAt, 0)

			fmt.Printf("\nName: %s\nType: %s\nCreated at: %v\n", file.Name, file.Filename, t)
			return
		}

		files, err := db.ListFiles()
		if err != nil {
			fatal(err)
		}

		for _, file := range files {
			t := time.Unix(file.CreatedAt, 0)
			bytes := len(file.Content)
			size := fmt.Sprintf("%d bytes", bytes)

			if bytes >= 1024 && bytes < 1048576 {
				size = fmt.Sprintf("%d KB", bytes/1024)
			} else if bytes >= 1048576 && bytes < 1073741824 {
				size = fmt.Sprintf("%d MB", bytes/1048576)
			} else if bytes >= 1073741824 {
				size = fmt.Sprintf("%d GB", bytes/1073741824)
			}

			fmt.Printf("\nName: %s\nFilename: %s\nSize: %s\nCreated at: %v\n", file.Name, file.Filename, size, t)
		}
	},
}

func init() {
	rootCmd.AddCommand(fileCmd)

	fileCmd.AddCommand(addFile)
	fileCmd.AddCommand(creatFile)
	fileCmd.AddCommand(deleteFile)
	fileCmd.AddCommand(listFile)

	addFile.Flags().StringVarP(&path, "path", "p", "", "file path")
	creatFile.Flags().StringSliceVarP(&many, "many", "m", nil, "a list of the files to create")
	creatFile.Flags().StringVarP(&path, "path", "p", "", "destination path (where the file will be created)")
	creatFile.Flags().BoolVarP(&overwrite, "overwrite", "o", false, "set to overwrite files if they already exist in the path provided")
}

// storeFile saves the file to the database
func storeFile(path, name string) {
	f, err := ioutil.ReadFile(path)
	if err != nil {
		fatalf("failed reading file: %v", err)
	}

	filename := filepath.Base(path)
	content := bytes.TrimSpace(f)
	createdAt := time.Now().Unix()

	file := file.New(name, filename, content, createdAt)

	if err := db.CreateFile(file); err != nil {
		fatal(err)
	}

	fmt.Printf("Created %s\n", filename)
}

// createFile creates a file with the filename and content provided.
func createFile(file *file.File, overwrite bool) error {
	filename := file.Name + filepath.Ext(file.Filename)

	_, err := os.Stat(filename)
	if os.IsNotExist(err) || overwrite {
		if err := ioutil.WriteFile(filename, file.Content, 0644); err != nil {
			return errors.Wrapf(err, "failed writing %s file", filename)
		}
		fmt.Printf("Created %s\n", filename)
		return nil
	}

	fmt.Printf("%s already exists, skipping", filename)
	return nil
}
