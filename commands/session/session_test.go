package session

import (
	"bytes"
	"testing"
	"time"

	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/config"
	"github.com/spf13/cobra"
)

func TestExecute(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	scripts := map[string]string{
		"login": "copy -u $1 && copy $1",
	}
	config.Set("session.scripts", scripts)

	cases := []struct {
		desc string
		args []string
	}{
		{
			desc: "Kure command",
			args: []string{"kure", "stats"},
		},
		{
			desc: "Session command",
			args: []string{"pwd"},
		},
		{
			desc: "No command",
			args: []string{},
		},
	}

	cmd := NewCmd(db, &bytes.Buffer{})
	cmd.RunE = nil
	root := cmd.Root()

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			if err := execute(root, tc.args, time.Now(), &sessionOptions{}); err != nil {
				t.Errorf("Failed executing command: %v", err)
			}
		})
	}
}

func TestCleanup(t *testing.T) {
	cmd := &cobra.Command{
		Use: "test",
	}
	cmd.Flags().Bool("testing", false, "")

	cmd.Flag("testing").Changed = true
	cleanup(cmd)

	changed := cmd.Flag("testing").Changed
	if changed {
		t.Errorf("Expected false and got %t", changed)
	}
}

func TestFillScript(t *testing.T) {
	cases := []struct {
		desc     string
		args     []string
		script   string
		expected string
	}{
		{
			desc:     "No arguments",
			script:   "edit test && rm test",
			args:     []string{"no_args"},
			expected: "edit test && rm test",
		},
		{
			desc:     "One argument",
			script:   "2fa $1 && ls -q $1",
			args:     []string{"test"},
			expected: "2fa test && ls -q test",
		},
		{
			desc:     "Two arguments",
			script:   "file cat $1 && copy $2",
			args:     []string{"notes/test.txt", "testing"},
			expected: "file cat notes/test.txt && copy testing",
		},
		{
			desc:     "Enclosed by double quotes",
			script:   "card ls $1",
			args:     []string{"\"test", "double", "quotes\""},
			expected: "card ls test double quotes",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			got := fillScript(tc.args, tc.script)
			if got != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, got)
			}
		})
	}
}

func TestParseCmds(t *testing.T) {
	cases := []struct {
		desc     string
		args     []string
		expected [][]string
	}{
		{
			desc:     "Single command",
			args:     []string{"kure", "ls"},
			expected: [][]string{{"kure", "ls"}},
		},
		{
			desc:     "Two commands",
			args:     []string{"kure", "ls", "&&", "copy", "test"},
			expected: [][]string{{"kure", "ls"}, {"copy", "test"}},
		},
		{
			desc:     "Three commands",
			args:     []string{"stats", "&&", "gen", "-l", "15", "&&", "config", "argon2"},
			expected: [][]string{{"stats"}, {"gen", "-l", "15"}, {"config", "argon2"}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			got := parseCmds(tc.args)

			for i := 0; i < len(got); i++ {
				for j := 0; j < len(got[i]); j++ {
					if got[i][j] != tc.expected[i][j] {
						t.Errorf("Expected %v, got %v", tc.expected[i][j], got[i][j])
					}
				}
			}
		})
	}
}
