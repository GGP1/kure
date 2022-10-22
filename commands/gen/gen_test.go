package gen

import (
	"bytes"
	"strconv"
	"testing"

	"github.com/atotto/clipboard"
	"github.com/stretchr/testify/assert"
)

func TestGen(t *testing.T) {
	if clipboard.Unsupported {
		t.Skip("No clipboard utilities available")
	}
	cases := []struct {
		desc   string
		length string
		levels string
		repeat bool
		copy   bool
		qr     bool
	}{
		{
			desc:   "Generate",
			length: "16",
			levels: "1,2,3",
			repeat: true,
			copy:   true,
		},
		{
			desc:   "QR code",
			length: "4",
			levels: "1,2,3,4,5",
			qr:     true,
		},
	}

	cmd := NewCmd()

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			f := cmd.Flags()
			f.Set("length", tc.length)
			f.Set("levels", tc.levels)
			f.Set("repeat", strconv.FormatBool(tc.repeat))
			f.Set("copy", strconv.FormatBool(tc.copy))
			f.Set("qr", strconv.FormatBool(tc.qr))
			f.Set("mute", "true")
			var buf bytes.Buffer
			cmd.SetOut(&buf)

			err := cmd.Execute()
			assert.NoError(t, err, "Failed generating a password")
		})
	}
}

func TestGenErrors(t *testing.T) {
	cases := []struct {
		desc    string
		length  string
		levels  string
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
			desc:   "Invalid levels",
			length: "22",
			levels: "123",
		},
		{
			desc:   "No levels nor defaults",
			length: "8",
		},
		{
			desc:   "QR code content too long",
			length: "1275",
			levels: "1,2,3,4,5",
			repeat: "true",
			qr:     "true",
		},
	}

	cmd := NewCmd()

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			f := cmd.Flags()
			f.Set("length", tc.length)
			f.Set("levels", tc.levels)
			f.Set("repeat", tc.repeat)
			f.Set("include", tc.include)
			f.Set("exclude", tc.exclude)
			f.Set("qr", tc.qr)

			err := cmd.Execute()
			assert.Error(t, err)
		})
	}
}

func TestPostRun(t *testing.T) {
	NewCmd().PostRun(nil, nil)
}
