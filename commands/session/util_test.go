package session

import (
	"bytes"
	"io"
	"strconv"
	"testing"

	"github.com/chzyer/readline"
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

func TestConcatenatedScripts(t *testing.T) {
	var buf bytes.Buffer
	buf.WriteString("show test && login testing && clear -H")

	scripts := map[string]string{
		"show":  "ls -s $1",
		"login": "copy -u $1 && copy $1 && 2fa -c $1",
	}
	timeout := &timeout{duration: 0}
	expected := [][]string{
		{"ls", "-s", "test"},
		{"copy", "-u", "testing"},
		{"copy", "testing"},
		{"2fa", "-c", "testing"},
		{"clear", "-H"},
	}

	rl, err := readline.NewEx(&readline.Config{
		Prompt: "",
		Stdin:  io.NopCloser(&buf),
	})

	got, err := scanInput(rl, timeout, scripts)
	assert.NoError(t, err)

	assert.Equal(t, expected, got)
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
	buf.WriteString("tom && \"jerry\"")
	timeout := &timeout{duration: 0}
	expected := [][]string{
		{"tom"},
		{"jerry"},
	}

	rl, err := readline.NewEx(&readline.Config{
		Prompt: "",
		Stdin:  io.NopCloser(&buf),
	})

	got, err := scanInput(rl, timeout, map[string]string{})
	assert.NoError(t, err)

	assert.Equal(t, expected, got)
}

func TestRemoveEmptyItems(t *testing.T) {
	cases := []struct {
		desc     string
		args     []string
		expected []string
	}{
		{
			desc:     "Remove leading empty items",
			args:     []string{"", "", " ", " ", "ls"},
			expected: []string{"ls"},
		},
		{
			desc:     "Remove all empty items",
			args:     []string{"", "kure", " ", "copy", "", "tom", "", "", "-t", "6s", ""},
			expected: []string{"kure", "copy", "tom", "-t", "6s"},
		},
		{
			desc:     "Remove all empty items in a script",
			args:     []string{"timeout", "", " ", "&&", "clear", "", " "},
			expected: []string{"timeout", "&&", "clear"},
		},
		{
			desc:     "Do not remove arguments surrounded by spaces",
			args:     []string{"kure", " file ", " ", "ls"},
			expected: []string{"kure", " file ", "ls"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			actual := removeEmptyItems(tc.args)

			assert.Equal(t, tc.expected, actual)
		})
	}
}
