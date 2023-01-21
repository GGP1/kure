package cat

import (
	"bytes"
	"testing"

	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/file"
	"github.com/GGP1/kure/pb"

	"github.com/atotto/clipboard"
	"github.com/stretchr/testify/assert"
	bolt "go.etcd.io/bbolt"
)

func TestCat(t *testing.T) {
	db := cmdutil.SetContext(t)

	name1 := "test.txt"
	name2 := "testall2/subfolder/file.txt"
	createFiles(t, db, name1, name2)

	cases := []struct {
		desc     string
		expected string
		copy     string
		args     []string
	}{
		{
			desc:     "Read files",
			args:     []string{name1, name2},
			expected: "test\ntesting file\n",
			copy:     "false",
		},
		{
			desc:     "Read and copy to clipboard",
			args:     []string{name1},
			expected: "test\n",
			copy:     "true",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			var buf bytes.Buffer
			cmd := NewCmd(db, &buf)
			cmd.SetArgs(tc.args)
			cmd.Flags().Set("copy", tc.copy)

			if clipboard.Unsupported && tc.copy == "true" {
				t.Skip("No clipboard utilities available")
			}

			err := cmd.Execute()
			assert.NoError(t, err)

			// Compare output
			got := buf.String()
			assert.Equal(t, tc.expected, got)

			// Compare clipboard
			if tc.copy == "true" {
				gotClip, err := clipboard.ReadAll()
				assert.NoError(t, err, "Failed reading form the clipboard")

				assert.Equal(t, tc.expected, gotClip)
			}
		})
	}
}

func TestCatErrors(t *testing.T) {
	db := cmdutil.SetContext(t)

	cases := []struct {
		desc string
		args []string
	}{
		{
			desc: "Invalid name",
			args: []string{""},
		},
		{
			desc: "Non-existent",
			args: []string{"non-existent"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			cmd := NewCmd(db, nil)
			cmd.SetArgs(tc.args)

			err := cmd.Execute()
			assert.Error(t, err)
		})
	}
}

func TestPostRun(t *testing.T) {
	NewCmd(nil, nil).PostRun(nil, nil)
}

func createFiles(t *testing.T, db *bolt.DB, name1, name2 string) {
	f1 := &pb.File{
		Name:    name1,
		Content: []byte("test"),
	}
	err := file.Create(db, f1)
	assert.NoError(t, err, "Failed creating first file")
	f2 := &pb.File{
		Name:    name2,
		Content: []byte("testing file"),
	}
	err = file.Create(db, f2)
	assert.NoError(t, err, "Failed creating second file")
}
