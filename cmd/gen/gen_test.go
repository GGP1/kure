package gen

import (
	"testing"

	"github.com/spf13/viper"
)

func TestGen(t *testing.T) {
	cases := map[string]struct {
		length string
		format string
		repeat string
		qr     string
	}{
		"Generate": {length: "16", format: "1,2,3", repeat: "true"},
		"QR code":  {length: "4", format: "1,2,3,4,5,6", qr: "true"},
	}

	cmd := NewCmd()
	f := cmd.Flags()
	for k, tc := range cases {
		f.Set("length", tc.length)
		f.Set("format", tc.format)
		f.Set("repeat", tc.repeat)
		f.Set("qr", tc.qr)

		if err := cmd.RunE(cmd, nil); err != nil {
			t.Errorf("%s: failed generating a password: %v", k, err)
		}

		cmd.ResetFlags()
	}
}

func TestPostRun(t *testing.T) {
	password := NewCmd()
	f := password.PostRun
	f(password, nil)

	passphrase := phraseSubCmd()
	f2 := passphrase.PostRun
	f2(passphrase, nil)
}

func TestGenDefaults(t *testing.T) {
	length := "10"
	viper.Set("entry.format", []int{1, 2, 3})
	viper.Set("entry.repeat", true)

	cmd := NewCmd()
	cmd.Flags().Set("length", length)

	if err := cmd.RunE(cmd, nil); err != nil {
		t.Errorf("Failed generating a password with default values: %v", err)
	}

	// Cleanup
	viper.Set("entry.format", nil)
	viper.Set("entry.repeat", nil)
}

func TestGenErrors(t *testing.T) {
	cases := map[string]struct {
		length  string
		format  string
		include string
		exclude string
		repeat  string
		qr      string
	}{
		"Invalid length":              {length: "0"},
		"Invalid include and exclude": {length: "10", include: "aA", exclude: "aA"},
		"Invalid format":              {length: "22", format: "123"},
		"No format nor defaults":      {length: "8"},
		"QR code content too long":    {length: "1275", format: "1,2,3,4,5", repeat: "true", qr: "true"},
	}

	cmd := NewCmd()
	f := cmd.Flags()
	for k, tc := range cases {
		f.Set("length", tc.length)
		f.Set("format", tc.format)
		f.Set("repeat", tc.repeat)
		f.Set("include", tc.include)
		f.Set("exclude", tc.exclude)
		f.Set("qr", tc.qr)

		if err := cmd.RunE(cmd, nil); err == nil {
			t.Errorf("%s: expected an error and got nil", k)
		}

		cmd.ResetFlags()
	}
}

func TestGenPhrase(t *testing.T) {
	cases := map[string]struct {
		length    string
		separator string
		list      string
		qr        string
	}{
		"No list":       {length: "6", separator: "%", list: "nolist"},
		"Word list":     {length: "5", list: "wordlist"},
		"Syllable list": {length: "4", separator: "-", list: "syllablelist"},
		"QR code":       {length: "7", qr: "true"},
	}

	cmd := phraseSubCmd()
	f := cmd.Flags()
	for k, tc := range cases {
		f.Set("length", tc.length)
		f.Set("separator", tc.separator)
		f.Set("list", tc.list)
		f.Set("qr", tc.qr)

		if err := cmd.RunE(cmd, nil); err != nil {
			t.Errorf("%s: failed generating a passphrase: %v", k, err)
		}

		cmd.ResetFlags()
	}
}

func TestGenPhraseErrors(t *testing.T) {
	cases := map[string]struct {
		length  string
		include string
		exclude string
		list    string
		qr      string
	}{
		"QR code content too long":    {length: "600", list: "wordlist", qr: "true"},
		"Invalid length":              {length: "0", list: "wordlist"},
		"Invalid include and exclude": {length: "3", list: "wordlist", include: "test", exclude: "test"},
	}

	cmd := phraseSubCmd()
	f := cmd.Flags()
	for k, tc := range cases {
		f.Set("length", tc.length)
		f.Set("include", tc.include)
		f.Set("exclude", tc.exclude)
		f.Set("qr", tc.qr)

		if err := cmd.RunE(cmd, nil); err == nil {
			t.Errorf("%s: expected an error and got nil", k)
		}

		cmd.ResetFlags()
	}
}
