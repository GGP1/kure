package add

import (
	"bytes"
	"fmt"
	"testing"

	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/file"
	"github.com/GGP1/kure/pb"

	"github.com/stretchr/testify/assert"
)

func TestAdd(t *testing.T) {
	db := cmdutil.SetContext(t)

	cases := []struct {
		desc      string
		name      string
		ignore    string
		path      string
		semaphore string
	}{
		{
			desc:      "Add a file",
			name:      "test",
			path:      "../testdata/test_file.txt",
			semaphore: "1",
		},
		{
			desc:      "Add all files synchronously",
			name:      "testAll",
			path:      "../testdata/",
			semaphore: "1",
		},
		{
			desc:      "Add all files using goroutines",
			name:      "testAll2",
			path:      "../testdata/",
			semaphore: "8",
		},
		{
			desc:      "Add all files ignoring subfolders",
			name:      "testAll3",
			ignore:    "true",
			path:      "../testdata/",
			semaphore: "3",
		},
	}

	cmd := NewCmd(db, nil)

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			cmd.SetArgs([]string{tc.name})
			f := cmd.Flags()
			f.Set("ignore", tc.ignore)
			f.Set("path", tc.path)
			f.Set("semaphore", tc.semaphore)

			var stderr bytes.Buffer
			cmd.SetErr(&stderr)

			err := cmd.Execute()
			assert.NoError(t, err)

			assert.LessOrEqual(t, stderr.Len(), 0, "Expected nothing on stderr")
		})
	}

	t.Run("Add note", func(t *testing.T) {
		name := "test-notes"
		expectedContent := []byte("note content")
		buf := bytes.NewBufferString("note content<\n")

		cmd := NewCmd(db, buf)
		cmd.SetArgs([]string{name})
		cmd.Flags().Set("note", "true")

		err := cmd.Execute()
		assert.NoError(t, err, "Failed creating the note")

		file, err := file.Get(db, fmt.Sprintf("notes/%s.txt", name))
		assert.NoError(t, err, "Couldn't find the note")

		assert.Equal(t, file.Content, expectedContent)
	})
}

func TestAddErrors(t *testing.T) {
	db := cmdutil.SetContext(t)

	err := file.Create(db, &pb.File{Name: "already exists.txt"})
	assert.NoError(t, err, "Failed creating the file")

	cases := []struct {
		desc      string
		name      string
		path      string
		semaphore string
	}{
		{
			desc: "Invalid name",
			name: "",
		},
		{
			desc:      "Invalid semaphore",
			name:      "test",
			path:      "../testdata/test_file.txt",
			semaphore: "0",
		},
		{
			desc:      "Already exists",
			name:      "already exists.txt",
			semaphore: "1",
			path:      "../testdata/test_file.txt",
		},
		{
			desc:      "Non-existent",
			name:      "non-existent",
			semaphore: "1",
			path:      "../testdata/non-existent.txt",
		},
		{
			desc:      "Empty directory",
			name:      "empty-dir",
			semaphore: "1",
			path:      "../testdata/empty",
		},
	}

	cmd := NewCmd(db, nil)

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			cmd.SetArgs([]string{tc.name})
			f := cmd.Flags()
			f.Set("path", tc.path)
			f.Set("semaphore", tc.semaphore)

			err := cmd.Execute()
			assert.Error(t, err)
		})
	}
}

func TestPostRun(t *testing.T) {
	NewCmd(nil, nil).PostRun(nil, nil)
}
