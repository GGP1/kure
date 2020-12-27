package export

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"testing"

	cmdutil "github.com/GGP1/kure/cmd"
	"github.com/GGP1/kure/db/entry"
	"github.com/GGP1/kure/pb"
)

func TestExport(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	defer db.Close()

	lockedBuf, e := pb.SecureEntry()
	e.Name = "May the force be with you"
	e.Expires = "Never"

	if err := entry.Create(db, lockedBuf, e); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		manager  string
		path     string
		expected [][]string
	}{
		{
			manager: "keepass",
			path:    "test_keepass",
			expected: [][]string{
				{"Account", "Login Name", "Password", "Web Site", "Comments"},
				{"May the force be with you", "", "", "", ""},
			},
		},
		{
			manager: "1password",
			path:    "test_1password.csv",
			expected: [][]string{
				{"Title", "Website", "Username", "Password", "Notes", "Member Number", "Recovery Codes"},
				{"May the force be with you", "", "", "", "", "", "", ""},
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
				{"", "", "Login", "May the force be with you", "", "", "", "", "", "", ""},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.manager, func(t *testing.T) {
			cmd := NewCmd(db)
			cmd.Flags().Set("path", tc.path)

			args := []string{tc.manager}

			// The file created will contain the entry added in SetContext()
			if err := cmd.RunE(cmd, args); err != nil {
				t.Fatalf("Failed exporting entries: %v", err)
			}

			if filepath.Ext(tc.path) == "" {
				tc.path += ".csv"
			}

			f, err := os.Open(tc.path)
			if err != nil {
				t.Errorf("Failed opening csv file: %v", err)
			}
			defer f.Close()

			r := csv.NewReader(f)
			got, err := r.ReadAll()
			if err != nil {
				t.Errorf("Failed reading records from the csv file: %v", err)
			}

			for i := range got {
				for j := range got {
					if got[i][j] != tc.expected[i][j] {
						t.Errorf("Expected %s, got %s", tc.expected[i][j], got[i][j])
					}
				}
			}

			// Cleanup
			f.Close()
			if err := os.Remove(tc.path); err != nil {
				t.Errorf("Failed removing csv file: %v", err)
			}
		})
	}
}

func TestInvalidExport(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	defer db.Close()

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
			cmd.Flags().Set("path", tc.path)

			args := []string{tc.manager}

			if err := cmd.RunE(cmd, args); err == nil {
				t.Fatalf("Expected test to fail but got nil")
			}
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
			expectedName:   "test",
			expectedFolder: "",
		},
		{
			desc:           "With folder",
			name:           "folder/test",
			expectedName:   "test",
			expectedFolder: "folder",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			gotName, gotFolder := splitName(tc.name)

			if gotName != tc.expectedName {
				t.Errorf("Expected %q, got %q", tc.expectedName, gotName)
			}

			if gotFolder != tc.expectedFolder {
				t.Errorf("Expected %q, got %q", tc.expectedFolder, gotFolder)
			}
		})
	}
}

func TestPostRun(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	defer db.Close()

	cmd := NewCmd(db)
	cmd.PostRun(nil, nil)
}
