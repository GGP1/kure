package tfa

import (
	"testing"
	"time"

	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/config"
	"github.com/GGP1/kure/db/entry"
	"github.com/GGP1/kure/db/totp"
	"github.com/GGP1/kure/pb"

	"github.com/atotto/clipboard"
	"github.com/stretchr/testify/assert"
	bolt "go.etcd.io/bbolt"
)

func Test2FA(t *testing.T) {
	if clipboard.Unsupported {
		t.Skip("No clipboard utilities available")
	}
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	createElements(t, db)

	cases := []struct {
		desc    string
		name    string
		copy    string
		info    string
		timeout string
	}{
		{
			desc: "List all",
			name: "",
		},
		{
			desc: "List one",
			name: "test",
		},
		{
			desc: "Copy with default timeout",
			name: "test",
			copy: "true",
		},
		{
			desc:    "Copy with custom timeout",
			name:    "test",
			copy:    "true",
			timeout: "1ns",
		},
		{
			desc: "Show setup key information",
			name: "test",
			info: "true",
		},
	}

	cmd := NewCmd(db)
	config.Set("clipboard.timeout", "1ns") // Set default

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			cmd.SetArgs([]string{tc.name})
			f := cmd.Flags()
			f.Set("copy", tc.copy)
			f.Set("info", tc.info)
			f.Set("timeout", tc.timeout)

			err := cmd.Execute()
			assert.NoError(t, err, "Failed generating TOTP code")
		})
	}
}

func Test2FAErrors(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")

	cases := []struct {
		desc string
		name string
	}{
		{
			desc: "Entry does not exist",
			name: "non-existent",
		},
	}

	cmd := NewCmd(db)

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			cmd.SetArgs([]string{tc.name})

			err := cmd.Execute()
			assert.Error(t, err)
		})
	}
}

func TestGenerateTOTP(t *testing.T) {
	cases := []struct {
		desc     string
		expected string
		digits   int
	}{
		{
			desc:     "6 digits",
			digits:   6,
			expected: "419244",
		},
		{
			desc:     "7 digits",
			digits:   7,
			expected: "0419244",
		},
		{
			desc:     "8 digits",
			digits:   8,
			expected: "80419244",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			unixTime := time.Unix(10, 0)

			got := GenerateTOTP("IFGEWRKSIFJUMR2R", unixTime, tc.digits)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestPostRun(t *testing.T) {
	NewCmd(nil).PostRun(nil, nil)
}

func createElements(t *testing.T, db *bolt.DB) {
	t.Helper()
	err := entry.Create(db, &pb.Entry{Name: "test"})
	assert.NoError(t, err)

	err = totp.Create(db, &pb.TOTP{Name: "test", Raw: "AG5H1H2"})
	assert.NoError(t, err)
}
