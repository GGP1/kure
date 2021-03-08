package importt

import (
	"os"
	"runtime"
	"testing"

	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/entry"
	"github.com/GGP1/kure/pb"
)

func TestImport(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")

	cases := []struct {
		manager  string
		path     string
		name     string
		expected *pb.Entry
	}{
		{
			manager: "Keepass",
			path:    "testdata/test_keepass",
			name:    "keepass",
			expected: &pb.Entry{
				Username: "test@keepass.com",
				Password: "keepass123",
				URL:      "https://keepass.info/",
				Notes:    "Notes",
				Expires:  "Never",
			},
		},
		{
			manager: "Keepassxc",
			path:    "testdata/test_keepassxc",
			name:    "test/keepassxc",
			expected: &pb.Entry{
				Username: "test@keepassxc.com",
				Password: "keepassxc123",
				URL:      "https://keepassxc.org",
				Notes:    "Notes",
				Expires:  "Never",
			},
		},
		{
			manager: "1password",
			path:    "testdata/test_1password.csv",
			name:    "1password",
			expected: &pb.Entry{
				Username: "test@1password.com",
				Password: "1password123",
				URL:      "https://1password.com/",
				Notes:    "Notes.\nMember number: 1234.\nRecovery Codes: The Shire",
				Expires:  "Never",
			},
		},
		{
			manager: "Lastpass",
			path:    "testdata/test_lastpass.csv",
			// Kure will join folders with the entry names
			name: "test/lastpass",
			expected: &pb.Entry{
				Username: "test@lastpass.com",
				Password: "lastpass123",
				URL:      "https://lastpass.com/",
				Notes:    "Notes",
				Expires:  "Never",
			},
		},
		{
			manager: "Bitwarden",
			path:    "testdata/test_bitwarden.csv",
			// Kure will join folders with the entry names
			name: "test/bitwarden",
			expected: &pb.Entry{
				Username: "test@bitwarden.com",
				Password: "bitwarden123",
				URL:      "https://bitwarden.com/",
				Notes:    "Notes",
				Expires:  "Never",
			},
		},
	}

	cmd := NewCmd(db)

	for _, tc := range cases {
		t.Run(tc.manager, func(t *testing.T) {
			cmd.SetArgs([]string{tc.manager})
			cmd.Flags().Set("path", tc.path)

			if err := cmd.Execute(); err != nil {
				t.Fatalf("Failed importing entries: %v", err)
			}

			got, err := entry.Get(db, tc.name)
			if err != nil {
				t.Fatalf("Failed listing entry: %v", err)
			}

			// They aren't deeply equal
			compareEntries(t, got, tc.expected)
		})
	}
}

func TestInvalidImport(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")

	cases := []struct {
		desc    string
		manager string
		path    string
	}{
		{
			desc:    "Invalid manager",
			manager: "",
			path:    "testdata/test_keepassx.csv",
		},
		{
			desc:    "Invalid path",
			manager: "keepass",
			path:    "",
		},
		{
			desc:    "Empty file",
			manager: "bitwarden",
			path:    "testdata/test_empty.csv",
		},
		{
			desc:    "Non-existent file",
			manager: "1password",
			path:    "test.csv",
		},
		{
			desc:    "Unsupported manager",
			manager: "unsupported",
			path:    "testdata/test_lastpass.csv",
		},
		{
			desc:    "Invalid format",
			manager: "lastpass",
			path:    "testdata/test_invalid_format.json",
		},
		{
			desc:    "Invalid entry",
			manager: "1password",
			path:    "testdata/test_invalid_entry.csv",
		},
	}

	cmd := NewCmd(db)

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			cmd.Flags().Set("path", tc.path)
			cmd.SetArgs([]string{tc.manager})

			if err := cmd.Execute(); err == nil {
				t.Error("Expected an error but got nil")
			}
		})
	}
}

func TestImportAndErase(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")

	tempFile, err := os.CreateTemp("", "*.csv")
	if err != nil {
		t.Errorf("Failed creating temporary file: %v", err)
	}
	tempFile.WriteString("test")
	tempFile.Close()

	cmd := NewCmd(db)
	cmd.SetArgs([]string{"keepass"})
	f := cmd.Flags()
	f.Set("path", tempFile.Name())
	f.Set("erase", "true")

	if err := cmd.Execute(); err != nil {
		t.Errorf("Failed importing entries: %v", err)
	}

	if _, err := os.Stat(tempFile.Name()); err == nil {
		t.Error("The file wasn't erased correctly")
	}
}

func TestImportAndEraseError(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.SkipNow()
	}
	db := cmdutil.SetContext(t, "../../db/testdata/database")

	tempFile, err := os.CreateTemp("", "*.csv")
	if err != nil {
		t.Errorf("Failed creating temporary file: %v", err)
	}
	tempFile.WriteString("test")

	cmd := NewCmd(db)
	cmd.SetArgs([]string{"lastpass"})
	f := cmd.Flags()
	f.Set("path", tempFile.Name())
	f.Set("erase", "true")

	// The file attempting to erase is currently being used by another process
	if err := cmd.Execute(); err == nil {
		t.Error("Expected an error and got nil")
	}
}

func TestCreateTOTP(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")

	cases := []struct {
		desc string
		name string
		raw  string
	}{
		{
			desc: "Create",
			name: "test",
			raw:  "afrtgq",
		},
		{
			desc: "Nothing",
			raw:  "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			if err := createTOTP(db, tc.name, tc.raw); err != nil {
				t.Errorf("Failed creating TOTP: %v", err)
			}
		})
	}
}

func TestArgs(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	cmd := NewCmd(db)

	t.Run("Supported", func(t *testing.T) {
		managers := []string{"1password", "bitwarden", "keepass", "keepassx", "keepassxc", "lastpass"}

		for _, m := range managers {
			t.Run(m, func(t *testing.T) {
				if err := cmd.Args(cmd, []string{m}); err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			})
		}
	})

	t.Run("Unsupported", func(t *testing.T) {
		invalids := []string{"", "unsupported"}
		for _, inv := range invalids {
			if err := cmd.Args(cmd, []string{inv}); err == nil {
				t.Error("Expected an error and got nil")
			}
		}
	})
}

func TestPostRun(t *testing.T) {
	NewCmd(nil).PostRun(nil, nil)
}

func compareEntries(t *testing.T, got, expected *pb.Entry) {
	if got.Username != expected.Username {
		t.Errorf("Expected %s, got %s", expected.Username, got.Username)
	}

	if got.Password != expected.Password {
		t.Errorf("Expected %s, got %s", expected.Password, got.Password)
	}

	if got.URL != expected.URL {
		t.Errorf("Expected %s, got %s", expected.URL, got.URL)
	}

	if got.Notes != expected.Notes {
		t.Errorf("Expected %s, got %s", expected.Notes, got.Notes)
	}

	if got.Expires != expected.Expires {
		t.Errorf("Expected %s, got %s", expected.Expires, got.Expires)
	}
}
