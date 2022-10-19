package cmdutil

import (
	"os"
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
	if got != expected {
		t.Errorf("Expected %s, got %s", expected, got)
	}
}

func TestErase(t *testing.T) {
	f, err := os.CreateTemp("", "")
	if err != nil {
		t.Fatalf("Failed creating temporary file: %v", err)
	}
	f.Close()

	if err := Erase(f.Name()); err != nil {
		t.Errorf("Failed erasing file: %v", err)
	}

	if err := Erase(f.Name()); err == nil {
		t.Error("Expected the file to be erased but it wasn't")
	}
}

func TestExistsTrue(t *testing.T) {
	db := SetContext(t, "../db/testdata/database")

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
			if err := Exists(db, name, tc.object); err == nil {
				t.Error("Expected an error but got nil")
			}

			if err := Exists(db, "naboo/tatooine/hoth", tc.object); err == nil {
				t.Error("Expected an error but got nil")
			}

			if err := Exists(db, "naboo", tc.object); err == nil {
				t.Error("Expected an error but got nil")
			}
		})
	}
}

func TestExistsFalse(t *testing.T) {
	db := SetContext(t, "../db/testdata/database")

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
			if err := Exists(db, tc.name, tc.object); err != nil {
				t.Errorf("Exists() failed: %v", err)
			}
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
			if err != nil {
				t.Errorf("Failed formatting expires: %v", err)
			}

			if got != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, got)
			}
		})
	}

	t.Run("Invalid format", func(t *testing.T) {
		if _, err := FmtExpires("invalid format"); err == nil {
			t.Error("Expected an error and got nil")
		}
	})
}

func TestMustExist(t *testing.T) {
	db := SetContext(t, "../db/testdata/database")

	name := "test/testing"
	createObjects(t, db, name)

	t.Run("Success", func(t *testing.T) {
		objects := []object{Card, Entry, File, TOTP}
		for _, obj := range objects {
			if err := MustExist(db, obj)(nil, []string{name}); err != nil {
				t.Error(err)
			}
		}
	})

	t.Run("Fail", func(t *testing.T) {
		cases := []struct {
			desc string
			name string
		}{
			{
				desc: "Record does not exist",
				name: "test",
			},
			{
				desc: "Empty name",
				name: "",
			},
			{
				desc: "Invalid name",
				name: "test//testing",
			},
		}

		for _, tc := range cases {
			t.Run(tc.desc, func(t *testing.T) {
				if err := MustExist(db, Card)(nil, []string{tc.name}); err == nil {
					t.Error("Expected an error and got nil")
				}
			})
		}
	})

	t.Run("Empty args", func(t *testing.T) {
		if err := MustExist(db, Card)(nil, []string{}); err == nil {
			t.Error("Expected an error and got nil")
		}
	})

	t.Run("Directories", func(t *testing.T) {
		t.Run("Exists", func(t *testing.T) {
			if err := MustExist(db, Card, true)(nil, []string{"test/"}); err != nil {
				t.Error(err)
			}
		})

		t.Run("Not exists", func(t *testing.T) {
			if err := MustExist(db, Card, true)(nil, []string{"unexistent/"}); err == nil {
				t.Error("Expected an error and got nil")
			}
		})
	})
}

func TestMustExistLs(t *testing.T) {
	db := SetContext(t, "../db/testdata/database")
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
					if tc.filter {
						cmd.Flags().Set("filter", "true")
					}

					if err := cmd.Args(cmd, []string{tc.name}); err != nil {
						t.Error(err)
					}
				}
			})
		}
	})

	t.Run("Fail", func(t *testing.T) {
		cmd.Args = MustExistLs(db, Entry)
		cmd.Flag("filter").Changed = false

		if err := cmd.Args(cmd, []string{"non-existent"}); err == nil {
			t.Error("Expected an error and got nil")
		}
	})
}

func TestMustNotExist(t *testing.T) {
	db := SetContext(t, "../db/testdata/database")
	cmd := &cobra.Command{}
	objects := []object{Card, Entry, File, TOTP}

	t.Run("Success", func(t *testing.T) {
		for _, obj := range objects {
			cmd.Args = MustNotExist(db, obj)
			if err := cmd.Args(cmd, []string{"test"}); err != nil {
				t.Error(err)
			}
		}
	})

	t.Run("Fail", func(t *testing.T) {
		if err := entry.Create(db, &pb.Entry{Name: "test"}); err != nil {
			t.Fatal(err)
		}
		if err := entry.Create(db, &pb.Entry{Name: "dir/"}); err != nil {
			t.Fatal(err)
		}
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
				if err := cmd.Args(cmd, []string{tc.name}); err == nil {
					t.Error("Expected an error and got nil")
				}
			})
		}
	})

	t.Run("No arguments", func(t *testing.T) {
		cmd.Args = MustNotExist(db, Entry, false)
		if err := cmd.Args(cmd, []string{}); err == nil {
			t.Error("Expected an error and got nil")
		}
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

			if got != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, got)
			}
		})
	}
}

