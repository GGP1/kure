package file

import (
	"bytes"
	"os"
	"testing"

	cmdutil "github.com/GGP1/kure/cmd"
	"github.com/atotto/clipboard"

	bolt "go.etcd.io/bbolt"
)

func TestFile(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	defer db.Close()

	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	NewCmd(db, nil, nil)
	t.Run("Add", add(db))
	t.Run("Cat", cat(db))
	t.Run("Touch", touch(db, currentDir))
	t.Run("Ls", ls(db))
	t.Run("Rename", rename(db))
	t.Run("Rm", rm(db))
}

func add(db *bolt.DB) func(*testing.T) {
	return func(t *testing.T) {
		cases := []struct {
			desc      string
			name      string
			ignore    string
			path      string
			semaphore string
			pass      bool
		}{
			{
				desc:      "Add a file",
				name:      "test",
				path:      "testdata/test_file.txt",
				semaphore: "1",
				pass:      true,
			},
			{
				desc:      "Add all files synchronously",
				name:      "testAll",
				path:      "testdata/",
				semaphore: "1",
				pass:      true,
			},
			{
				desc:      "Add all files using goroutines",
				name:      "testAll2",
				path:      "testdata/",
				semaphore: "8",
				pass:      true,
			},
			{
				desc:      "Add all files ignoring subfolders",
				name:      "testAll3",
				ignore:    "true",
				path:      "testdata/",
				semaphore: "3",
				pass:      true,
			},
			{
				desc: "Invalid name",
				name: "",
				pass: false,
			},
			{
				desc: "Invalid path",
				name: "test",
				path: "",
				pass: false,
			},
			{
				desc:      "Invalid semaphore",
				name:      "test",
				path:      "testdata/test_file.txt",
				semaphore: "0",
				pass:      false,
			},
			{
				desc:      "Non-existent",
				name:      "non-existent",
				semaphore: "1",
				path:      "testdata/non-existent.txt",
				pass:      false,
			},
		}

		cmd := addSubCmd(db)

		for _, tc := range cases {
			t.Run(tc.desc, func(t *testing.T) {
				args := []string{tc.name}

				f := cmd.Flags()
				f.Set("ignore", tc.ignore)
				f.Set("path", tc.path)
				f.Set("semaphore", tc.semaphore)

				err := cmd.RunE(cmd, args)
				if err != nil && tc.pass {
					t.Errorf("Failed adding file: %v", err)
				}
				if err == nil && !tc.pass {
					t.Error("Expected an error and got nil")
				}
			})
		}
	}
}

func cat(db *bolt.DB) func(*testing.T) {
	return func(t *testing.T) {
		file1 := "test.txt"
		file2 := "testall2/subfolder/file.txt"

		cases := []struct {
			desc     string
			args     []string
			expected string // Hardcoded to avoid listing files before being created
			copy     string
		}{
			{
				desc:     "Read files",
				args:     []string{file1, file2},
				expected: "testing filetest",
				copy:     "false",
			},
			{
				desc:     "Read and copy to clipboard",
				args:     []string{file1},
				expected: "testing file",
				copy:     "true",
			},
		}

		for _, tc := range cases {
			t.Run(tc.desc, func(t *testing.T) {
				var buf bytes.Buffer
				cmd := catSubCmd(db, &buf)
				cmd.Flags().Set("copy", tc.copy)

				if err := cmd.RunE(cmd, tc.args); err != nil {
					t.Errorf("Failed writing file to standard output: %v", err)
				}

				// Compare output
				got := buf.String()
				if tc.expected != got {
					t.Errorf("Expected %q, got %q", tc.expected, got)
				}

				// Compare clipboard
				if tc.copy == "true" {
					gotClip, err := clipboard.ReadAll()
					if err != nil {
						t.Errorf("Failed reading from clipboard")
					}

					if tc.expected != gotClip {
						t.Errorf("Expected %q, got %q", tc.expected, gotClip)
					}
				}
			})
		}

	}
}

func touch(db *bolt.DB, rootDir string) func(*testing.T) {
	return func(t *testing.T) {
		rootDir, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}

		cases := []struct {
			desc  string
			names []string
			path  string
		}{
			{
				desc:  "Create one file",
				names: []string{"test.txt"},
				path:  "testdata/create/one",
			},
			{
				desc: "Create all files",
				path: "testdata/create",
			},
			{
				desc:  "Create multiple files",
				names: []string{"testAll/file.csv", "testAll/file"}, // Leave ext empty on purpose
				path:  "testdata/create",
			},
			{
				desc:  "Create a directory",
				names: []string{"testdata/subfolder"},
				path:  "testdata/create",
			},
		}

		cmd := touchSubCmd(db)

		for _, tc := range cases {
			t.Run(tc.desc, func(t *testing.T) {
				args := tc.names
				f := cmd.Flags()
				f.Set("path", tc.path)

				if err := cmd.RunE(cmd, args); err != nil {
					t.Errorf("Failed running create: %v", err)
				}

				// Go back to the directory where the test was executed
				if err := os.Chdir(rootDir); err != nil {
					t.Fatal(err)
				}
			})
		}

		if err := os.RemoveAll("testdata/create"); err != nil {
			t.Fatalf("Failed removing the folder containing created files: %v", err)
		}
	}
}

