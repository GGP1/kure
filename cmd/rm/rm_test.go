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

	if err := entry.Create(db, &pb.Entry{Name: "test", Expires: "Never"}); err != nil {
		t.Fatal(err)
	}

	buf := bytes.NewBufferString("y")

	cmd := NewCmd(db, buf)
	args := []string{"test"}

	if err := cmd.RunE(cmd, args); err != nil {
		t.Fatalf("Failed removing the entry: %v", err)
	}
}

func TestRmErrors(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	defer db.Close()

	buf := bytes.NewBufferString("y")

	cmd := NewCmd(db, buf)
	args := []string{"non-existent"}

	if err := cmd.RunE(cmd, args); err == nil {
		t.Fatal("Expected Rm() to fail but it didn't")
	}
}
