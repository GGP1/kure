package edit

import (
	"os"
	"testing"
	"time"

	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/file"
	"github.com/GGP1/kure/pb"

	"github.com/stretchr/testify/assert"
)

func TestEditErrors(t *testing.T) {
	db := cmdutil.SetContext(t)

	err := file.Create(db, &pb.File{Name: "test"})
	assert.NoError(t, err, "Failed creating file")

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

			err := cmd.Execute()
			assert.Error(t, err)
		})
	}
}

func TestCreateTempFile(t *testing.T) {
	expected := []byte("content")

	filename, err := createTempFile(".txt", expected)
	assert.NoError(t, err, "Failed creating the file")

	got, err := os.ReadFile(filename)
	assert.NoError(t, err, "Failed reading temporary file")

	assert.Equal(t, expected, got)
}

func TestWatchFile(t *testing.T) {
	f, err := os.CreateTemp("", "*")
	assert.NoError(t, err)
	defer f.Close()

	go func(f *os.File) {
		// Sleep to wait for the file to be watched
		time.Sleep(50 * time.Millisecond)
		_, err := f.Write([]byte("anything"))
		assert.NoError(t, err)
	}(f)

	err = watchFile(f.Name())
	assert.NoError(t, err)
}

func TestUpdate(t *testing.T) {
	db := cmdutil.SetContext(t)

	expectedContent := []byte("test")
	name := "test_read_and_update.txt"
	f := &pb.File{
		Name:    name,
		Content: expectedContent,
	}

	err := update(db, f, "../testdata/test_read&update.txt")
	assert.NoError(t, err, "Updating record")

	got, err := file.Get(db, name)
	assert.NoError(t, err, "The file wasn't created")

	assert.Equal(t, expectedContent, got.Content)
}

func TestPostRun(t *testing.T) {
	NewCmd(nil).PostRun(nil, nil)
}
