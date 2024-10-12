package cmdutil

import (
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/GGP1/kure/config"
	"github.com/GGP1/kure/db/card"
	"github.com/GGP1/kure/db/entry"
	"github.com/GGP1/kure/db/file"
	"github.com/GGP1/kure/db/totp"
	"github.com/GGP1/kure/orderedmap"
	"github.com/GGP1/kure/pb"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	bolt "go.etcd.io/bbolt"
)

func TestBuildBox(t *testing.T) {
	expected := `╭────── Box ─────╮
│ Jedi   │ Luke  │
│ Hobbit │ Frodo │
│        │ Sam   │
│ Wizard │ Harry │
╰────────────────╯`

	mp := orderedmap.New()
	mp.Set("Jedi", "Luke")
	mp.Set("Hobbit", `Frodo
Sam`)
	mp.Set("Wizard", "Harry")

	got := BuildBox("test/box", mp)
	assert.Equal(t, expected, got)
}

func TestErase(t *testing.T) {
	f, err := os.CreateTemp("", "")
	assert.NoError(t, err, "Failed creating temporary file")
	f.Close()

	err = Erase(f.Name())
	assert.NoError(t, err, "Failed erasing file")

	err = Erase(f.Name())
	assert.Error(t, err, "Expected the file to be erased")
}

func TestExistsTrue(t *testing.T) {
	db := SetContext(t)

	name := "naboo/tatooine"
	createObjects(t, db, name)

	cases := []struct {
		desc   string
		object object
	}{
		{
			desc:   "card",
			object: Card,
		},
		{
			desc:   "entry",
			object: Entry,
		},
		{
			desc:   "file",
			object: File,
		},
		{
			desc:   "totp",
			object: TOTP,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			err := Exists(db, name, tc.object)
			assert.Error(t, err)

			err = Exists(db, "naboo/tatooine/hoth", tc.object)
			assert.Error(t, err)

			err = Exists(db, "naboo", tc.object)
			assert.Error(t, err)
		})
	}
}

func TestExistsFalse(t *testing.T) {
	db := SetContext(t)

	cases := []struct {
		desc   string
		name   string
		object object
	}{
		{
			desc:   "card",
			name:   "test",
			object: Card,
		},
		{
			desc:   "entry",
			name:   "test",
			object: Entry,
		},
		{
			desc:   "file",
			name:   "testing/test",
			object: File,
		},
		{
			desc:   "totp",
			name:   "testing",
			object: TOTP,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			err := Exists(db, tc.name, tc.object)
			assert.NoError(t, err)
		})
	}
}

func TestFmtExpires(t *testing.T) {
	cases := []struct {
		desc     string
		expires  string
		expected string
	}{
		{
			desc:     "Never",
			expires:  "Never",
			expected: "Never",
		},
		{
			desc:     "dd/mm/yy",
			expires:  "26/06/2029",
			expected: "Tue, 26 Jun 2029 00:00:00 +0000",
		},
		{
			desc:     "yy/mm/dd",
			expires:  "2029/06/26",
			expected: "Tue, 26 Jun 2029 00:00:00 +0000",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := FmtExpires(tc.expires)
			assert.NoError(t, err, "Failed formatting expires")

			assert.Equal(t, tc.expected, got)
		})
	}

	t.Run("Invalid format", func(t *testing.T) {
		_, err := FmtExpires("invalid format")
		assert.Error(t, err)
	})
}

