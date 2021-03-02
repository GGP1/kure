package tfa

import (
	"testing"
	"time"

	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/config"
	"github.com/GGP1/kure/db/entry"
	"github.com/GGP1/kure/db/totp"
	"github.com/GGP1/kure/pb"

	bolt "go.etcd.io/bbolt"
)

func Test2FA(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	createElements(t, db)

	cases := []struct {
		desc    string
		name    string
		copy    string
		timeout string
	}{
		{
			desc: "List all",
			name: "",
		},
		{
			desc: "List one and copy with default timeout",
			name: "test",
			copy: "true",
		},
		{
			desc:    "List one and copy with custom timeout",
			name:    "test",
			copy:    "true",
			timeout: "1ns",
		},
	}

	cmd := NewCmd(db)
	config.Set("clipboard.timeout", "1ns") // Set default

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			cmd.SetArgs([]string{tc.name})
			f := cmd.Flags()
			f.Set("copy", tc.copy)
			f.Set("timeout", tc.timeout)

			if err := cmd.Execute(); err != nil {
				t.Errorf("Failed generating TOTP code: %v", err)
			}
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

			if err := cmd.Execute(); err == nil {
				t.Error("Expected an error and got nil")
			}
		})
	}
}

func TestGenerateTOTP(t *testing.T) {
	cases := []struct {
		desc     string
		digits   int
		expected string
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
			if got != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, got)
			}
		})
	}
}

func TestPostRun(t *testing.T) {
	NewCmd(nil).PostRun(nil, nil)
}

func createElements(t *testing.T, db *bolt.DB) {
	t.Helper()
	if err := entry.Create(db, &pb.Entry{Name: "test"}); err != nil {
		t.Fatal(err)
	}
	if err := totp.Create(db, &pb.TOTP{Name: "test"}); err != nil {
		t.Fatal(err)
	}
}
