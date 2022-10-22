package copy

import (
	"strconv"
	"testing"

	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/config"
	"github.com/GGP1/kure/db/entry"
	"github.com/GGP1/kure/pb"

	"github.com/atotto/clipboard"
	"github.com/stretchr/testify/assert"
	bolt "go.etcd.io/bbolt"
)

func TestCopy(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	e := createEntry(t, db)

	cases := []struct {
		desc         string
		value        string
		timeout      string
		copyUsername bool
	}{
		{
			desc:  "Copy password",
			value: e.Password,
		},
		{
			desc:         "Copy username",
			value:        e.Username,
			copyUsername: true,
		},
		{
			desc:    "Copy with timeout",
			value:   "",
			timeout: "1ms",
		},
	}

	cmd := NewCmd(db)
	f := cmd.Flags()

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			cmd.SetArgs([]string{e.Name})
			f.Set("timeout", tc.timeout)
			f.Set("username", strconv.FormatBool(tc.copyUsername))

			err := cmd.Execute()
			assert.NoError(t, err)

			got, err := clipboard.ReadAll()
			assert.NoError(t, err, "Failed reading from clipboard")

			assert.Equal(t, tc.value, got)
		})
	}
}

func TestCopyWithConfigTimeout(t *testing.T) {
	if clipboard.Unsupported {
		t.Skip("No clipboard utilities available")
	}
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	e := createEntry(t, db)

	config.Set("clipboard.timeout", "1ns")
	cmd := NewCmd(db)
	cmd.SetArgs([]string{e.Name})

	err := cmd.Execute()
	assert.NoError(t, err, "Failed to copy password to clipboard")

	got, err := clipboard.ReadAll()
	assert.NoError(t, err, "Failed reading from clipboard")

	assert.Empty(t, got)
}

func TestCopyErrors(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")

	cases := []struct {
		desc string
		name string
	}{
		{desc: "Non-existent", name: "non-existent"},
		{desc: "Invalid name", name: ""},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			cmd := NewCmd(db)
			cmd.SetArgs([]string{tc.name})

			err := cmd.Execute()
			assert.Error(t, err)
		})
	}
}

func TestPostRun(t *testing.T) {
	NewCmd(nil).PostRun(nil, nil)
}

func createEntry(t *testing.T, db *bolt.DB) *pb.Entry {
	t.Helper()

	e := &pb.Entry{
		Name:     "test",
		Username: "Go",
		Password: "Gopher",
		Expires:  "Never",
	}

	err := entry.Create(db, e)
	assert.NoError(t, err)

	return e
}
