package session

import (
	"bytes"
	"strconv"
	"testing"

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
			desc: "Kure help",
			args: []string{"kure"},
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
			if err := execute(root, tc.args, &timeout{}); err != nil {
				t.Errorf("Failed executing command: %v", err)
			}
		})
	}
}

func TestCleanup(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
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
		script   string
		expected string
		args     []string
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

			for i := 0; i < len(tc.expected); i++ {
				for j := 0; j < len(tc.expected[i]); j++ {
					if got[i][j] != tc.expected[i][j] {
						t.Errorf("Expected %#v, got %#v", tc.expected[i], got[i])
					}
				}
			}
		})
	}
}

func TestParseDoubleQuotes(t *testing.T) {
	cases := []struct {
		args     []string
		expected []string
	}{
		{
			args:     []string{"file", "touch", "\"file", "with", "spaces\""},
			expected: []string{"file", "touch", "file with spaces"},
		},
		{
			args:     []string{"rm", "one", "\"two", "three"},
			expected: []string{"rm", "one", "\"two", "three"},
		},
		{
			args:     []string{"\"test\""},
			expected: []string{"test"},
		},
		{
			args:     []string{"\"file\"", "\"the", "wind\"", "\"is", "actually", "\"rocking\""},
			expected: []string{"file", "the wind", "is actually \"rocking"},
		},
		{
			args:     []string{"\"a\"", "\"little bit", "\"more\"", "testing\""},
			expected: []string{"a", "little bit \"more", "testing\""},
		},
	}

	for i, tc := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got := parseDoubleQuotes(tc.args)
			for i := 0; i < len(tc.expected); i++ {
				if got[i] != tc.expected[i] {
					t.Errorf("Expected %#v, got %#v", tc.expected, got)
				}
			}
		})
	}
}
