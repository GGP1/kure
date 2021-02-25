package add

import (
	"bytes"
	"testing"

	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/entry"
	"github.com/GGP1/kure/pb"

	bolt "go.etcd.io/bbolt"
)

func TestAdd(t *testing.T) {
	db := cmdutil.SetContext(t, "../../../db/testdata/database")
	createEntries(t, db, "6-digit", "7-digit", "8-digit")

	cases := []struct {
		desc   string
		name   string
		digits string
	}{
		{
			desc:   "TOTP with 6 digits",
			name:   "6-digit",
			digits: "6",
		},
		{
			desc:   "TOTP with 7 digits",
			name:   "7-digit",
			digits: "7",
		},
		{
			desc:   "TOTP with 8 digits",
			name:   "8-digit",
			digits: "8",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			buf := bytes.NewBufferString("IFGEWRKSIFJUMR2R")
			cmd := NewCmd(db, buf)
			cmd.Flags().Set("digits", tc.digits)
			cmd.SetArgs([]string{tc.name})

			if err := cmd.Execute(); err != nil {
				t.Errorf("Failed adding TOTP: %v", err)
			}
		})
	}
}

func TestAddErrors(t *testing.T) {
	db := cmdutil.SetContext(t, "../../../db/testdata/database")

	createEntries(t, db, "fail")

	cases := []struct {
		desc   string
		key    string
		name   string
		digits string
	}{
		{
			desc: "Invalid name",
			name: "",
		},
		{
			desc:   "Invalid digits",
			name:   "fail",
			digits: "10",
		},
		{
			desc:   "Entry does not exist",
			name:   "non-existent",
			digits: "6",
		},
		{
			desc:   "Invalid key",
			name:   "fail",
			digits: "7",
			key:    "invalid/%",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			buf := bytes.NewBufferString(tc.key)
			cmd := NewCmd(db, buf)
			cmd.Flags().Set("digits", tc.digits)
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

func createEntries(t *testing.T, db *bolt.DB, names ...string) {
	t.Helper()

	for _, name := range names {
		e := &pb.Entry{
			Name:    name,
			Expires: "Never",
		}
		if err := entry.Create(db, e); err != nil {
			t.Fatal(err)
		}
	}
}
