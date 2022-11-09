package session

import (
	"bufio"
	"bytes"
	"strconv"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestCleanup(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Bool("testing", false, "")
	cmd.Flag("testing").Changed = true
	cleanup(cmd)

	changed := cmd.Flag("testing").Changed
	assert.False(t, changed)
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
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestParseCommands(t *testing.T) {
	cases := []struct {
		desc     string
		args     []string
		expected [][]string
	}{
		{
			desc:     "One command",
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
			got := parseCommands(tc.args)
			assert.Equal(t, tc.expected, got)
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
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestScanInput(t *testing.T) {
	var buf bytes.Buffer
	reader := bufio.NewReader(&buf)
	buf.WriteString("tom && \"jerry\"")
	timeout := &timeout{
		duration: 0,
	}
	expected := [][]string{
		{"tom"},
		{"jerry"},
	}

	got, err := scanInput(reader, timeout, map[string]string{})
	assert.NoError(t, err)

	assert.Equal(t, expected, got)
}
