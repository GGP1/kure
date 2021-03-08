package add

import (
	"bytes"
	"fmt"
	"testing"

	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/file"
	"github.com/GGP1/kure/pb"
)

func TestAdd(t *testing.T) {
	db := cmdutil.SetContext(t, "../../../db/testdata/database")

	cases := []struct {
		desc      string
		name      string
		ignore    string
		path      string
		semaphore string
	}{
		{
			desc:      "Add a file",
			name:      "test",
			path:      "../testdata/test_file.txt",
			semaphore: "1",
		},
		{
			desc:      "Add all files synchronously",
			name:      "testAll",
			path:      "../testdata/",
			semaphore: "1",
		},
		{
			desc:      "Add all files using goroutines",
			name:      "testAll2",
			path:      "../testdata/",
			semaphore: "8",
		},
		{
			desc:      "Add all files ignoring subfolders",
			name:      "testAll3",
			ignore:    "true",
			path:      "../testdata/",
			semaphore: "3",
		},
	}

	cmd := NewCmd(db, nil)

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			cmd.SetArgs([]string{tc.name})
			f := cmd.Flags()
			f.Set("ignore", tc.ignore)
			f.Set("path", tc.path)
			f.Set("semaphore", tc.semaphore)

			var stderr bytes.Buffer
			cmd.SetErr(&stderr)

			if err := cmd.Execute(); err != nil {
				t.Error(err)
			}

			if stderr.Len() > 0 {
				t.Errorf("Expected nothing on stderr, got %q", stderr.String())
			}
		})
	}
}

func TestAddErrors(t *testing.T) {
	db := cmdutil.SetContext(t, "../../../db/testdata/database")

	if err := file.Create(db, &pb.File{Name: "already exists.txt"}); err != nil {
		t.Fatalf("Failed creating the file: %v", err)
	}

	cases := []struct {
		desc      string
		name      string
		path      string
		semaphore string
	}{
		{
			desc: "Invalid name",
			name: "",
		},
		{
			desc:      "Invalid semaphore",
			name:      "test",
			path:      "../testdata/test_file.txt",
			semaphore: "0",
		},
		{
			desc:      "Already exists",
			name:      "already exists.txt",
			semaphore: "1",
			path:      "../testdata/test_file.txt",
		},
		{
			desc:      "Non-existent",
			name:      "non-existent",
			semaphore: "1",
			path:      "../testdata/non-existent.txt",
		},
		{
			desc:      "Empty directory",
			name:      "empty-dir",
			semaphore: "1",
			path:      "../testdata/empty",
		},
	}

	cmd := NewCmd(db, nil)

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			cmd.SetArgs([]string{tc.name})
			f := cmd.Flags()
			f.Set("path", tc.path)
			f.Set("semaphore", tc.semaphore)

			if err := cmd.Execute(); err == nil {
				t.Error("Expected an error and got nil")
			}
		})
	}
}

func TestAddNote(t *testing.T) {
	db := cmdutil.SetContext(t, "../../../db/testdata/database")

	expectedContent := []byte("note content")
	buf := bytes.NewBufferString("note content<\n")

	name := "test-notes"
	if err := addNote(db, buf, name); err != nil {
		t.Fatalf("Failed creating note: %v", err)
	}

	file, err := file.Get(db, fmt.Sprintf("notes/%s.txt", name))
	if err != nil {
		t.Fatalf("The note wasn't created: %v", err)
	}

	if !bytes.Equal(file.Content, expectedContent) {
		t.Errorf("Expected %q, got %q", string(expectedContent), string(file.Content))
	}
}

func BenchmarkAdd(b *testing.B) {
	db := cmdutil.SetContext(b, "../../../db/testdata/database")

	cmd := NewCmd(db, nil)
	f := cmd.Flags()
	f.Set("path", "../testdata/test_file.txt")

	for i := 0; i < b.N; i++ {
		cmd.SetArgs([]string{fmt.Sprintf("%d", i)})

		if err := cmd.Execute(); err != nil {
			b.Error(err)
		}
	}
}

func TestPostRun(t *testing.T) {
	NewCmd(nil, nil).PostRun(nil, nil)
}