func TestMustExist(t *testing.T) {
	db := SetContext(t)

	name := "test/testing"
	createObjects(t, db, name)

	t.Run("Success", func(t *testing.T) {
		objects := []object{Card, Entry, File, TOTP}
		for _, obj := range objects {
			err := MustExist(db, obj)(nil, []string{name})
			assert.NoError(t, err)
		}
	})

	t.Run("Fail", func(t *testing.T) {
		cases := []struct {
			desc       string
			name       string
			errMessage string
		}{
			{
				desc:       "Record does not exist",
				name:       "test",
				errMessage: "\"test\" does not exist. Did you mean \"test/testing\"?",
			},
			{
				desc:       "Record does not exist 2",
				name:       "test/tstng",
				errMessage: "\"test/tstng\" does not exist. Did you mean \"test/testing\"?",
			},
			{
				desc:       "Empty name",
				name:       "",
				errMessage: "invalid name",
			},
			{
				desc:       "Invalid name",
				name:       "test//testing",
				errMessage: "invalid name",
			},
		}

		for _, tc := range cases {
			t.Run(tc.desc, func(t *testing.T) {
				err := MustExist(db, Card)(nil, []string{tc.name})
				assert.Error(t, err)

				assert.Equal(t, tc.errMessage, err.Error())
			})
		}
	})

	t.Run("Empty args", func(t *testing.T) {
		err := MustExist(db, Card)(nil, []string{})
		assert.Error(t, err)
	})

	t.Run("Directories", func(t *testing.T) {
		t.Run("Exists", func(t *testing.T) {
			err := MustExist(db, Card, true)(nil, []string{"test/"})
			assert.NoError(t, err)
		})

		t.Run("Not exists", func(t *testing.T) {
			err := MustExist(db, Card, true)(nil, []string{"unexistent/"})
			assert.Error(t, err)
		})
	})
}

func TestMustExistLs(t *testing.T) {
	db := SetContext(t)
	cmd := &cobra.Command{}
	cmd.Flags().Bool("filter", false, "")
	objects := []object{Card, Entry, File, TOTP}

	name := "test"
	createObjects(t, db, name)

	cases := []struct {
		desc   string
		name   string
		filter bool
	}{
		{
			desc: "Found name",
			name: name,
		},
		{
			desc: "Empty name",
			name: "",
		},
		{
			desc:   "Filtering",
			name:   "t",
			filter: true,
		},
	}

	t.Run("Success", func(t *testing.T) {
		for _, tc := range cases {
			t.Run(tc.desc, func(t *testing.T) {
				for _, obj := range objects {
					cmd.Args = MustExistLs(db, obj)
					cmd.Flags().Set("filter", strconv.FormatBool(tc.filter))

					err := cmd.Args(cmd, []string{tc.name})
					assert.NoError(t, err)
				}
			})
		}
	})

	t.Run("Fail", func(t *testing.T) {
		cmd.Args = MustExistLs(db, Entry)
		cmd.Flag("filter").Changed = false

		err := cmd.Args(cmd, []string{"non-existent"})
		assert.Error(t, err)
	})
}

func TestMustNotExist(t *testing.T) {
	db := SetContext(t)
	cmd := &cobra.Command{}
	objects := []object{Card, Entry, File, TOTP}

	t.Run("Success", func(t *testing.T) {
		for _, obj := range objects {
			cmd.Args = MustNotExist(db, obj)
			err := cmd.Args(cmd, []string{"test"})
			assert.NoError(t, err)
		}
	})

	t.Run("Fail", func(t *testing.T) {
		err := entry.Create(db, &pb.Entry{Name: "test"})
		assert.NoError(t, err)
		err = entry.Create(db, &pb.Entry{Name: "dir/"})
		assert.NoError(t, err)

		cases := []struct {
			desc     string
			name     string
			allowDir []bool
		}{
			{
				desc: "Exists",
				name: "test",
			},
			{
				desc:     "Directory exists",
				name:     "dir/",
				allowDir: []bool{true},
			},
			{
				desc: "Empty name",
				name: "",
			},
			{
				desc: "Invalid name",
				name: "testing//test",
			},
		}

		for _, tc := range cases {
			t.Run(tc.desc, func(t *testing.T) {
				cmd.Args = MustNotExist(db, Entry, tc.allowDir...)
				err := cmd.Args(cmd, []string{tc.name})
				assert.Error(t, err)
			})
		}
	})

	t.Run("No arguments", func(t *testing.T) {
		cmd.Args = MustNotExist(db, Entry, false)
		err := cmd.Args(cmd, []string{})
		assert.Error(t, err)
	})
}