func ls(db *bolt.DB) func(*testing.T) {
	return func(t *testing.T) {
		cases := []struct {
			desc   string
			name   string
			filter string
			pass   bool
		}{
			{
				desc: "List one",
				name: "test.txt",
				pass: true,
			},
			{
				desc:   "Filter by name",
				name:   "test",
				filter: "true",
				pass:   true,
			},
			{
				desc: "List all",
				name: "",
				pass: true,
			},
			{
				desc:   "File does not exist",
				name:   "non-existent",
				filter: "false",
				pass:   false,
			},
			{
				desc:   "No files found",
				name:   "non-existent",
				filter: "true",
				pass:   false,
			},
		}

		cmd := lsSubCmd(db)

		for _, tc := range cases {
			t.Run(tc.desc, func(t *testing.T) {
				args := []string{tc.name}
				f := cmd.Flags()
				f.Set("filter", tc.filter)

				err := cmd.RunE(cmd, args)
				if err != nil && tc.pass {
					t.Errorf("Failed listing file: %v", err)
				}
				if err == nil && !tc.pass {
					t.Error("Expected an error and got nil")
				}
			})
		}
	}
}

func rename(db *bolt.DB) func(*testing.T) {
	return func(t *testing.T) {
		cases := []struct {
			desc    string
			name    string
			newName string
			pass    bool
		}{
			{
				desc:    "Rename",
				name:    "test.txt",
				newName: "renamed-test.txt",
				pass:    true,
			},
			{
				desc:    "Invalid new name",
				name:    "test",
				newName: "",
				pass:    false,
			},
			{
				desc: "File does not exist",
				name: "non-existent",
				pass: false,
			},
		}

		cmd := renameSubCmd(db, nil)

		for _, tc := range cases {
			t.Run(tc.desc, func(t *testing.T) {
				args := []string{tc.name, tc.newName}

				err := cmd.RunE(cmd, args)
				if err != nil && tc.pass {
					t.Errorf("Failed renaming file: %v", err)
				}
				if err == nil && !tc.pass {
					t.Error("Expected an error and got nil")
				}
			})
		}
	}
}

func rm(db *bolt.DB) func(*testing.T) {
	return func(t *testing.T) {
		cases := []struct {
			desc  string
			name  string
			dir   string
			input string
			pass  bool
		}{
			{
				desc:  "Remove",
				name:  "renamed-test.txt",
				dir:   "false",
				input: "y",
				pass:  true,
			},
			{
				desc:  "Remove directory",
				name:  "testall2",
				dir:   "true",
				input: "y",
				pass:  true,
			},
			{
				desc:  "Do not proceed",
				name:  "testAll/file.csv",
				input: "n",
				dir:   "false",
				pass:  true,
			},
			{
				desc: "Invalid name",
				name: "",
				pass: false,
			},
			{
				desc:  "Non existent file",
				name:  "non-existent",
				input: "y",
				pass:  false,
			},
		}

		for _, tc := range cases {
			t.Run(tc.desc, func(t *testing.T) {
				buf := bytes.NewBufferString(tc.input)
				cmd := rmSubCmd(db, buf)

				args := []string{tc.name}
				f := cmd.Flags()
				f.Set("dir", tc.dir)

				err := cmd.RunE(cmd, args)
				if err != nil && tc.pass {
					t.Errorf("Failed removing file %v", err)
				}
				if err == nil && !tc.pass {
					t.Error("Expected an error and got nil")
				}
			})
		}
	}
}

func TestPostRun(t *testing.T) {
	add := addSubCmd(nil)
	f := add.PostRun
	f(add, nil)

	ls := lsSubCmd(nil)
	f2 := ls.PostRun
	f2(ls, nil)

	rm := rmSubCmd(nil, nil)
	f3 := rm.PostRun
	f3(rm, nil)

	touch := touchSubCmd(nil)
	f4 := touch.PostRun
	f4(touch, nil)
}

func assertError(t *testing.T, name, funcName string, err error, pass bool) {
	if err != nil && pass {
		t.Errorf("%s: failed running %s: %v", name, funcName, err)
	}
	if err == nil && !pass {
		t.Errorf("%s: expected an error and got nil", name)
	}
}
