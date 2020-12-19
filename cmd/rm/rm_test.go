package rm

import (
	"bytes"
	"testing"

	cmdutil "github.com/GGP1/kure/cmd"
	"github.com/GGP1/kure/db/entry"
	"github.com/GGP1/kure/pb"
)

func TestRm(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	defer db.Close()

	// Create the entry that is going to be removed
	if err := entry.Create(db, &pb.Entry{Name: "test", Expires: "Never"}); err != nil {
		t.Fatal(err)
	}

	buf := bytes.NewBufferString("y")

	cmd := NewCmd(db, buf)
	args := []string{"test"}

	if err := cmd.RunE(cmd, args); err != nil {
		t.Fatalf("Failed removing the entry: %v", err)
	}

	// Check if the entry was removed successfully
	if _, err := entry.Get(db, "test"); err == nil {
		t.Error("Expected Get() to fail but it didn't")
	}
}

func TestRmDir(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	defer db.Close()

	// Create the entries inside a folder to remove them
	if err := entry.Create(db, &pb.Entry{Name: "test/entry1", Expires: "Never"}); err != nil {
		t.Fatal(err)
	}
	if err := entry.Create(db, &pb.Entry{Name: "test/entry2", Expires: "Never"}); err != nil {
		t.Fatal(err)
	}

	buf := bytes.NewBufferString("y")

	cmd := NewCmd(db, buf)
	cmd.Flags().Set("dir", "true")
	args := []string{"test"}

	if err := cmd.RunE(cmd, args); err != nil {
		t.Fatalf("Failed removing the entry: %v", err)
	}

	// Check if the entries were removed successfully
	if _, err := entry.Get(db, "test/entry1"); err == nil {
		t.Error("Expected Get() to fail but it didn't")
	}
	if _, err := entry.Get(db, "test/entry2"); err == nil {
		t.Error("Expected Get() to fail but it didn't")
	}
}

func TestRmAbort(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	defer db.Close()

	buf := bytes.NewBufferString("n") // Abort operation

	cmd := NewCmd(db, buf)
	args := []string{"May the force be with you"}

	if err := cmd.RunE(cmd, args); err != nil {
		t.Errorf("Rm() failed: %v", err)
	}
}

func TestRmInvalidName(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	defer db.Close()

	cmd := NewCmd(db, nil)
	args := []string{""}

	if err := cmd.RunE(cmd, args); err == nil {
		t.Fatal("Expected Rm() to fail but it didn't")
	}
}

func TestRmNotExists(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	defer db.Close()

	buf := bytes.NewBufferString("y")

	cmd := NewCmd(db, buf)
	args := []string{"non-existent"}

	if err := cmd.RunE(cmd, args); err == nil {
		t.Fatal("Expected Rm() to fail but it didn't")
	}
}

func TestPostRun(t *testing.T) {
	cmd := NewCmd(nil, nil)
	f := cmd.PostRun
	f(cmd, nil)
}
