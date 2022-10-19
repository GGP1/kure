package copy

import (
	"strconv"
	"testing"

	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/card"
	"github.com/GGP1/kure/pb"

	"github.com/atotto/clipboard"
)

func TestCopy(t *testing.T) {
	if clipboard.Unsupported {
		t.Skip("No clipboard utilities available")
	}
	db := cmdutil.SetContext(t, "../../../db/testdata/database")

	c := &pb.Card{
		Name:         "test",
		Number:       "1826352187",
		SecurityCode: "213",
	}
	if err := card.Create(db, c); err != nil {
		t.Fatalf("Failed creating the card: %v", err)
	}

	cases := []struct {
		desc    string
		value   string
		timeout string
		copyCVC bool
	}{
		{
			desc:  "Copy number",
			value: c.Number,
		},
		{
			desc:    "Copy CVC",
			value:   c.SecurityCode,
			copyCVC: true,
		},
		{
			desc:    "Copy w/Timeout",
			value:   "",
			timeout: "1ns",
		},
	}

	cmd := NewCmd(db)
	f := cmd.Flags()

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			cmd.SetArgs([]string{c.Name})
			f.Set("timeout", tc.timeout)
			f.Set("cvc", strconv.FormatBool(tc.copyCVC))

			if err := cmd.Execute(); err != nil {
				t.Error(err)
			}

			got, err := clipboard.ReadAll()
			if err != nil {
				t.Fatalf("Failed reading from clipboard: %v", err)
			}

			if got != tc.value {
				t.Errorf("Expected %s, got %s", tc.value, got)
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
