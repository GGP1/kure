package session

import (
	"bufio"
	"bytes"
	"os"
	"testing"
	"time"

	cmdutil "github.com/GGP1/kure/cmd"
)

func TestSession(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	defer db.Close()

	cmd := NewCmd(db, os.Stdin, nil)
	cmd.Flags().Set("timeout", "10ms")

	t.Run("Session", func(t *testing.T) {
		if err := cmd.RunE(cmd, nil); err != nil {
			t.Fatalf("Failed running the session command: %v", err)
		}
	})
}

func TestStartSession(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	defer db.Close()

	cases := []struct {
		desc    string
		command string
	}{
		{
			desc:    "Exit command",
			command: "exit",
		},
		{
			desc:    "Kure command",
			command: "kure ls\nexit",
		},
		{
			desc:    "Show timeout",
			command: "timeout\nexit",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			buf := bytes.NewBufferString(tc.command)
			scanner := bufio.NewScanner(buf)
			start := time.Now()
			zero := time.Duration(0)
			interrupt := make(chan os.Signal, 1)

			startSession(scanner, start, zero, interrupt)
		})
	}
}
