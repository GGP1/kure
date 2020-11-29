package copy

import (
	"testing"

	cmdutil "github.com/GGP1/kure/cmd"
	"github.com/GGP1/kure/db/entry"
	"github.com/GGP1/kure/pb"

	"github.com/atotto/clipboard"
	bolt "go.etcd.io/bbolt"
)

func TestCopy(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	defer db.Close()

	if err := entry.Create(db, &pb.Entry{Name: "test", Expires: "Never"}); err != nil {
		t.Fatal(err)
	}

	t.Run("Copy password", copyPassword(db))
	t.Run("Copy username", copyUsername(db))
	t.Run("Invalid copy", copyErrors(db))
}

func copyPassword(db *bolt.DB) func(*testing.T) {
	return func(t *testing.T) {
		cmd := NewCmd(db)
		cmd.Flags().Set("timeout", "30ms")
		args := []string{"test"}

		err := cmd.RunE(cmd, args)
		if err != nil {
			t.Fatalf("Failed to copy password to clipboard: %v", err)
		}

		clip, err := clipboard.ReadAll()
		if err != nil {
			t.Fatalf("Failed reading from clipboard: %v", err)
		}

		if clip != "" {
			t.Errorf("Didn't copy from clipboard")
		}
	}
}

func copyUsername(db *bolt.DB) func(*testing.T) {
	return func(t *testing.T) {
		cmd := NewCmd(db)
		args := []string{"test"}
		cmd.Flags().Set("username", "true")

		err := cmd.RunE(cmd, args)
		if err != nil {
			t.Fatalf("Failed to copy username to clipboard: %v", err)
		}

		clip, err := clipboard.ReadAll()
		if err != nil {
			t.Fatalf("Failed reading from clipboard: %v", err)
		}

		if clip != "" {
			t.Errorf("Didn't copy from clipboard")
		}
	}
}

func copyErrors(db *bolt.DB) func(*testing.T) {
	return func(t *testing.T) {
		cmd := NewCmd(db)
		args := []string{"non-existent"}

		err := cmd.RunE(cmd, args)
		if err == nil {
			t.Error("Expected to return an error and got nil")
		}
	}
}

func TestPostRun(t *testing.T) {
	cmd := NewCmd(nil)
	f := cmd.PostRun

	f(cmd, nil)
}
