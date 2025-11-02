package session

import (
	"bytes"
	"testing"

	"github.com/GGP1/kure/config"

	"github.com/stretchr/testify/assert"
)

func TestExecute(t *testing.T) {
	scripts := map[string]string{
		"login": "copy -u $1 && copy $1",
	}
	config.Set("session.scripts", scripts)

	cases := []struct {
		desc string
		args [][]string
	}{
		{
			desc: "Root command",
			args: [][]string{{"kure", "stats"}},
		},
		{
			desc: "Session command",
			args: [][]string{{"pwd"}},
		},
		{
			desc: "Help command",
			args: [][]string{{"kure"}},
		},
		{
			desc: "No command",
			args: [][]string{},
		},
	}

	cmd := NewCmd(&bytes.Buffer{})
	cmd.RunE = nil
	root := cmd.Root()

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			err := execute(root, tc.args, &timeout{})
			assert.NoError(t, err, "Failed executing command")
		})
	}
}
