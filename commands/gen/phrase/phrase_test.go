package phrase

import (
	"math"
	"os"
	"testing"

	"github.com/atotto/clipboard"
)

func TestGenPhrase(t *testing.T) {
	if clipboard.Unsupported {
		t.Skip("No clipboard utilities available")
	}
	cases := []struct {
		desc      string
		length    string
		separator string
		list      string
		copy      string
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
			copy:   "true",
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

	cmd := NewCmd()
	os.Stdout = os.NewFile(0, "") // Mute stdout
	os.Stderr = os.NewFile(0, "") // Mute stderr

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			f := cmd.Flags()
			f.Set("length", tc.length)
			f.Set("separator", tc.separator)
			f.Set("list", tc.list)
			f.Set("copy", tc.copy)
			f.Set("qr", tc.qr)

			if err := cmd.Execute(); err != nil {
				t.Errorf("Failed generating a passphrase: %v", err)
			}
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
		{
			desc:   "Invalid list",
			length: "4",
			list:   "invalid",
		},
	}

	cmd := NewCmd()
	os.Stdout = os.NewFile(0, "") // Mute stdout
	os.Stderr = os.NewFile(0, "") // Mute stderr

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			f := cmd.Flags()
			f.Set("length", tc.length)
			f.Set("include", tc.include)
			f.Set("exclude", tc.exclude)
			f.Set("list", tc.list)
			f.Set("qr", tc.qr)

			if err := cmd.Execute(); err == nil {
				t.Error("Expected an error and got nil")
			}
		})
	}
}

func TestFormatSecurity(t *testing.T) {
	cases := []struct {
		desc             string
		keyspace         float64
		timeToCrack      float64
		expectedKeyspace string
		expectedTime     string
	}{
		{
			desc:             "Inf",
			keyspace:         math.Inf(1),
			timeToCrack:      math.Inf(1),
			expectedKeyspace: "+Inf",
			expectedTime:     "+Inf",
		},
		{
			desc:             "Sextillion-Millenniums",
			keyspace:         sextillion,
			timeToCrack:      float64(millennium),
			expectedKeyspace: "1.00 sextillion",
			expectedTime:     "1.00 millenniums",
		},
		{
			desc:             "Quintillion-Century",
			keyspace:         quintillion,
			timeToCrack:      century,
			expectedKeyspace: "1.00 quintillion",
			expectedTime:     "1.00 centuries",
		},
		{
			desc:             "Quadrillion-Decades",
			keyspace:         quadrillion,
			timeToCrack:      decade,
			expectedKeyspace: "1.00 quadrillion",
			expectedTime:     "1.00 decades",
		},
		{
			desc:             "Trillion-Years",
			keyspace:         trillion,
			timeToCrack:      year,
			expectedKeyspace: "1.00 trillion",
			expectedTime:     "1.00 years",
		},
		{
			desc:             "Billion-Months",
			keyspace:         billion,
			timeToCrack:      month,
			expectedKeyspace: "1.00 billion",
			expectedTime:     "1.00 months",
		},
		{
			desc:             "Million-Days",
			keyspace:         million,
			timeToCrack:      day,
			expectedKeyspace: "1.00 million",
			expectedTime:     "1.00 days",
		},
		{
			desc:             "Thousand-Hours",
			keyspace:         thousand,
			timeToCrack:      hour,
			expectedKeyspace: "1.00 thousand",
			expectedTime:     "1.00 hours",
		},
		{
			desc:             "Minutes",
			keyspace:         15,
			timeToCrack:      minute,
			expectedKeyspace: "15.00",
			expectedTime:     "1.00 minutes",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			gotKeyspace, gotTime := FormatSecretSecurity(tc.keyspace, tc.timeToCrack)

			if gotKeyspace != tc.expectedKeyspace {
				t.Errorf("Expected %q, got %q", tc.expectedKeyspace, gotKeyspace)
			}

			if gotTime != tc.expectedTime {
				t.Errorf("Expected %q, got %q", tc.expectedTime, gotTime)
			}
		})
	}
}

func TestPostRun(t *testing.T) {
	NewCmd().PostRun(nil, nil)
}
