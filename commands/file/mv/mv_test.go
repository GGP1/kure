package mv

import (
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

func TestMvErrors(t *testing.T) {
	db := cmdutil.SetContext(t, "../../../db/testdata/database")
	createFile(t, db, "exists")

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
