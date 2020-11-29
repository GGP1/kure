package file

import (
	"bytes"
	"os"
	"testing"

	cmdutil "github.com/GGP1/kure/cmd"

	bolt "go.etcd.io/bbolt"
)

func TestFile(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	defer db.Close()

	d, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	currentDir := d

	NewCmd(db)
	t.Run("Add", add(db))
	t.Run("Create", create(db, currentDir))
	t.Run("Ls", ls(db))
	t.Run("Rename", rename(db))
	t.Run("Rm", rm(db))
}

func add(db *bolt.DB) func(*testing.T) {
	return func(t *testing.T) {
		cases := map[string]struct {
			name      string
			buffer    string
			ignore    string
			path      string
			semaphore string
			pass      bool
		}{
			"Add a file": {
				name:      "test",
				buffer:    "0",
				path:      "testdata/test_file.txt",
				semaphore: "1",
				pass:      true,
			},
			"Add all files using a buffer": {
				name:      "testAll",
				buffer:    "8000",
				path:      "testdata/",
				semaphore: "1",
				pass:      true,
			},
			"Add all files using goroutines": {
				name:      "testAll2",
				buffer:    "0",
				path:      "testdata/",
				semaphore: "15",
				pass:      true,
			},
			"Add all files using a buffer and goroutines": {
				name:      "testAll3",
				buffer:    "2000",
				path:      "testdata/",
				semaphore: "20",
				pass:      true,
			},
			"Add all files ignoring subfolders": {
				name:      "testAll4",
				buffer:    "0",
				ignore:    "true",
				path:      "testdata/",
				semaphore: "2",
				pass:      true,
			},
			"Invalid name": {
				name: "",
				pass: false,
			},
			"Invalid path": {
				name: "test",
				path: "",
				pass: false,
			},
			"Name too long": {
				name: "01234567890123456789012345678901234567890123456789",
				path: "fail",
				pass: false,
			},
		}

		cmd := addSubCmd(db)
		f := cmd.Flags()

		for k, tc := range cases {
			args := []string{tc.name}
			f.Set("buffer", tc.buffer)
			f.Set("ignore", tc.ignore)
			f.Set("path", tc.path)
			f.Set("semaphore", tc.semaphore)

			err := cmd.RunE(cmd, args)
			assertError(t, k, "add", err, tc.pass)

			cmd.ResetFlags()
		}
	}
}

func create(db *bolt.DB, currentDir string) func(*testing.T) {
	return func(t *testing.T) {
		// Touch flag returns an error (no way to pass []string as a flag)
		cases := map[string]struct {
			name  string
			path  string
			touch string
		}{
			"Create one file":       {name: "test.txt", path: "testdata/create/one"},
			"Create all files":      {path: "testdata/create"},
			"Create multiple files": {path: "testdata/create", touch: "testAll/file.csv"},
		}

		cmd := touchSubCmd(db)
		f := cmd.Flags()

		for k, tc := range cases {
			args := []string{tc.name}
			f.Set("path", tc.path)
			f.Set("touch", tc.touch)

			if err := cmd.RunE(cmd, args); err != nil {
				t.Errorf("%s: failed running create: %v", k, err)
			}

			// Go back to the directory where the test was executed
			if err := os.Chdir(currentDir); err != nil {
				t.Fatal(err)
			}

			cmd.ResetFlags()
		}

		if err := os.RemoveAll("testdata/create"); err != nil {
			t.Fatalf("Failed removing the folder containing created files: %v", err)
		}
	}
}

func ls(db *bolt.DB) func(*testing.T) {
	return func(t *testing.T) {
		cases := map[string]struct {
			name   string
			filter string
			pass   bool
		}{
			"List one": {
				name: "test.txt",
				pass: true,
			},
			"Filter by name": {
				name:   "test",
				filter: "true",
				pass:   true,
			},
			"List all": {
				name: "",
				pass: true,
			},
			"Wallet does not exist": {
				name:   "non-existent",
				filter: "false",
				pass:   false,
			},
			"No files found": {
				name:   "non-existent",
				filter: "true",
				pass:   false,
			},
		}

		cmd := lsSubCmd(db)
		f := cmd.Flags()

		for k, tc := range cases {
			args := []string{tc.name}
			f.Set("filter", tc.filter)

			err := cmd.RunE(cmd, args)
			assertError(t, k, "ls", err, tc.pass)
		}
	}
}

func rm(db *bolt.DB) func(*testing.T) {
	return func(t *testing.T) {
		cases := map[string]struct {
			name  string
			dir   string
			input string
			pass  bool
		}{
			"Remove": {
				name:  "renamed test",
				dir:   "true",
				input: "y",
				pass:  true,
			},
			"Do not proceed": {
				name:  "quit",
				input: "n",
				pass:  true,
			},
			"Invalid name": {name: "",
				pass: false,
			},
			"Non existent file": {
				name:  "non-existent",
				input: "y",
				pass:  false,
			},
		}

		for k, tc := range cases {
			buf := bytes.NewBufferString(tc.input)

			cmd := rmSubCmd(db, buf)
			f := cmd.Flags()

			args := []string{tc.name}
			f.Set("dir", tc.dir)

			err := cmd.RunE(cmd, args)
			assertError(t, k, "rm", err, tc.pass)
		}
	}
}

func rename(db *bolt.DB) func(*testing.T) {
	return func(t *testing.T) {
		cases := map[string]struct {
			name  string
			input string
			pass  bool
		}{
			"Rename":              {name: "test.txt", input: "renamed test", pass: true},
			"Invalid new name":    {name: "test", input: "", pass: false},
			"File does not exist": {name: "non-existent", pass: false},
		}

		for k, tc := range cases {
			buf := bytes.NewBufferString(tc.input)

			cmd := renameSubCmd(db, buf)
			args := []string{tc.name}

			err := cmd.RunE(cmd, args)
			assertError(t, k, "rename", err, tc.pass)
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
