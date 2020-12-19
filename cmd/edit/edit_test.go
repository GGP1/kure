package edit

import (
	"bytes"
	"testing"

	cmdutil "github.com/GGP1/kure/cmd"
	"github.com/GGP1/kure/db/entry"
	"github.com/GGP1/kure/pb"
)

func TestEdit(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	defer db.Close()

	if err := entry.Create(db, &pb.Entry{Name: "test", Expires: "Never"}); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		desc     string
		name     string
		nameFlag string
		input    string
		pass     bool
	}{
		{
			desc:     "Edit",
			name:     "test",
			nameFlag: "true",
			input:    "testname\n\npassword\n-\n!end\n26/02/2021",
			pass:     true,
		},
		{
			desc: "Invalid name",
			name: "non-existent",
			pass: false,
		},
		{
			desc:     "Invalid new name",
			name:     "testname",
			nameFlag: "true",
			input:    "\nusername\npassword\ntestedit.com\n!end\n26/02/2021",
			pass:     false,
		},
		{
			desc:  "Invalid expiration format",
			name:  "testname",
			input: "-\n\npassword\ntestedit.com\n!end\ninvalid format",
			pass:  false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			buf := bytes.NewBufferString(tc.input)

			cmd := NewCmd(db, buf)
			f := cmd.Flags()

			args := []string{tc.name}
			f.Set("name", tc.nameFlag)

			err := cmd.RunE(cmd, args)
			assertError(t, "edit", err, tc.pass)
		})
	}

	if _, err := entry.Get(db, "testname"); err != nil {
		t.Errorf("Expected to get the edited entry but failed: %v", err)
	}
}

func TestPostRun(t *testing.T) {
	cmd := NewCmd(nil, nil)
	f := cmd.PostRun
	f(cmd, nil)
}

func assertError(t *testing.T, funcName string, err error, pass bool) {
	if err != nil && pass {
		t.Errorf("Failed running %s: %v", funcName, err)
	}
	if err == nil && !pass {
		t.Error("Expected an error and got nil")
	}
}
