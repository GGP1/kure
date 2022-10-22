package export

import (
	"bytes"
	"encoding/csv"
	"os"
	"path/filepath"
	"testing"

	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/entry"
	"github.com/GGP1/kure/db/totp"
	"github.com/GGP1/kure/pb"

	"github.com/stretchr/testify/assert"
	bolt "go.etcd.io/bbolt"
)

func TestExport(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	createEntry(t, db)

	cases := []struct {
		manager  string
		path     string
		expected [][]string
	}{
		{
			manager: "keepassx",
			path:    "test_keepassx",
			expected: [][]string{
				{"Account", "Login Name", "Password", "Web Site", "Comments"},
				{"May the force be with you", "", "", "", ""},
			},
		},
		{
			manager: "keepassxc",
			path:    "test_keepassxc",
			expected: [][]string{
				{"Group", "Title", "Username", "Password", "URL", "Notes"},
				{"", "May the force be with you", "", "", "", ""},
			},
		},
		{
			manager: "1password",
			path:    "test_1password.csv",
			expected: [][]string{
				{"Title", "Website", "Username", "Password", "Notes", "Member Number", "Recovery Codes"},
				{"May the force be with you", "", "", "", "", "", ""},
			},
		},
		{
			manager: "lastpass",
			path:    "test_lastpass.csv",
			expected: [][]string{
				{"URL", "Username", "Password", "Extra", "Name", "Grouping", "Fav"},
				{"", "", "", "", "May the force be with you", "", ""},
			},
		},
		{
			manager: "bitwarden",
			path:    "test_bitwarden.csv",
			expected: [][]string{
				{"Folder", "Favorite", "Type", "Name", "Notes", "Fields", "Login_uri", "Login_username", "Login_password", "Login_totp"},
				{"", "", "login", "May the force be with you", "", "", "", "", "", ""},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.manager, func(t *testing.T) {
			cmd := NewCmd(db)
			cmd.SetArgs([]string{tc.manager})
			cmd.Flags().Set("path", tc.path)

			err := cmd.Execute()
			assert.NoError(t, err, "Failed exporting entries")

			if filepath.Ext(tc.path) == "" {
				tc.path += ".csv"
			}

			content, err := os.ReadFile(tc.path)
			assert.NoError(t, err, "Failed opening csv file")
			defer os.Remove(tc.path)

			r := csv.NewReader(bytes.NewReader(content))
			got, err := r.ReadAll()
			assert.NoError(t, err, "Failed reading records from the csv file")

			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestInvalidExport(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")

	cases := []struct {
		desc    string
		manager string
		path    string
	}{
		{desc: "Invalid name", manager: "", path: "test.csv"},
		{desc: "Invalid path", manager: "keepass", path: ""},
		{desc: "Unsupported manager", manager: "unsupported", path: "test.csv"},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			cmd := NewCmd(db)
			cmd.SetArgs([]string{tc.manager})
			cmd.Flags().Set("path", tc.path)

			err := cmd.Execute()
			assert.Error(t, err)
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
				err := cmd.Args(cmd, []string{m})
				assert.NoError(t, err)
			})
		}
	})

	t.Run("Unsupported", func(t *testing.T) {
		invalids := []string{"", "unsupported"}
		for _, inv := range invalids {
			err := cmd.Args(cmd, []string{inv})
			assert.Error(t, err)
		}
	})
}

func TestGetTOTP(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")

	tp := &pb.TOTP{
		Name:   "test",
		Raw:    "awtapr",
		Digits: 6,
	}
	err := totp.Create(db, tp)
	assert.NoError(t, err, "Failed creating TOTP")

	cases := []struct {
		desc     string
		name     string
		expected string
	}{
		{
			desc:     "Exists",
			name:     tp.Name,
			expected: tp.Raw,
		},
		{
			desc:     "Does not exists",
			name:     "non-existent",
			expected: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			got := getTOTP(db, tc.name)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestSplitName(t *testing.T) {
	cases := []struct {
		desc           string
		name           string
		expectedName   string
		expectedFolder string
	}{
		{
			desc:           "No folder",
			name:           "test",
			expectedFolder: "",
			expectedName:   "test",
		},
		{
			desc:           "With folder",
			name:           "folder/test",
			expectedFolder: "folder",
			expectedName:   "test",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			gotFolder, gotName := splitName(tc.name)

			assert.Equal(t, tc.expectedFolder, gotFolder)
			assert.Equal(t, tc.expectedName, gotName)
		})
	}
}

func TestPostRun(t *testing.T) {
	NewCmd(nil).PostRun(nil, nil)
}

func createEntry(t *testing.T, db *bolt.DB) {
	t.Helper()
	e := &pb.Entry{
		Name:    "May the force be with you",
		Expires: "Never",
	}
	err := entry.Create(db, e)
	assert.NoError(t, err)
}
