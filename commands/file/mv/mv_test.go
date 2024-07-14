package mv

import (
	"runtime"
	"strconv"
	"strings"
	"testing"

	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/file"
	"github.com/GGP1/kure/pb"

	"github.com/stretchr/testify/assert"
	bolt "go.etcd.io/bbolt"
)

func TestMv(t *testing.T) {
	db := cmdutil.SetContext(t)

	oldName := "test.txt"
	newName := "renamed-test" // omit extension on purpose
	createFile(t, db, oldName)

	cmd := NewCmd(db)
	cmd.SetArgs([]string{oldName, newName})

	err := cmd.Execute()
	assert.NoError(t, err)

	_, err = file.GetCheap(db, newName+".txt")
	assert.NoError(t, err, "Failed getting the renamed file")
}

func TestMvDir(t *testing.T) {
	db := cmdutil.SetContext(t)

	oldDir := "directory/"
	for i := 0; i < 2; i++ {
		createFile(t, db, oldDir+strconv.Itoa(i))
	}

	newDir := "folder/"
	cmd := NewCmd(db)
	cmd.SetArgs([]string{oldDir, newDir})

	err := cmd.Execute()
	assert.NoError(t, err)

	names, err := file.ListNames(db)
	assert.NoError(t, err)

	for _, name := range names {
		if !strings.HasPrefix(name, newDir) {
			t.Errorf("%q wasn't moved into %q", name, newDir)
		}
	}
}

func TestMvFileIntoDir(t *testing.T) {
	db := cmdutil.SetContext(t)

	filename := "directory/test.csv"
	newDir := "folder/"
	createFile(t, db, filename)

	cmd := NewCmd(db)
	cmd.SetArgs([]string{filename, newDir})

	err := cmd.Execute()
	assert.NoError(t, err)

	newName := newDir + strings.Split(filename, "/")[1]
	_, err = file.GetCheap(db, newName)
	assert.NoError(t, err, "Failed getting the renamed file")
}

func TestMvErrors(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("For some strange reason this test fails in unix systems because the database file throws \"bad file descriptor\"")
	}

	db := cmdutil.SetContext(t)
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
			oldName: "exists",
			newName: "",
		},
		{
			desc:    "New name already exists",
			oldName: "exists",
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
			err := cmd.Execute()
			assert.Error(t, err)
		})
	}
}

func TestMissingArguments(t *testing.T) {
	db := cmdutil.SetContext(t)

	cmd := NewCmd(db)
	cmd.SetArgs([]string{"oldName"})
	err := cmd.Execute()
	assert.Error(t, err)
}

func createFile(t *testing.T, db *bolt.DB, name string) {
	err := file.Create(db, &pb.File{Name: name})
	assert.NoError(t, err, "Failed creating file")
}
