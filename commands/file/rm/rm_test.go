package rm

import (
	"bytes"
	"testing"

	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/file"
	"github.com/GGP1/kure/pb"
)

func TestRm(t *testing.T) {
	db := cmdutil.SetContext(t, "../../../db/testdata/database")

	names := []string{"test.txt", "directory/test.txt"}
	for _, name := range names {
		if err := file.Create(db, &pb.File{Name: name}); err != nil {
			t.Fatalf("Failed creating %q: %v", name, err)
		}
	}

	cases := []struct {
		desc  string
		name  string
		input string
	}{
		{
			desc:  "Do not proceed",
			name:  "test.txt",
			input: "n",
		},
		{
			desc:  "Remove",
			name:  "test.txt",
			input: "y",
		},
		{
			desc:  "Remove directory",
			name:  "directory/",
			input: "y",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			buf := bytes.NewBufferString(tc.input)
			cmd := NewCmd(db, buf)
			cmd.SetArgs([]string{tc.name})

			if err := cmd.Execute(); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestRmErrors(t *testing.T) {
	db := cmdutil.SetContext(t, "../../../db/testdata/database")

	cases := []struct {
		desc  string
		name  string
		input string
	}{
		{
			desc: "Invalid name",
			name: "",
		},
		{
			desc:  "Non existent file",
			name:  "non-existent",
			input: "y",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			buf := bytes.NewBufferString(tc.input)
			cmd := NewCmd(db, buf)
			cmd.SetArgs([]string{tc.name})

			if err := cmd.Execute(); err == nil {
				t.Error("Expected an error and got nil")
			}
		})
	}
}
