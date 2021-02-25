package ls

import (
	"testing"

	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/card"
	"github.com/GGP1/kure/pb"
)

func TestLs(t *testing.T) {
	db := cmdutil.SetContext(t, "../../../db/testdata/database")

	if err := card.Create(db, &pb.Card{Name: "test"}); err != nil {
		t.Fatalf("Failed creating the card: %v", err)
	}

	cases := []struct {
		desc   string
		name   string
		filter string
		hide   string
	}{
		{
			desc: "List one",
			name: "test",
		},
		{
			desc:   "Filter by name",
			name:   "te",
			filter: "true",
		},
		{
			desc: "List all",
			name: "",
		},
		{
			desc: "List one and hide",
			name: "test",
			hide: "true",
		},
	}

	cmd := NewCmd(db)
	f := cmd.Flags()

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			cmd.SetArgs([]string{tc.name})
			f.Set("filter", tc.filter)
			f.Set("hide", tc.hide)

			if err := cmd.Execute(); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestLsErrors(t *testing.T) {
	db := cmdutil.SetContext(t, "../../../db/testdata/database")

	cases := []struct {
		desc   string
		name   string
		filter string
	}{
		{
			desc:   "Card does not exist",
			name:   "non-existent",
			filter: "false",
		},
		{
			desc:   "No cards found",
			name:   "non-existent",
			filter: "true",
		},
	}

	cmd := NewCmd(db)

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			cmd.SetArgs([]string{tc.name})
			f := cmd.Flags()
			f.Set("filter", tc.filter)

			if err := cmd.Execute(); err == nil {
				t.Error("Expected an error and got nil")
			}
		})
	}
}

func TestPostRun(t *testing.T) {
	NewCmd(nil).PostRun(nil, nil)
}
