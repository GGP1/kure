package add

import (
	"bytes"
	"os"
	"reflect"
	"testing"

	cmdutil "github.com/GGP1/kure/cmd"
	"github.com/GGP1/kure/pb"

	"github.com/spf13/viper"
)

func TestAddWithPassword(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	defer db.Close()

	cases := []struct {
		desc   string
		name   string
		length string
		format string
		pass   bool
	}{
		{
			desc:   "Add",
			name:   "test",
			length: "18",
			format: "1,2,3,4,5",
			pass:   true,
		},
		{
			desc: "Invalid name",
			name: "",
			pass: false,
		},
		{
			desc:   "Invalid length",
			name:   "test",
			length: "0",
			pass:   false,
		},
		{
			desc:   "Already exists",
			name:   "test",
			length: "10",
			pass:   false,
		},
	}

	cmd := NewCmd(db, os.Stdin)

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			args := []string{tc.name}

			f := cmd.Flags()
			f.Set("length", tc.length)
			f.Set("format", tc.format)

			err := cmd.RunE(cmd, args)
			assertError(t, "add", err, tc.pass)
		})
	}
}

func TestAddWithPassphrase(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	defer db.Close()

	cases := []struct {
		desc      string
		name      string
		separator string
		length    string
		list      string
		pass      bool
	}{
		{
			desc:      "No list",
			name:      "test",
			separator: "/",
			length:    "7",
			list:      "nolist",
			pass:      true,
		},
		{
			desc:      "Word list",
			name:      "test wordlist",
			separator: "&",
			length:    "5",
			list:      "wordlist",
			pass:      true,
		},
		{
			desc:   "Syllable list",
			name:   "test syllablelist",
			length: "9",
			list:   "syllablelist",
			pass:   true,
		},
		{
			desc:   "Invalid name",
			name:   "",
			length: "4",
			list:   "word list",
			pass:   false,
		},
		{
			desc:   "Invalid length",
			name:   "test",
			length: "0",
			list:   "nolist",
			pass:   false,
		},
		{
			desc:   "Already exists",
			name:   "test",
			length: "3",
			pass:   false,
		},
	}

	cmd := phraseSubCmd(db, os.Stdin)

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			args := []string{tc.name}

			f := cmd.Flags()
			f.Set("separator", tc.separator)
			f.Set("length", tc.length)
			f.Set("list", tc.list)

			err := cmd.RunE(cmd, args)
			assertError(t, "add with passphrase", err, tc.pass)
		})
	}
}

func TestAddWithCustomPassword(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	defer db.Close()

	cmd := NewCmd(db, os.Stdin)
	args := []string{"test"}
	f := cmd.Flags()
	f.Set("custom", "true")

	if err := cmd.RunE(cmd, args); err != nil {
		t.Errorf("Failed running add with a custom password: %v", err)
	}
}

func TestInput(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	defer db.Close()

	expected := &pb.Entry{
		Name:     "test",
		Username: "username",
		Password: "password",
		URL:      "url",
		Notes:    "notes",
		Expires:  "Fri, 03 May 2024 00:00:00 +0000",
	}

	buf := bytes.NewBufferString("username\npassword\nurl\nnotes!end\n03/05/2024")

	_, got, err := input(db, buf, "test", true)
	if err != nil {
		t.Fatalf("Failed creating entry: %v", err)
	}

	if !reflect.DeepEqual(expected, got) {
		t.Errorf("Expected %v, got %v", expected, got)
	}
}

func TestGenPassword(t *testing.T) {
	cases := []struct {
		desc     string
		length   uint64
		format   []int
		include  string
		exclude  string
		repeat   bool
		defaults bool
		pass     bool
	}{
		{
			desc:     "Use defaults",
			length:   14,
			format:   nil,
			defaults: true,
			pass:     true,
		},
		{
			desc:   "Invalid length",
			length: 0,
			format: []int{1, 2, 3},
			pass:   false,
		},
		{
			desc:   "Invalid format",
			length: 14,
			format: nil,
			pass:   false,
		},
		{
			desc:    "Include and exclude error",
			length:  5,
			include: "a",
			exclude: "a",
			pass:    false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			viper.Set("entry.format", nil)
			if tc.defaults {
				viper.Set("entry.format", []int{1, 2, 3, 4})
			}

			_, err := genPassword(tc.length, tc.format, tc.include, tc.exclude, tc.repeat)
			assertError(t, "genPassword", err, tc.pass)
		})
	}
}

func TestInvalidExpirationFormat(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	defer db.Close()

	buf := bytes.NewBufferString("username\npassword\nurl\nnotes!end\n201/02/003")

	_, _, err := input(db, buf, "test", true)
	if err == nil {
		t.Fatal("Expected input() to fail but it didn't")
	}
}

func TestPostRun(t *testing.T) {
	password := NewCmd(nil, nil)
	f := password.PostRun
	f(password, nil)

	passphrase := phraseSubCmd(nil, nil)
	f2 := passphrase.PostRun
	f2(passphrase, nil)
}

func assertError(t *testing.T, funcName string, err error, pass bool) {
	if err != nil && pass {
		t.Errorf("Failed running %s: %v", funcName, err)
	}
	if err == nil && !pass {
		t.Error("Expected an error and got nil")
	}
}
