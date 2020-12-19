package restore

import (
	"fmt"
	"os"
	"runtime"
	"testing"

	cmdutil "github.com/GGP1/kure/cmd"
	"github.com/GGP1/kure/db/card"
	"github.com/GGP1/kure/db/entry"
	"github.com/GGP1/kure/db/file"
	"github.com/GGP1/kure/db/note"
	"github.com/GGP1/kure/pb"

	bolt "go.etcd.io/bbolt"
)

func TestArgon2(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	defer db.Close()

	NewCmd(db)
	// Use the mock_config file as configuration
	os.Setenv("KURE_CONFIG", "testdata/mock_config.yaml")

	c := &pb.Card{
		Name: "test",
	}
	e := &pb.Entry{
		Name:    "test",
		Expires: "Never",
	}
	f := &pb.File{
		Name: "test",
	}
	n := &pb.Note{
		Name: "test",
	}

	createRecords(t, db, c, e, f, n)

	cmd := argon2SubCmd(db)

	flags := cmd.Flags()
	flags.Set("iterations", "1")
	flags.Set("memory", "5000")
	flags.Set("threads", fmt.Sprintf("%d", runtime.NumCPU()))

	if err := cmd.RunE(cmd, nil); err != nil {
		t.Errorf("Failed restoring the database with new argon2 parameters: %v", err)
	}
}

func TestArgon2InvalidParams(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	defer db.Close()

	cases := []struct {
		desc       string
		iterations string
		memory     string
		threads    string
	}{
		{
			desc:       "Invalid iterations number",
			iterations: "0",
			memory:     "1",
			threads:    fmt.Sprintf("%d", runtime.NumCPU()),
		},
		{
			desc:       "Invalid memory number",
			iterations: "1",
			memory:     "0",
			threads:    fmt.Sprintf("%d", runtime.NumCPU()),
		},
		{
			desc:       "Invalid threads number",
			iterations: "1",
			memory:     "1",
			threads:    "0",
		},
	}

	cmd := argon2SubCmd(db)

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			f := cmd.Flags()
			f.Set("iterations", tc.iterations)
			f.Set("memory", tc.memory)
			f.Set("threads", tc.threads)

			if err := cmd.RunE(cmd, nil); err == nil {
				t.Error("Expected Argon2() to fail but it didn't")
			}
		})
	}
}

func TestPostRun(t *testing.T) {
	f1 := argon2SubCmd(nil)
	f1.PostRun(nil, nil)
}

func createRecords(t *testing.T, db *bolt.DB, c *pb.Card, e *pb.Entry, f *pb.File, n *pb.Note) {
	if err := card.Create(db, c); err != nil {
		t.Fatalf("Failed creating the card: %v", err)
	}

	if err := entry.Create(db, e); err != nil {
		t.Fatalf("Failed creating the entry: %v", err)
	}

	if err := file.Create(db, f); err != nil {
		t.Fatalf("Failed creating the file: %v", err)
	}

	if err := note.Create(db, n); err != nil {
		t.Fatalf("Failed creating the note: %v", err)
	}
}
