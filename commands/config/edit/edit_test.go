package edit

import (
	"testing"

	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/config"
)

func TestEditErrors(t *testing.T) {
	db := cmdutil.SetContext(t, "../../../db/testdata/database")

	cases := []struct {
		desc string
		path string
	}{
		{
			desc: "Invalid filename",
			path: "\\",
		},
	}

	cmd := NewCmd(db)

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			config.SetFilename(tc.path)

			if err := cmd.Execute(); err == nil {
				t.Error("Expected an error and got nil")
			}
		})
	}
}
