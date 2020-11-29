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

	cases := map[string]struct {
		name    string
		length  string
		format  string
		include string
		exclude string
		repeat  string
		pass    bool
	}{
		"Add": {
			name:    "test",
			length:  "18",
			format:  "1,2,3,4,5",
			include: "17",
			exclude: "aA",
			repeat:  "true",
			pass:    true,
		},
		"Default values": {
			name:   "test2",
			length: "8",
			pass:   true,
		},
		"Invalid name": {
			name: "",
			pass: false,
		},
		"Invalid length": {
			name:   "test",
			length: "0",
			pass:   false,
		},
		"Name too long": {
			name: "012345678901234567890123456789012345678901234567890123456789",
			pass: false,
		},
		"Include and exclude error": {
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

	for k, tc := range cases {
		args := []string{tc.name}
		f.Set("length", tc.length)
		f.Set("format", tc.format)
		f.Set("include", tc.include)
		f.Set("exclude", tc.exclude)
		f.Set("repeat", tc.repeat)

		err := cmd.RunE(cmd, args)
		assertError(t, k, "add", err, tc.pass)

		cmd.ResetFlags()
	}

	// Test already exists separate to avoid "Add" executing after it and fail
	args := []string{"test"}
	f.Set("length", "10")
	if err := cmd.RunE(cmd, args); err == nil {
		t.Error("Expected an error and got nil")
	}
}

func TestAddWithPassphrase(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	defer db.Close()

	cases := map[string]struct {
		name      string
		separator string
		length    string
		list      string
		pass      bool
	}{
		"No list": {
			name:      "test",
			separator: "/",
			length:    "7",
			list:      "nolist",
			pass:      true,
		},
		"Word list": {
			name:      "test wordlist",
			separator: "&",
			length:    "5",
			list:      "wordlist",
			pass:      true,
		},
		"Syllable list": {
			name:   "test syllablelist",
			length: "9",
			list:   "syllablelist",
			pass:   true,
		},
		"Invalid name": {
			name:   "",
			length: "4",
			list:   "word list",
			pass:   false,
		},
		"Invalid length": {
			name:   "test",
			length: "0",
			list:   "nolist",
			pass:   false,
		},
		"Name too long": {
			name: "0123456789012345678901234567890123456789012345678901234567890",
			pass: false,
		},
	}

	cmd := phraseSubCmd(db, os.Stdin)
	f := cmd.Flags()

	for k, tc := range cases {
		args := []string{tc.name}
		f.Set("separator", tc.separator)
		f.Set("length", tc.length)
		f.Set("list", tc.list)

		err := cmd.RunE(cmd, args)
		assertError(t, k, "add with passphrase", err, tc.pass)

		cmd.ResetFlags()
	}

	// Test already exists separate to avoid NoList executing after it and fail
	args := []string{"test"}
	f.Set("length", "5")
	if err := cmd.RunE(cmd, args); err == nil {
		t.Error("Expected an error and got nil")
	}
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

func TestEntryInput(t *testing.T) {
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

	got, err := entryInput(db, "test", true, buf)
	if err != nil {
		t.Fatalf("Failed creating entry: %v", err)
	}

	if !reflect.DeepEqual(expected, got) {
		t.Errorf("Expected %v, got %v", expected, got)
	}
}

func TestFormatExpiration(t *testing.T) {
	expected := "Fri, 20 Aug 2021 00:00:00 +0000"

	got, err := formatExpiration("20/08/2021")
	if err != nil {
		t.Fatal(err)
	}

	if got != expected {
		t.Errorf("Expected %s, got %s", expected, got)
	}
}

func TestInvalidTimeFormat(t *testing.T) {
	_, err := formatExpiration("204/09/2021")
	if err == nil {
		t.Error("Expected to get an error but got nil")
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

func assertError(t *testing.T, name, funcName string, err error, pass bool) {
	if err != nil && pass {
		t.Errorf("%s: failed running %s: %v", name, funcName, err)
	}
	if err == nil && !pass {
		t.Errorf("%s: expected an error and got nil", name)
	}
}
