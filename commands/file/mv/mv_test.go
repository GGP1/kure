package mv

import (
	"strconv"
	"strings"
	"testing"

	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/file"
	"github.com/GGP1/kure/pb"

	bolt "go.etcd.io/bbolt"
)

func TestMv(t *testing.T) {
	db := cmdutil.SetContext(t, "../../../db/testdata/database")

	oldName := "test.txt"
	newName := "renamed-test" // omit extension on purpose
	createFile(t, db, oldName)

	cmd := NewCmd(db)
	cmd.SetArgs([]string{oldName, newName})

	if err := cmd.Execute(); err != nil {
		t.Error(err)
	}

	if _, err := file.GetCheap(db, newName+".txt"); err != nil {
		t.Errorf("Failed getting the renamed file: %v", err)
	}
}

func TestMvDir(t *testing.T) {
	db := cmdutil.SetContext(t, "../../../db/testdata/database")

	oldDir := "directory/"
	for i := 0; i < 3; i++ {
		createFile(t, db, oldDir+strconv.Itoa(i))
	}

	newDir := "folder/"
	cmd := NewCmd(db)
	cmd.SetArgs([]string{oldDir, newDir})

	if err := cmd.Execute(); err != nil {
		t.Error(err)
	}

	names, err := file.ListNames(db)
	if err != nil {
		t.Error(err)
	}

	for _, name := range names {
		if !strings.HasPrefix(name, newDir) {
			t.Errorf("%q wasn't moved into %q", name, newDir)
		}
	}
}

func TestMvErrors(t *testing.T) {
	db := cmdutil.SetContext(t, "../../../db/testdata/database")
	createFile(t, db, "exists")
	dir := "dir/"
	createFile(t, db, dir+"1")

	cases := []struct {
		desc    string
		oldName string
		newName string
	}{
		{
			desc:    "Invalid new name",
			oldName: "test",
			newName: "",
		},
		{
			desc:    "New name already exists",
			oldName: "test",
			newName: "exists",
		},
		{
			desc:    "File does not exist",
			oldName: "non-existent",
			newName: "test.txt",
		},
		{
			desc:    "Dir into file",
			oldName: dir,
			newName: "test.txt",
		},
	}

	cmd := NewCmd(db)

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			cmd.SetArgs([]string{tc.oldName, tc.newName})
			if err := cmd.Execute(); err == nil {
				t.Error("Expected an error and got nil")
			}
		})
	}
}

func createFile(t *testing.T, db *bolt.DB, name string) {
	t.Helper()

	if err := file.Create(db, &pb.File{Name: name}); err != nil {
		t.Fatalf("Failed creating the file: %v", err)
	}
}
