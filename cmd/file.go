package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/GGP1/kure/db"
	"github.com/GGP1/kure/model/file"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var fileCmd = &cobra.Command{
	Use:   "file",
	Short: "File operations",
	Long: `Save the encrypted bytes of the file in the database, delete or list them creating a file wherever you want. The user must specify the complete path, including the file type, for example: path/to/file/sample.png

In case you don't specify where to list the file, it will be created in your current directory, for listing all the files kure will create a folder with the name "kure_files" in your directory and save all the files there.`}

var addFile = &cobra.Command{
	Use:   "add <name> [-p path]",
	Short: "Add a file to the database",
	Run: func(cmd *cobra.Command, args []string) {
		name := strings.Join(args, " ")
		if name == "" {
			fatal(errInvalidName)
		}

		if path == "" {
			fatal(errors.New("please specify which file to save"))
		}

		content, err := ioutil.ReadFile(path)
		if err != nil {
			fatalf("failed reading file: %v", err)
		}

		split := strings.Split(path, ".")
		fileType := split[len(split)-1]

		name = strings.TrimSpace(strings.ToLower(name))
		createdAt := time.Now().Unix()

		file := file.New(name, fileType, content, createdAt)

		if err := db.CreateFile(path, file); err != nil {
			fatal(err)
		}
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

var infoFile = &cobra.Command{
	Use:   "info <name>",
	Short: "Display information about each file in the bucket",
	Run: func(cmd *cobra.Command, args []string) {
		name := strings.Join(args, " ")

		if name != "" {
			file, err := db.GetFile(name)
			if err != nil {
				fatal(err)
			}

			t := time.Unix(file.CreatedAt, 0)

			fmt.Printf("\nName: %s\nType: %s\nCreated at: %v\n", file.Name, file.Type, t)
			return
		}

		files, err := db.ListFiles()
		if err != nil {
			fatal(err)
		}

		for i, file := range files {
			t := time.Unix(file.CreatedAt, 0)

			fmt.Printf(`
%d:
   Name: %s
   Type: %s
   Created at: %v
		`, i, file.Name, file.Type, t)
		}
	},
}

var listFile = &cobra.Command{
	Use:   "list <name> [-p path]",
	Short: "List a files or all the files from the database",
	Run: func(cmd *cobra.Command, args []string) {
		var wg sync.WaitGroup
		name := strings.Join(args, " ")

		if name != "" {
			file, err := db.GetFile(name)
			if err != nil {
				fatal(err)
			}

			if err := createFileFile(path, file); err != nil {
				fatal(err)
			}
			return
		}

		if err := os.Mkdir("kure_files", os.ModeDir); err != nil {
			fatalf("failed creating directory: %v", err)
		}

		if err := os.Chdir("kure_files"); err != nil {
			fatalf("failed changing directory: %v", err)
		}

		files, err := db.ListFiles()
		if err != nil {
			fatal(err)
		}

		wg.Add(len(files))
		for _, f := range files {
			go func(path string, f *file.File) {
				defer wg.Done()
				if err := createFileFile(path, f); err != nil {
					fatal(err)
				}
				fmt.Printf("Created %s.%s\n", f.Name, f.Type)
			}(path, f)
		}
		wg.Wait()
	},
}

func init() {
	rootCmd.AddCommand(fileCmd)

	fileCmd.AddCommand(addFile)
	fileCmd.AddCommand(deleteFile)
	fileCmd.AddCommand(infoFile)
	fileCmd.AddCommand(listFile)

	addFile.Flags().StringVarP(&path, "path", "p", "", "file path")
	listFile.Flags().StringVarP(&path, "path", "p", "", "file path")
}

// createFileFile checks the file type and creates the file.
func createFileFile(path string, file *file.File) error {
	buf := bytes.NewReader(file.Content)

	if path != "" {
		path = fmt.Sprintf("%s/%s.%s", path, file.Name, file.Type)
	} else {
		path = fmt.Sprintf("%s.%s", file.Name, file.Type)
	}

	f, err := os.Create(path)
	if err != nil {
		return errors.Wrapf(err, "failed creating %s.%s file", file.Name, file.Type)
	}

	_, err = buf.WriteTo(f)
	if err != nil {
		return errors.Wrapf(err, "failed writing %s.%s file", file.Name, file.Type)
	}

	return nil
}
