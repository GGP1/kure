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
		desc    string
		name    string
		length  string
		format  string
		include string
		exclude string
		repeat  string
		pass    bool
	}{
		{
			desc:    "Add",
			name:    "test",
			length:  "18",
			format:  "1,2,3,4,5",
			include: "17",
			exclude: "aA",
			repeat:  "true",
			pass:    true,
		},
		{
			desc:   "Default values",
			name:   "test2",
			length: "8",
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
			desc:    "Include and exclude error",
			name:    "test",
			length:  "5",
			include: "a",
			exclude: "a",
			pass:    false,
		},
	}

	// Defaults
	viper.Set("entry.format", []int{1, 2, 3, 4})
	viper.Set("entry.repeat", true)

	cmd := NewCmd(db, os.Stdin)
	f := cmd.Flags()

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			args := []string{tc.name}
			f.Set("length", tc.length)
			f.Set("format", tc.format)
			f.Set("include", tc.include)
			f.Set("exclude", tc.exclude)
			f.Set("repeat", tc.repeat)

			err := cmd.RunE(cmd, args)
			assertError(t, "add", err, tc.pass)

			cmd.ResetFlags()
		})
	}

	// Test separated to avoid "Add" executing after it and fail
	t.Run("Already exists", func(t *testing.T) {
		args := []string{"test"}
		f.Set("length", "10")
		if err := cmd.RunE(cmd, args); err == nil {
			t.Error("Expected an error and got nil")
		}
	})
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
	}

	cmd := phraseSubCmd(db, os.Stdin)
	f := cmd.Flags()

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			args := []string{tc.name}
			f.Set("separator", tc.separator)
			f.Set("length", tc.length)
			f.Set("list", tc.list)

			err := cmd.RunE(cmd, args)
			assertError(t, "add with passphrase", err, tc.pass)

			cmd.ResetFlags()
		})
	}

	// Test separated to avoid NoList executing after it and fail
	t.Run("Already exists", func(t *testing.T) {
		args := []string{"test"}
		f.Set("length", "5")
		if err := cmd.RunE(cmd, args); err == nil {
			t.Error("Expected an error and got nil")
		}
	})
}

func TestAddWithCustomPassword(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	defer db.Close()

	cmd := NewCmd(db, os.Stdin)
	f := cmd.Flags()
	f.Set("custom", "true")
	args := []string{"test"}

	if err := cmd.RunE(cmd, args); err != nil {
		t.Errorf("Failed running add with a custom password: %v", err)
	}

	cmd.ResetFlags()
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
		Expires:  "Never",
	}

	buf := bytes.NewBufferString("username\npassword\nurl\nnotes\n!end\nnever")

	got, err := input(db, "test", true, buf)
	if err != nil {
		t.Fatalf("Failed creating entry: %v", err)
	}

	if !reflect.DeepEqual(expected, got) {
		t.Errorf("Expected %v, got %v", expected, got)
	}
}

func TestInvalidExpirationFormat(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	defer db.Close()

	buf := bytes.NewBufferString("username\npassword\nurl\nnotes\n!end\n201/02/003")

	_, err := input(db, "test", true, buf)
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
