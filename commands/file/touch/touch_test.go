package touch

import (
	"bytes"
	"os"
	"testing"

	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/file"
	"github.com/GGP1/kure/pb"

	bolt "go.etcd.io/bbolt"
)

func TestTouch(t *testing.T) {
	db := cmdutil.SetContext(t, "../../../db/testdata/database")
	createTestFiles(t, db)

	rootDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		desc  string
		path  string
		names []string
	}{
		{
			desc:  "Create one file",
			names: []string{"test.txt"},
			path:  "testdata/create/one",
		},
		{
			desc: "Create all files",
			path: "testdata/create",
		},
		{
			desc:  "Create multiple files",
			names: []string{"testAll/file.csv", "testAll/file.pdf"},
			path:  "testdata/create",
		},
		{
			desc:  "Create a directory",
			names: []string{"testdata/subfolder/"},
			path:  "testdata/create",
		},
	}

	cmd := NewCmd(db)

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			var stderr bytes.Buffer
			cmd.SetErr(&stderr)
			cmd.SetArgs(tc.names)
			f := cmd.Flags()
			f.Set("path", "../"+tc.path)

			if err := cmd.Execute(); err != nil {
				t.Errorf("Failed running touch: %v", err)
			}

			if stderr.Len() > 0 {
				t.Errorf("Expected nothing on stderr, got %q", stderr.String())
			}

			// Go back to the directory where the test was executed
			os.Chdir(rootDir)
		})
	}

	if err := os.RemoveAll("../testdata/create"); err != nil {
		t.Fatalf("Failed removing the folder containing created files: %v", err)
	}
}

func TestPostRun(t *testing.T) {
	NewCmd(nil).PostRun(nil, nil)
}

func createTestFiles(t *testing.T, db *bolt.DB) {
	t.Helper()

	names := []string{"test.txt", "testAll/file.csv", "testAll/file.pdf", "testdata/subfolder/file.txt"}
	for _, name := range names {
		if err := file.Create(db, &pb.File{Name: cmdutil.NormalizeName(name)}); err != nil {
			t.Errorf("Failed creating %q: %v", name, err)
		}
	}
}
