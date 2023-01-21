package rm

import (
	"bytes"
	"testing"

	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/card"
	"github.com/GGP1/kure/pb"

	"github.com/stretchr/testify/assert"
)

func TestRm(t *testing.T) {
	db := cmdutil.SetContext(t)

	names := []string{"test", "directory/test"}
	for _, name := range names {
		err := card.Create(db, &pb.Card{Name: name})
		assert.NoErrorf(t, err, "Failed creating %q card", name)
	}

	cases := []struct {
		desc  string
		name  string
		input string
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
			name:  "directory/",
			input: "y",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			buf := bytes.NewBufferString(tc.input)

			cmd := NewCmd(db, buf)
			cmd.SetArgs([]string{tc.name})

			err := cmd.Execute()
			assert.NoError(t, err)
		})
	}
}

func TestRmErrors(t *testing.T) {
	db := cmdutil.SetContext(t)

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

			err := cmd.Execute()
			assert.Error(t, err)
		})
	}
}
