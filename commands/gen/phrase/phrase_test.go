package phrase

import (
	"math"
	"os"
	"strconv"
	"testing"

	"github.com/atotto/clipboard"
	"github.com/stretchr/testify/assert"
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
		copy      bool
		qr        bool
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
			copy:   true,
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
			qr:     true,
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
			f.Set("copy", strconv.FormatBool(tc.copy))
			f.Set("qr", strconv.FormatBool(tc.qr))

			err := cmd.Execute()
			assert.NoError(t, err, "Failed generating a passphrase")
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
		qr      bool
	}{
		{
			desc:   "QR code content too long",
			length: "600",
			list:   "wordlist",
			qr:     true,
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
			f.Set("qr", strconv.FormatBool(tc.qr))

			err := cmd.Execute()
			assert.Error(t, err)
		})
	}
}

func TestFormatSecurity(t *testing.T) {
	cases := []struct {
		desc             string
		expectedKeyspace string
		expectedTime     string
		keyspace         float64
		timeToCrack      float64
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

			assert.Equal(t, tc.expectedKeyspace, gotKeyspace)
			assert.Equal(t, tc.expectedTime, gotTime)
		})
	}
}

func TestBeautify(t *testing.T) {
	cases := []struct {
		actual   string
		expected string
	}{
		{
			actual:   "27738957312183663402227335168.00",
			expected: "27,738,957,312,183,663,402,227,335,168.00",
		},
		{
			actual:   "7456198.39",
			expected: "7,456,198.39",
		},
		{
			actual:   "124561.65",
			expected: "124,561.65",
		},
		{
			actual:   "379.78",
			expected: "379.78",
		},
	}

	for i, tc := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got := prettify(tc.actual)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestPostRun(t *testing.T) {
	NewCmd().PostRun(nil, nil)
}
