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

	names := []string{"test", "directory/test", "kure", "atoll"}
	for _, name := range names {
		err := card.Create(db, &pb.Card{Name: name})
		assert.NoErrorf(t, err, "Failed creating %q", name)
	}

	cases := []struct {
		desc  string
		input string
		names []string
	}{
		{
			desc:  "Do not proceed",
			names: []string{"test"},
			input: "n",
		},
		{
			desc:  "Remove one card",
			names: []string{"test"},
			input: "y",
		},
		{
			desc:  "Remove multiple cards",
			names: []string{"kure", "atoll"},
			input: "y",
		},
		{
			desc:  "Remove directory",
			names: []string{"directory/"},
			input: "y",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			buf := bytes.NewBufferString(tc.input)
			cmd := NewCmd(db, buf)
			cmd.SetArgs(tc.names)

			err := cmd.Execute()
			assert.NoError(t, err)

			if tc.input == "y" {
				for _, name := range tc.names {
					_, err := card.Get(db, name)
					assert.Error(t, err)
				}
			}
		})
	}
}

func TestRmErrors(t *testing.T) {
	db := cmdutil.SetContext(t)

	name := "random"
	err := card.Create(db, &pb.Card{Name: name})
	assert.NoErrorf(t, err, "Failed creating %q", name)

	cases := []struct {
		desc         string
		confirmation string
		names        []string
	}{
		{
			desc:  "Invalid name",
			names: []string{""},
		},
		{
			desc:         "Does not exists",
			names:        []string{"non-existent"},
			confirmation: "y",
		},
		{
			desc:  "Second name does not exist",
			names: []string{"random", "non-existent"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			buf := bytes.NewBufferString(tc.confirmation)
			cmd := NewCmd(db, buf)
			cmd.SetArgs(tc.names)

			err := cmd.Execute()
			assert.Error(t, err)
		})
	}
}
