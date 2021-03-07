package copy

import (
	"testing"

	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/config"
	"github.com/GGP1/kure/db/card"
	"github.com/GGP1/kure/pb"

	"github.com/atotto/clipboard"
)

func TestCopy(t *testing.T) {
	if clipboard.Unsupported {
		t.Skip("No clipboard utilities available")
	}
	db := cmdutil.SetContext(t, "../../../db/testdata/database")

	if err := card.Create(db, &pb.Card{Name: "test"}); err != nil {
		t.Fatalf("Failed creating the card: %v", err)
	}

	cases := []struct {
		desc    string
		name    string
		cvc     string
		timeout string
	}{
		{
			desc: "Copy number",
			name: "test",
			cvc:  "false",
		},
		{
			desc: "Copy CVC",
			name: "test",
			cvc:  "true",
		},
		{
			desc:    "Copy w/Timeout",
			name:    "test",
			timeout: "1ns",
		},
	}

	cmd := NewCmd(db)
	f := cmd.Flags()
	config.Set("clipboard.timeout", "1ns") // Set default

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			cmd.SetArgs([]string{tc.name})
			f.Set("timeout", tc.timeout)
			f.Set("cvc", tc.cvc)

			if err := cmd.Execute(); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestCopyErrors(t *testing.T) {
	db := cmdutil.SetContext(t, "../../../db/testdata/database")

	cases := []struct {
		desc string
		name string
	}{
		{
			desc: "Invalid name",
			name: "",
		},
		{
			desc: "Non existent card",
			name: "non-existent",
		},
	}

	cmd := NewCmd(db)

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			cmd.SetArgs([]string{tc.name})

			if err := cmd.Execute(); err == nil {
				t.Error("Expected an error and got nil")
			}
		})
	}

}

func TestPostRun(t *testing.T) {
	NewCmd(nil).PostRun(nil, nil)
}
