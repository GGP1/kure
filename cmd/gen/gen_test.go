package gen

import (
	"testing"

	"github.com/spf13/viper"
)

func TestGen(t *testing.T) {
	cases := []struct {
		desc   string
		length string
		format string
		repeat string
		qr     string
	}{
		{desc: "Generate", length: "16", format: "1,2,3", repeat: "true"},
		{desc: "QR code", length: "4", format: "1,2,3,4,5,6", qr: "true"},
	}

	cmd := NewCmd()
	f := cmd.Flags()

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			f.Set("length", tc.length)
			f.Set("format", tc.format)
			f.Set("repeat", tc.repeat)
			f.Set("qr", tc.qr)

			if err := cmd.RunE(cmd, nil); err != nil {
				t.Errorf("Failed generating a password: %v", err)
			}

			cmd.ResetFlags()
		})
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
	cases := []struct {
		desc    string
		length  string
		format  string
		include string
		exclude string
		repeat  string
		qr      string
	}{
		{
			desc:   "Invalid length",
			length: "0",
		},
		{
			desc:    "Invalid include and exclude",
			length:  "10",
			include: "aA",
			exclude: "aA",
		},
		{
			desc:   "Invalid format",
			length: "22",
			format: "123",
		},
		{
			desc:   "No format nor defaults",
			length: "8",
		},
		{
			desc:   "QR code content too long",
			length: "1275",
			format: "1,2,3,4,5",
			repeat: "true",
			qr:     "true",
		},
	}

	cmd := NewCmd()
	f := cmd.Flags()

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			f.Set("length", tc.length)
			f.Set("format", tc.format)
			f.Set("repeat", tc.repeat)
			f.Set("include", tc.include)
			f.Set("exclude", tc.exclude)
			f.Set("qr", tc.qr)

			if err := cmd.RunE(cmd, nil); err == nil {
				t.Error("Expected an error and got nil")
			}

			cmd.ResetFlags()
		})
	}
}

func TestGenPhrase(t *testing.T) {
	cases := []struct {
		desc      string
		length    string
		separator string
		list      string
		qr        string
	}{
		{
			desc:      "No list",
			length:    "6",
			separator: "%",
			list:      "nolist",
		},
		{
			desc:   "Word list",
			length: "5",
			list:   "wordlist",
		},
		{
			desc:      "Syllable list",
			length:    "4",
			separator: "-",
			list:      "syllablelist",
		},
		{
			desc:   "QR code",
			length: "7",
			qr:     "true",
		},
	}

	cmd := phraseSubCmd()
	f := cmd.Flags()

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			f.Set("length", tc.length)
			f.Set("separator", tc.separator)
			f.Set("list", tc.list)
			f.Set("qr", tc.qr)

			if err := cmd.RunE(cmd, nil); err != nil {
				t.Errorf("Failed generating a passphrase: %v", err)
			}

			cmd.ResetFlags()
		})
	}
}

func TestGenPhraseErrors(t *testing.T) {
	cases := []struct {
		desc    string
		length  string
		include string
		exclude string
		list    string
		qr      string
	}{
		{
			desc:   "QR code content too long",
			length: "600",
			list:   "wordlist",
			qr:     "true",
		},
		{
			desc:   "Invalid length",
			length: "0",
			list:   "wordlist",
		},
		{
			desc:    "Invalid include and exclude",
			length:  "3",
			list:    "wordlist",
			include: "test",
			exclude: "test",
		},
	}

	cmd := phraseSubCmd()
	f := cmd.Flags()

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			f.Set("length", tc.length)
			f.Set("include", tc.include)
			f.Set("exclude", tc.exclude)
			f.Set("qr", tc.qr)

			if err := cmd.RunE(cmd, nil); err == nil {
				t.Error("Expected an error and got nil")
			}

			cmd.ResetFlags()
		})
	}
}
