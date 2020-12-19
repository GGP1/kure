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

	if err := entry.Create(db, &pb.Entry{Name: "test", Password: "top secret", Expires: "Never"}); err != nil {
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
	f := cmd.Flags()

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			args := []string{tc.name}
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

			cmd.ResetFlags()
		})
	}
}

func TestPostRun(t *testing.T) {
	cmd := NewCmd(nil)
	f := cmd.PostRun
	f(cmd, nil)
}
