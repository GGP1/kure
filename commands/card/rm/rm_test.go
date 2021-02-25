package rm

import (
	"bytes"
	"testing"

	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/card"
	"github.com/GGP1/kure/pb"
)

func TestRm(t *testing.T) {
	db := cmdutil.SetContext(t, "../../../db/testdata/database")

	names := []string{"test", "directory/test"}
	for _, name := range names {
		if err := card.Create(db, &pb.Card{Name: name}); err != nil {
			t.Fatalf("Failed creating %q: %v", name, err)
		}
	}

	cases := []struct {
		desc  string
		name  string
		input string
		dir   string
	}{
		{
			desc:  "Do not proceed",
			name:  "test",
			input: "n",
		},
		{
			desc:  "Remove",
			name:  "test",
			input: "y",
		},
		{
			desc:  "Remove directory",
			name:  "directory",
			input: "y",
			dir:   "true",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			buf := bytes.NewBufferString(tc.input)

			cmd := NewCmd(db, buf)
			cmd.SetArgs([]string{tc.name})
			cmd.Flags().Set("dir", tc.dir)

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
			desc:  "Non existent card",
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

func TestPostRun(t *testing.T) {
	NewCmd(nil, nil).PostRun(nil, nil)
}
