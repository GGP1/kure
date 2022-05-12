package edit

import (
	"bytes"
	"os"
	"testing"
	"time"

	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/file"
	"github.com/GGP1/kure/pb"
)

func TestEditErrors(t *testing.T) {
	db := cmdutil.SetContext(t, "../../../db/testdata/database")

	if err := file.Create(db, &pb.File{Name: "test"}); err != nil {
		t.Fatalf("Failed creating file: %v", err)
	}

	cases := []struct {
		desc   string
		name   string
		editor string
		create bool
	}{
		{
			desc: "Invalid name",
			name: "",
		},
		{
			desc:   "Non installed text editor",
			name:   "test",
			editor: "non-installed-text-editor",
		},
		{
			desc:   "Non-existent entry",
			name:   "non-existent",
			editor: "nano",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			cmd := NewCmd(db)
			cmd.SetArgs([]string{tc.name})
			cmd.Flags().Set("editor", tc.editor)

			if err := cmd.Execute(); err == nil {
				t.Error("Expected an error and got nil")
			}
		})
	}
}

func TestCreateTempFile(t *testing.T) {
	expected := []byte("content")

	filename, err := createTempFile(".txt", expected)
	if err != nil {
		t.Fatalf("Failed creating the file: %v", err)
	}
	defer os.Remove(filename)

	got, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed reading temporary file: %v", err)
	}

	if !bytes.Equal(got, expected) {
		t.Errorf("Expected %q, got %s", expected, got)
	}
}

func TestWatchFile(t *testing.T) {
	f, err := os.CreateTemp("", "*")
	if err != nil {
		t.Error(err)
	}
	defer f.Close()

	go func(f *os.File) {
		// Sleep to wait for the file to be watched
		time.Sleep(50 * time.Millisecond)
		if _, err := f.Write([]byte("anything")); err != nil {
			t.Error(err)
		}
	}(f)

	if err := watchFile(f.Name()); err != nil {
		t.Error(err)
	}
}

func TestUpdate(t *testing.T) {
	db := cmdutil.SetContext(t, "../../../db/testdata/database")

	name := "test_read_and_update.txt"
	f := &pb.File{
		Name:    name,
		Content: []byte("test"),
	}

	if err := update(db, f, "../testdata/test_read&update.txt"); err != nil {
		t.Errorf("updating record: %v", err)
	}

	got, err := file.Get(db, name)
	if err != nil {
		t.Fatalf("The file wasn't created: %v", err)
	}

	if !bytes.Equal([]byte("test"), got.Content) {
		t.Error("Failed editing file, corrupted content")
	}
}

func TestPostRun(t *testing.T) {
	NewCmd(nil).PostRun(nil, nil)
}
