package rm

import (
	"bytes"
	"testing"

	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/totp"
	"github.com/GGP1/kure/pb"
)

func TestRm(t *testing.T) {
	db := cmdutil.SetContext(t, "../../../db/testdata/database")

	if err := totp.Create(db, &pb.TOTP{Name: "test"}); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		desc         string
		name         string
		confirmation string
	}{
		{
			desc:         "Abort",
			name:         "test",
			confirmation: "n",
		},
		{
			desc:         "Proceed",
			name:         "test",
			confirmation: "y",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			buf := bytes.NewBufferString(tc.confirmation)
			cmd := NewCmd(db, buf)
			cmd.SetArgs([]string{tc.name})

			if err := cmd.Execute(); err != nil {
				t.Errorf("Failed removing the TOTP: %v", err)
			}

		})
	}
}

func TestRmErrors(t *testing.T) {
	db := cmdutil.SetContext(t, "../../../db/testdata/database")

	cases := []struct {
		desc         string
		name         string
		confirmation string
	}{
		{
			desc: "Invalid name",
			name: "",
		},
		{
			desc:         "Does not exists",
			name:         "non-existent",
			confirmation: "y",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			buf := bytes.NewBufferString(tc.confirmation)
			cmd := NewCmd(db, buf)
			cmd.SetArgs([]string{tc.name})

			if err := cmd.Execute(); err == nil {
				t.Error("Expected an error and got nil")
			}
		})
	}
}
