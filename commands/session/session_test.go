package session

import (
	"bytes"
	"io"
	"testing"
	"time"

	cmdutil "github.com/GGP1/kure/commands"
)

func TestStartSession(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")

	cases := []struct {
		desc    string
		command string
		timeout time.Duration
	}{
		{
			desc:    "Kure command",
			command: "kure gen -l 15 -L 1,2,3",
		},
		{
			desc:    "Show current directory",
			command: "pwd",
		},
		{
			desc:    "Show timeout",
			command: "timeout",
			timeout: time.Duration(500),
		},
		{
			desc:    "No timeout",
			command: "timeout",
			timeout: time.Duration(0),
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			buf := bytes.NewBufferString(tc.command)
			cmd := NewCmd(db, buf)

			cmd.SetOut(io.Discard)
			start := time.Now()
			opts := &sessionOptions{
				prefix:  "",
				timeout: tc.timeout,
			}

			// Start a goroutine so it doesn't block and we can skip the test
			go startSession(cmd, db, buf, start, opts)
			t.SkipNow()
		})
	}
}
