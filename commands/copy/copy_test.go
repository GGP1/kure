package copy

import (
	"testing"

	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/entry"
	"github.com/GGP1/kure/pb"

	"github.com/atotto/clipboard"
	bolt "go.etcd.io/bbolt"
)

func TestCopyPassword(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	createEntry(t, db)

	cmd := NewCmd(db)
	cmd.SetArgs([]string{"test"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Failed to copy password to clipboard: %v", err)
	}

	got, err := clipboard.ReadAll()
	if err != nil {
		t.Fatalf("Failed reading from clipboard: %v", err)
	}

	if got != "Gopher" {
		t.Errorf("Expected Gopher, got %s", got)
	}
}

func TestCopyUsername(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	createEntry(t, db)

	cmd := NewCmd(db)
	cmd.SetArgs([]string{"test"})
	cmd.Flags().Set("username", "true")

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Failed to copy username to clipboard: %v", err)
	}

	got, err := clipboard.ReadAll()
	if err != nil {
		t.Fatalf("Failed reading from clipboard: %v", err)
	}

	if got != "Go" {
		t.Errorf("Expected Go, got %s", got)
	}
}

func TestCopyWithTimeout(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	createEntry(t, db)

	cmd := NewCmd(db)
	cmd.SetArgs([]string{"test"})
	cmd.Flags().Set("timeout", "1ms")

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Failed to copy password to clipboard: %v", err)
	}

	got, err := clipboard.ReadAll()
	if err != nil {
		t.Fatalf("Failed reading from clipboard: %v", err)
	}

	if got != "" {
		t.Errorf("Expected clipboard to be empty and got %s", got)
	}
}

func TestCopyErrors(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	createEntry(t, db)

	cases := []struct {
		desc string
		name string
	}{
		{desc: "Non-existent", name: "non-existent"},
		{desc: "Invalid name", name: ""},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			cmd := NewCmd(db)
			cmd.SetArgs([]string{tc.name})

			if err := cmd.Execute(); err == nil {
				t.Error("Expected to return an error and got nil")
			}
		})
	}
}

func TestPostRun(t *testing.T) {
	NewCmd(nil).PostRun(nil, nil)
}

func createEntry(t *testing.T, db *bolt.DB) {
	t.Helper()

	e := &pb.Entry{
		Name:     "test",
		Username: "Go",
		Password: "Gopher",
		Expires:  "Never",
	}

	if err := entry.Create(db, e); err != nil {
		t.Fatal(err)
	}
}
