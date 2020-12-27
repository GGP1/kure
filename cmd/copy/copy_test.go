package copy

import (
	"testing"

	cmdutil "github.com/GGP1/kure/cmd"
	"github.com/GGP1/kure/db/entry"
	"github.com/GGP1/kure/pb"

	"github.com/atotto/clipboard"
	bolt "go.etcd.io/bbolt"
)

func TestCopyPassword(t *testing.T) {
	db := createEntry(t)
	defer db.Close()

	cmd := NewCmd(db)
	args := []string{"test"}

	if err := cmd.RunE(cmd, args); err != nil {
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
	db := createEntry(t)
	defer db.Close()

	cmd := NewCmd(db)
	args := []string{"test"}
	cmd.Flags().Set("username", "true")

	if err := cmd.RunE(cmd, args); err != nil {
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
	db := createEntry(t)
	defer db.Close()

	cmd := NewCmd(db)
	args := []string{"test"}
	cmd.Flags().Set("timeout", "1ms")

	if err := cmd.RunE(cmd, args); err != nil {
		t.Fatalf("Failed to copy password to clipboard: %v", err)
	}

	got, err := clipboard.ReadAll()
	if err != nil {
		t.Fatalf("Failed reading from clipboard: %v", err)
	}

	if got != "" {
		t.Errorf("Expected clipboard to be empty but got %s", got)
	}
}

func TestCopyErrors(t *testing.T) {
	db := createEntry(t)
	defer db.Close()

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
			args := []string{tc.name}

			err := cmd.RunE(cmd, args)
			if err == nil {
				t.Error("Expected to return an error and got nil")
			}
		})
	}
}

func TestPostRun(t *testing.T) {
	cmd := NewCmd(nil)
	f := cmd.PostRun

	f(cmd, nil)
}

func createEntry(t *testing.T) *bolt.DB {
	db := cmdutil.SetContext(t, "../../db/testdata/database")

	lockedBuf, e := pb.SecureEntry()
	e.Name = "test"
	e.Username = "Go"
	e.Password = "Gopher"
	e.Expires = "Never"

	if err := entry.Create(db, lockedBuf, e); err != nil {
		t.Fatal(err)
	}

	return db
}