func TestNormalizeName(t *testing.T) {
	cases := []struct {
		desc     string
		name     string
		expected string
		allowDir []bool
	}{
		{
			desc:     "Normalize",
			name:     " / Go/Forum / ",
			expected: "go/forum",
		},
		{
			desc:     "Empty",
			name:     "",
			expected: "",
		},
		{
			desc:     "Allow dir",
			name:     "testing/",
			expected: "testing/",
			allowDir: []bool{true},
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			got := NormalizeName(tc.name, tc.allowDir...)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestSelectEditor(t *testing.T) {
	t.Run("Default editor", func(t *testing.T) {
		expected := "nano"
		config.Set("editor", expected)
		defer config.Reset()

		got := SelectEditor()
		assert.Equal(t, expected, got)
	})

	t.Run("EDITOR", func(t *testing.T) {
		expected := "editor"
		os.Setenv("EDITOR", expected)
		defer os.Unsetenv("EDITOR")

		got := SelectEditor()
		assert.Equal(t, expected, got)
	})

	t.Run("VISUAL", func(t *testing.T) {
		expected := "visual"
		os.Setenv("VISUAL", expected)
		defer os.Unsetenv("VISUAL")

		got := SelectEditor()
		assert.Equal(t, expected, got)
	})

	t.Run("Default", func(t *testing.T) {
		got := SelectEditor()
		assert.Equal(t, "vim", got)
	})
}

func TestSupportedManagers(t *testing.T) {
	t.Run("Supported", func(t *testing.T) {
		list := []string{"1password", "bitwarden", "keepass", "keepassx", "keepassxc", "lastpass"}
		for _, name := range list {
			err := SupportedManagers()(nil, []string{name})
			assert.NoError(t, err)
		}
	})

	t.Run("Unsupported", func(t *testing.T) {
		list := []string{"", "unsupported"}
		for _, name := range list {
			err := SupportedManagers()(nil, []string{name})
			assert.Error(t, err)
		}
	})
}

func TestWatchFile(t *testing.T) {
	f, err := os.CreateTemp("", "*")
	assert.NoError(t, err)
	defer f.Close()

	_, err = f.Write([]byte("test"))
	assert.NoError(t, err)

	done := make(chan struct{}, 1)
	errCh := make(chan error, 1)
	go WatchFile(f.Name(), done, errCh)

	// Sleep to write after the file is being watched
	time.Sleep(50 * time.Millisecond)
	_, err = f.Write([]byte("test-watch-file"))
	assert.NoError(t, err)

	select {
	case <-done:

	case <-errCh:
		t.Errorf("Watching file failed: %v", err)
	}
}

func TestWatchFileErrors(t *testing.T) {
	cases := []struct {
		desc     string
		filename string
		initial  bool
	}{
		{
			desc:     "Initial stat error",
			filename: "test_error.json",
			initial:  true,
		},
		{
			desc:     "For loop stat error",
			filename: "test_error.json",
			initial:  false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			done := make(chan struct{}, 1)
			errCh := make(chan error, 1)

			if !tc.initial {
				err := os.WriteFile(tc.filename, []byte("test error"), 0o644)
				assert.NoError(t, err, "Failed creating file")
			}

			go WatchFile(tc.filename, done, errCh)

			// Sleep to wait until the file is created and fail once inside the for loop
			if !tc.initial {
				time.Sleep(10 * time.Millisecond)
				err := os.Remove(tc.filename)
				assert.NoError(t, err, "Failed removing the file")
			}

			select {
			case <-done:
				t.Error("Expected an error and it succeeded")

			case <-errCh:
			}
		})
	}
}

func TestWriteClipboard(t *testing.T) {
	if clipboard.Unsupported {
		t.Skip("No clipboard utilities available")
	}

	cmd := &cobra.Command{}

	t.Run("Default timeout", func(t *testing.T) {
		config.Set("clipboard.timeout", 10*time.Millisecond)
		defer config.Reset()

		err := WriteClipboard(cmd, 0, "", "test")
		assert.NoError(t, err)

		got, err := clipboard.ReadAll()
		assert.NoError(t, err)

		assert.Empty(t, got, "Expected the clipboard to be empty")
	})

	t.Run("t > 0", func(t *testing.T) {
		err := WriteClipboard(cmd, 10*time.Millisecond, "", "test")
		assert.NoError(t, err)

		got, err := clipboard.ReadAll()
		assert.NoError(t, err)

		assert.Empty(t, got, "Expected the clipboard to be empty")
	})

	t.Run("t = 0", func(t *testing.T) {
		clip := "test"
		err := WriteClipboard(cmd, 0, "", clip)
		assert.NoError(t, err)

		got, err := clipboard.ReadAll()
		assert.NoError(t, err)

		assert.Equal(t, clip, got)
	})
}

func TestFormatSuggestions(t *testing.T) {
	cases := []struct {
		desc           string
		expectedResult string
		suggestions    []string
	}{
		{
			desc: "One suggestion",
			suggestions: []string{
				"car",
			},
			expectedResult: "\"car\"",
		},
		{
			desc: "Two suggestions",
			suggestions: []string{
				"car",
				"cur",
			},
			expectedResult: "\"car\" or \"cur\"",
		},
		{
			desc: "Three suggestions",
			suggestions: []string{
				"car",
				"cur",
				"core",
			},
			expectedResult: "\"car\", \"cur\" or \"core\"",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			suggestion := formatSuggestions(tc.suggestions)

			assert.Equal(t, tc.expectedResult, suggestion)
		})
	}
}

func TestGetNameSuggestions(t *testing.T) {
	cases := []struct {
		desc                string
		names               []string
		name                string
		expectedSuggestions []string
	}{
		{
			desc: "By distance",
			names: []string{
				"cat",
				"bat",
				"category",
				"rat",
				"car",
				"pop",
			},
			name: "hay",
			expectedSuggestions: []string{
				"cat",
				"bat",
				"rat",
				"car",
			},
		},
		{
			desc: "By prefix",
			names: []string{
				"category",
				"careful",
				"career",
			},
			name: "car",
			expectedSuggestions: []string{
				"careful",
				"career",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			suggestions := getNameSuggestions(tc.name, tc.names)

			assert.Equal(t, tc.expectedSuggestions, suggestions)
		})
	}
}

func TestLevenshteinDistance(t *testing.T) {
	cases := []struct {
		nameA            string
		nameB            string
		expectedDistance int
	}{
		{
			nameA:            "car",
			nameB:            "cat",
			expectedDistance: 1,
		},
		{
			nameA:            "car",
			nameB:            "hat",
			expectedDistance: 2,
		},
		{
			nameA:            "car",
			nameB:            "carry",
			expectedDistance: 2,
		},
		{
			nameA:            "car",
			nameB:            "born",
			expectedDistance: 3,
		},
		{
			nameA:            "car",
			nameB:            "karting",
			expectedDistance: 5,
		},
	}

	for _, tc := range cases {
		t.Run(strconv.Itoa(tc.expectedDistance), func(t *testing.T) {
			distance := levenshteinDistance(tc.nameA, tc.nameB)

			assert.Equal(t, tc.expectedDistance, distance)
		})
	}
}

func createObjects(t *testing.T, db *bolt.DB, name string) {
	t.Helper()
	err := entry.Create(db, &pb.Entry{Name: name})
	assert.NoError(t, err)
	err = card.Create(db, &pb.Card{Name: name})
	assert.NoError(t, err)
	err = file.Create(db, &pb.File{Name: name})
	assert.NoError(t, err)
	err = totp.Create(db, &pb.TOTP{Name: name})
	assert.NoError(t, err)
}