func TestSelectEditor(t *testing.T) {
	t.Run("Default editor", func(t *testing.T) {
		expected := "nano"
		config.Set("editor", expected)
		defer config.Reset()

		got := SelectEditor()
		if got != expected {
			t.Errorf("Expected %q, got %q", expected, got)
		}
	})

	t.Run("EDITOR", func(t *testing.T) {
		expected := "editor"
		os.Setenv("EDITOR", expected)
		defer os.Unsetenv("EDITOR")

		got := SelectEditor()
		if got != expected {
			t.Errorf("Expected %q, got %q", expected, got)
		}
	})

	t.Run("VISUAL", func(t *testing.T) {
		expected := "visual"
		os.Setenv("VISUAL", expected)
		defer os.Unsetenv("VISUAL")

		got := SelectEditor()
		if got != expected {
			t.Errorf("Expected %q, got %q", expected, got)
		}
	})

	t.Run("Default", func(t *testing.T) {
		got := SelectEditor()
		if got != "vim" {
			t.Errorf("Expected vim, got %q", got)
		}
	})
}

func TestSetContext(t *testing.T) {
	path := "../db/testdata/database"
	db := SetContext(t, path)

	gotPath := db.Path()
	if gotPath != path {
		t.Errorf("Expected path to be %q, got %q", path, gotPath)
	}

	gotOpenTx := db.Stats().OpenTxN
	if gotOpenTx != 0 {
		t.Errorf("Expected to have 0 opened transactions and got %d", gotOpenTx)
	}
}

func TestWatchFile(t *testing.T) {
	f, err := os.CreateTemp("", "*")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	if _, err := f.Write([]byte("test")); err != nil {
		t.Fatal(err)
	}

	done := make(chan struct{}, 1)
	errCh := make(chan error, 1)
	go WatchFile(f.Name(), done, errCh)

	// Sleep to write after the file is being watched
	time.Sleep(50 * time.Millisecond)
	if _, err := f.Write([]byte("test-watch-file")); err != nil {
		t.Fatal(err)
	}

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
				if err := os.WriteFile(tc.filename, []byte("test error"), 0o644); err != nil {
					t.Fatalf("Failed creating file: %v", err)
				}
			}

			go WatchFile(tc.filename, done, errCh)

			// Sleep to wait until the file is created and fail once inside the for loop
			if !tc.initial {
				time.Sleep(10 * time.Millisecond)
				if err := os.Remove(tc.filename); err != nil {
					t.Fatalf("Failed removing the file: %v", err)
				}
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

		if err := WriteClipboard(cmd, 0, "", "test"); err != nil {
			t.Fatal(err)
		}

		got, err := clipboard.ReadAll()
		if err != nil {
			t.Error(err)
		}

		if got != "" {
			t.Errorf("Expected the clipboard to be empty and got %q", got)
		}
	})

	t.Run("t > 0", func(t *testing.T) {
		if err := WriteClipboard(cmd, 10*time.Millisecond, "", "test"); err != nil {
			t.Fatal(err)
		}

		got, err := clipboard.ReadAll()
		if err != nil {
			t.Error(err)
		}

		if got != "" {
			t.Errorf("Expected the clipboard to be empty and got %q", got)
		}
	})

	t.Run("t = 0", func(t *testing.T) {
		clip := "test"
		if err := WriteClipboard(cmd, 0, "", clip); err != nil {
			t.Fatal(err)
		}

		got, err := clipboard.ReadAll()
		if err != nil {
			t.Error(err)
		}

		if got != clip {
			t.Errorf("Expected %q, got %q", clip, got)
		}
	})
}

func createObjects(t *testing.T, db *bolt.DB, name string) {
	t.Helper()
	if err := entry.Create(db, &pb.Entry{Name: name}); err != nil {
		t.Fatal(err)
	}
	if err := card.Create(db, &pb.Card{Name: name}); err != nil {
		t.Fatal(err)
	}
	if err := file.Create(db, &pb.File{Name: name}); err != nil {
		t.Fatal(err)
	}
	if err := totp.Create(db, &pb.TOTP{Name: name}); err != nil {
		t.Fatal(err)
	}
}
