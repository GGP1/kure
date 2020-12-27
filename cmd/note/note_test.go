package note

import (
	"bytes"
	"os"
	"reflect"
	"testing"

	cmdutil "github.com/GGP1/kure/cmd"
	"github.com/GGP1/kure/pb"

	bolt "go.etcd.io/bbolt"
)

func TestNote(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	defer db.Close()

	NewCmd(db, nil)
	t.Run("Add", add(db))
	t.Run("Copy", copy(db))
	t.Run("Ls", ls(db))
	t.Run("Rm", rm(db))
}

func add(db *bolt.DB) func(*testing.T) {
	return func(t *testing.T) {
		cases := []struct {
			desc string
			name string
			pass bool
		}{
			{
				desc: "Add",
				name: "test",
				pass: true,
			},
			{
				desc: "Add2",
				name: "test2",
				pass: true,
			},
			{
				desc: "Add within directory",
				name: "directory/test",
				pass: true,
			},
			{
				desc: "Already exists",
				name: "test",
				pass: false,
			},
			{
				desc: "Invalid name",
				name: "",
				pass: false,
			},
		}

		cmd := addSubCmd(db, os.Stdin)

		for _, tc := range cases {
			t.Run(tc.desc, func(t *testing.T) {
				args := []string{tc.name}

				err := cmd.RunE(cmd, args)
				assertError(t, "add", err, tc.pass)
			})
		}
	}
}

func copy(db *bolt.DB) func(*testing.T) {
	return func(t *testing.T) {
		cases := []struct {
			desc    string
			name    string
			timeout string
			pass    bool
		}{
			{
				desc: "Copy",
				name: "test",
				pass: true,
			},
			{
				desc:    "Copy w/Timeout",
				name:    "test",
				timeout: "30ms",
				pass:    true,
			},
			{
				desc: "Invalid name",
				name: "",
				pass: false,
			},
			{
				desc: "Non existent note",
				name: "non-existent",
				pass: false,
			},
		}

		cmd := copySubCmd(db)
		f := cmd.Flags()

		for _, tc := range cases {
			t.Run(tc.desc, func(t *testing.T) {
				args := []string{tc.name}
				if tc.timeout != "" {
					f.Set("timeout", tc.timeout)
				}

				err := cmd.RunE(cmd, args)
				assertError(t, "copy", err, tc.pass)
			})
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
				name: "test",
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
				desc: "List one and hide",
				name: "test",
				pass: true,
			},
			{
				desc:   "Note does not exist",
				name:   "non-existent",
				filter: "false",
				pass:   false,
			},
			{
				desc:   "No notes found",
				name:   "non-existent",
				filter: "true",
				pass:   false,
			},
		}

		cmd := lsSubCmd(db)
		f := cmd.Flags()

		for _, tc := range cases {
			t.Run(tc.desc, func(t *testing.T) {
				args := []string{tc.name}
				f.Set("filter", tc.filter)

				err := cmd.RunE(cmd, args)
				assertError(t, "ls", err, tc.pass)
			})
		}
	}
}

func rm(db *bolt.DB) func(*testing.T) {
	return func(t *testing.T) {
		cases := []struct {
			desc  string
			name  string
			input string
			dir   string
			pass  bool
		}{
			{
				desc:  "Remove",
				name:  "test",
				input: "y",
				pass:  true,
			},
			{
				desc:  "Remove directory",
				name:  "directory",
				input: "y",
				dir:   "true",
				pass:  true,
			},
			{
				desc:  "Do not proceed",
				name:  "test2",
				input: "n",
				pass:  true,
			},
			{
				desc: "Invalid name",
				name: "",
				pass: false,
			},
			{
				desc:  "Non existent note",
				name:  "non-existent",
				input: "y",
				pass:  false,
			},
		}

		for _, tc := range cases {
			t.Run(tc.desc, func(t *testing.T) {
				buf := bytes.NewBufferString(tc.input)

				cmd := rmSubCmd(db, buf)
				cmd.Flags().Set("dir", tc.dir)
				args := []string{tc.name}

				err := cmd.RunE(cmd, args)
				assertError(t, "rm", err, tc.pass)
			})
		}
	}
}

func TestInput(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	defer db.Close()

	expected := &pb.Note{
		Name: "test",
		Text: "text",
	}

	buf := bytes.NewBufferString("text!end\n")

	_, got, err := input(db, "test", buf)
	if err != nil {
		t.Fatalf("Failed creating the note: %v", err)
	}

	if !reflect.DeepEqual(expected, got) {
		t.Errorf("Expected %v, got %v", expected, got)
	}
}

func TestPostRun(t *testing.T) {
	copy := copySubCmd(nil)
	f := copy.PostRun
	f(copy, nil)

	ls := lsSubCmd(nil)
	f2 := ls.PostRun
	f2(ls, nil)

	rm := rmSubCmd(nil, nil)
	f3 := rm.PostRun
	f3(ls, nil)
}

func assertError(t *testing.T, funcName string, err error, pass bool) {
	if err != nil && pass {
		t.Errorf("Failed running %s: %v", funcName, err)
	}
	if err == nil && !pass {
		t.Error("Expected an error and got nil")
	}
}
