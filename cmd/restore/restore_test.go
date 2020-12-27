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
	createRecords(t, db)

	cmd := argon2SubCmd(db)

	f := cmd.Flags()
	f.Set("iterations", "1")
	f.Set("memory", "5000")
	f.Set("threads", fmt.Sprintf("%d", runtime.NumCPU()))

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

func TestPassword(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	defer db.Close()

	cmd := passwordSubCmd(db)

	// For this not to fail terminal.ReadPassword() should be stubbed
	if err := cmd.RunE(cmd, nil); err == nil {
		t.Errorf("Expected an error and got nil")
	}
}

func TestInvalidBuckets(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	defer db.Close()
	objects := []string{"card", "entry", "file", "note"}

	// Test if each of the List fails when we remove its bucket
	for _, obj := range objects {
		t.Run(obj, func(t *testing.T) {

			db.Update(func(tx *bolt.Tx) error {
				bucket := "kure_" + obj
				tx.DeleteBucket([]byte(bucket))
				return nil
			})

			password := passwordSubCmd(db)
			argon2 := argon2SubCmd(db)

			if err := password.RunE(password, nil); err == nil {
				t.Error("Password: expected an error and got nil")
			}

			if err := argon2.RunE(argon2, nil); err == nil {
				t.Error("Argon2: expected an error and got nil")
			}
		})
	}
}

func TestPostRun(t *testing.T) {
	f1 := argon2SubCmd(nil)
	f1.PostRun(nil, nil)
}

func createRecords(t *testing.T, db *bolt.DB) {
	name := "test"
	cBuf, c := pb.SecureCard()
	c.Name = name

	eBuf, e := pb.SecureEntry()
	e.Name = name
	e.Expires = "Never"

	nBuf, n := pb.SecureNote()
	n.Name = name

	if err := card.Create(db, cBuf, c); err != nil {
		t.Fatalf("Failed creating the card: %v", err)
	}

	if err := entry.Create(db, eBuf, e); err != nil {
		t.Fatalf("Failed creating the entry: %v", err)
	}

	if err := file.Create(db, &pb.File{Name: name}); err != nil {
		t.Fatalf("Failed creating the file: %v", err)
	}

	if err := note.Create(db, nBuf, n); err != nil {
		t.Fatalf("Failed creating the note: %v", err)
	}
}
