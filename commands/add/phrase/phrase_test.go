package phrase

import (
	"bytes"
	"testing"

	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/entry"
	"github.com/GGP1/kure/pb"

	"github.com/stretchr/testify/assert"
)

func TestPhrase(t *testing.T) {
	db := cmdutil.SetContext(t, "../../../db/testdata/database")

	cases := []struct {
		desc      string
		name      string
		separator string
		length    string
		list      string
	}{
		{
			desc:      "No list",
			name:      "test",
			separator: "/",
			length:    "7",
			list:      "nolist",
		},
		{
			desc:      "Word list",
			name:      "test wordlist",
			separator: "&",
			length:    "5",
			list:      "wordlist",
		},
		{
			desc:   "Syllable list",
			name:   "test syllablelist",
			length: "9",
			list:   "syllablelist",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			buf := bytes.NewBufferString("username\nurl\n03/05/2024\nnotes<")
			cmd := NewCmd(db, buf)
			cmd.SetArgs([]string{tc.name})

			f := cmd.Flags()
			f.Set("separator", tc.separator)
			f.Set("length", tc.length)
			f.Set("list", tc.list)

			err := cmd.Execute()
			assert.NoError(t, err)
		})
	}
}

func TestPhraseErrors(t *testing.T) {
	db := cmdutil.SetContext(t, "../../../db/testdata/database")

	err := entry.Create(db, &pb.Entry{Name: "test"})
	assert.NoError(t, err)

	cases := []struct {
		desc    string
		name    string
		length  string
		include string
		exclude string
		list    string
	}{
		{
			desc:   "Invalid name",
			name:   "",
			length: "4",
			list:   "word list",
		},
		{
			desc:   "Invalid length",
			name:   "test-invalid-length",
			length: "0",
			list:   "nolist",
		},
		{
			desc:   "Invalid list",
			name:   "test-invalid-list",
			length: "3",
			list:   "invalid",
		},
		{
			desc:   "Already exists",
			name:   "test",
			length: "3",
		},
		{
			desc:    "Included word is also excluded",
			name:    "test words",
			length:  "4",
			include: "test",
			exclude: "test",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			buf := bytes.NewBufferString("username\nurl\n03/05/2024\nnotes<")
			cmd := NewCmd(db, buf)
			cmd.SetArgs([]string{tc.name})

			f := cmd.Flags()
			f.Set("length", tc.length)
			f.Set("include", tc.include)
			f.Set("exclude", tc.exclude)
			f.Set("list", tc.list)

			err := cmd.Execute()
			assert.Error(t, err)
		})
	}
}

func TestEntryInput(t *testing.T) {
	expected := &pb.Entry{
		Name:     "test",
		Username: "username",
		URL:      "url",
		Notes:    "notes",
		Expires:  "Fri, 03 May 2024 00:00:00 +0000",
	}

	buf := bytes.NewBufferString("username\nurl\n03/05/2024\nnotes<")
	got, err := entryInput(buf, "test")
	assert.NoError(t, err, "Failed creating entry")

	// As it's randomly generated, use the same one
	expected.Password = got.Password
	assert.Equal(t, expected, got)
}

func TestInvalidExpirationTime(t *testing.T) {
	buf := bytes.NewBufferString("username\nurl\nnotes\ninvalid<\n")
	_, err := entryInput(buf, "test")
	assert.Error(t, err)
}

func TestPostRun(t *testing.T) {
	NewCmd(nil, nil).PostRun(nil, nil)
}
