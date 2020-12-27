package ls

import (
	"testing"

	cmdutil "github.com/GGP1/kure/cmd"
	"github.com/GGP1/kure/db/entry"
	"github.com/GGP1/kure/pb"
)

func TestLs(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	defer db.Close()

	lockedBuf, e := pb.SecureEntry()
	e.Name = "test"
	e.Password = "testing"
	e.Expires = "Never"

	if err := entry.Create(db, lockedBuf, e); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		desc   string
		name   string
		filter string
		hide   string
		qr     string
		pass   bool
	}{
		{
			desc: "List one",
			name: "test",
			pass: true,
		},
		{
			desc: "List one and show qr",
			name: "test",
			qr:   "true",
			pass: true,
		},
		{
			desc:   "Filter by name",
			name:   "test",
			filter: "true",
			pass:   true,
		},
		{
			desc: "List all",
			name: "",
			pass: true,
		},
		{
			desc: "List one and hide",
			name: "test",
			hide: "true",
			pass: true,
		},
		{
			desc:   "Entry does not exist",
			name:   "non-existent",
			filter: "false",
			pass:   false,
		},
		{
			desc:   "No entries found",
			name:   "non-existent",
			filter: "true",
			pass:   false,
		},
	}

	cmd := NewCmd(db)

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			args := []string{tc.name}

			f := cmd.Flags()
			f.Set("filter", tc.filter)
			f.Set("hide", tc.hide)
			f.Set("qr", tc.qr)

			err := cmd.RunE(cmd, args)
			if err != nil && tc.pass {
				t.Errorf("Failed running ls: %v", err)
			}
			if err == nil && !tc.pass {
				t.Error("Expected an error and got nil")
			}
		})
	}
}

func TestPostRun(t *testing.T) {
	cmd := NewCmd(nil)
	f := cmd.PostRun
	f(cmd, nil)
}
