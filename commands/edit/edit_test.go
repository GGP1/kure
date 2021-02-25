package edit

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/entry"
	"github.com/GGP1/kure/pb"

	"github.com/spf13/viper"
	bolt "go.etcd.io/bbolt"
)

func TestEditErrors(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	createEntry(t, db, "test")

	cases := []struct {
		desc string
		name string
		it   string
		set  func()
	}{
		{
			desc: "Invalid name",
			name: "",
			set:  func() {},
		},
		{
			desc: "Executable not found",
			name: "test",
			it:   "true",
			set: func() {
				viper.Set("editor", "non-existent")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			tc.set()
			cmd := NewCmd(db)
			cmd.SetArgs([]string{tc.name})
			f := cmd.Flags()
			f.Set("it", tc.it)

			if err := cmd.Execute(); err == nil {
				t.Error("Expected an error and got nil")
			}
		})
	}
}

func TestCreateTempFile(t *testing.T) {
	e := &pb.Entry{
		Name:    "test-create-file",
		Expires: "Never",
	}

	filename, err := createTempFile(e)
	if err != nil {
		t.Fatalf("Failed creating the file: %v", err)
	}
	defer os.Remove(filename)

	content, err := os.ReadFile(filename)
	if err != nil {
		t.Errorf("Failed reading the file: %v", err)
	}

	var got pb.Entry
	if err := json.Unmarshal(content, &got); err != nil {
		t.Errorf("Failed decoding the file: %v", err)
	}

	if !reflect.DeepEqual(e, &got) {
		t.Error("Expected entries to be deep equal")
	}
}

func TestReadTmpFile(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		if _, err := readTmpFile("testdata/test_read.json"); err != nil {
			t.Error(err)
		}
	})

	t.Run("Errors", func(t *testing.T) {
		dir, _ := os.Getwd()
		os.Chdir("testdata")
		// Go back to the initial directory
		defer os.Chdir(dir)

		cases := []struct {
			desc     string
			filename string
		}{
			{
				desc:     "Does not exists",
				filename: "does_not_exists.json",
			},
			{
				desc:     "EOF",
				filename: "test_read_EOF.json",
			},
		}

		for _, tc := range cases {
			t.Run(tc.desc, func(t *testing.T) {
				if _, err := readTmpFile(tc.filename); err == nil {
					t.Error("Expected an error and got nil")
				}
			})
		}
	})
}

func TestUpdateEntry(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	name := "test_update"
	createEntry(t, db, name)

	newName := "new_name"
	newEntry := &pb.Entry{
		Name:     newName,
		Username: "test",
		Password: "q8rvq63r/q",
		URL:      "https://www.github.com/GGP1/kure",
		Expires:  "02/12/2023",
		Notes:    "",
	}

	if err := updateEntry(db, name, newEntry); err != nil {
		t.Error(err)
	}

	e, err := entry.Get(db, newName)
	if err != nil {
		t.Error(err)
	}

	if reflect.DeepEqual(e, newEntry) {
		t.Errorf("Expected %v, got %v", newEntry, e)
	}

	t.Run("Invalid name", func(t *testing.T) {
		newEntry.Name = ""
		if err := updateEntry(db, "fail", newEntry); err == nil {
			t.Error("Expected an error and got nil")
		}
	})
}

func TestPostRun(t *testing.T) {
	NewCmd(nil).PostRun(nil, nil)
}

func createEntry(t *testing.T, db *bolt.DB, name string) {
	t.Helper()
	e := &pb.Entry{
		Name:    name,
		Expires: "Never",
	}

	if err := entry.Create(db, e); err != nil {
		t.Fatalf("Failed creating the entry: %v", err)
	}
}
