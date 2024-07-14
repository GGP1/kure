package importt

import (
	"os"
	"runtime"
	"testing"

	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/entry"
	"github.com/GGP1/kure/pb"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestImport(t *testing.T) {
	db := cmdutil.SetContext(t)

	cases := []struct {
		expected *pb.Entry
		manager  string
		path     string
	}{
		{
			manager: "Keepass",
			path:    "testdata/test_keepass",
			expected: &pb.Entry{
				Name:     "keepass",
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
			expected: &pb.Entry{
				Name:     "test/keepassxc",
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
			expected: &pb.Entry{
				Name:     "1password",
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
			expected: &pb.Entry{
				Name:     "test/lastpass",
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
			expected: &pb.Entry{
				Name:     "test/bitwarden",
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

			err := cmd.Execute()
			assert.NoError(t, err, "Failed importing entries")

			got, err := entry.Get(db, tc.expected.Name)
			assert.NoError(t, err, "Failed listing entry")

			equal := proto.Equal(tc.expected, got)
			assert.True(t, equal)
		})
	}
}

func TestInvalidImport(t *testing.T) {
	db := cmdutil.SetContext(t)

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

			err := cmd.Execute()
			assert.Error(t, err)
		})
	}
}

func TestImportAndErase(t *testing.T) {
	db := cmdutil.SetContext(t)

	tempFile, err := os.CreateTemp("", "*.csv")
	assert.NoError(t, err, "Failed creating temporary file")
	tempFile.WriteString("test")
	tempFile.Close()

	cmd := NewCmd(db)
	cmd.SetArgs([]string{"keepass"})
	f := cmd.Flags()
	f.Set("path", tempFile.Name())
	f.Set("erase", "true")

	err = cmd.Execute()
	assert.NoError(t, err, "Failed importing entries")

	_, err = os.Stat(tempFile.Name())
	assert.Error(t, err, "The file wasn't erased correctly")
}

func TestImportAndEraseError(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.SkipNow()
	}
	db := cmdutil.SetContext(t)

	tempFile, err := os.CreateTemp("", "*.csv")
	assert.NoError(t, err, "Failed creating temporary file")
	tempFile.WriteString("test")

	cmd := NewCmd(db)
	cmd.SetArgs([]string{"lastpass"})
	f := cmd.Flags()
	f.Set("path", tempFile.Name())
	f.Set("erase", "true")

	// The file attempting to erase is currently being used by another process
	err = cmd.Execute()
	assert.Error(t, err)
}

func TestCreateTOTP(t *testing.T) {
	db := cmdutil.SetContext(t)

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
			err := createTOTP(db, tc.name, tc.raw)
			assert.NoError(t, err, "Failed creating TOTP")
		})
	}
}

func TestArgs(t *testing.T) {
	db := cmdutil.SetContext(t)
	cmd := NewCmd(db)

	t.Run("Supported", func(t *testing.T) {
		list := []string{"1password", "bitwarden", "keepass", "keepassx", "keepassxc", "lastpass"}
		for _, name := range list {
			err := cmd.Args(cmd, []string{name})
			assert.NoError(t, err)
		}
	})

	t.Run("Unsupported", func(t *testing.T) {
		list := []string{"", "unsupported"}
		for _, name := range list {
			err := cmd.Args(cmd, []string{name})
			assert.Error(t, err)
		}
	})
}

func TestPostRun(t *testing.T) {
	NewCmd(nil).PostRun(nil, nil)
}
