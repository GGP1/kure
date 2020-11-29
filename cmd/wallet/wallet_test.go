package wallet

import (
	"bytes"
	"os"
	"reflect"
	"testing"

	cmdutil "github.com/GGP1/kure/cmd"
	"github.com/GGP1/kure/pb"

	bolt "go.etcd.io/bbolt"
)

func TestWallet(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	defer db.Close()

	NewCmd(db)
	t.Run("Add", add(db))
	t.Run("Copy", copy(db))
	t.Run("Ls", ls(db))
	t.Run("Rm", rm(db))
}

func add(db *bolt.DB) func(*testing.T) {
	return func(t *testing.T) {
		cases := map[string]struct {
			name string
			pass bool
		}{
			"Add":          {name: "test", pass: true},
			"Invalid name": {name: "", pass: false},
			"Name too long": {name: `012345678901234567890123456789
			0123456789012345678901234567890123456789`, pass: false},
		}

		cmd := addSubCmd(db, os.Stdin)

		for k, tc := range cases {
			args := []string{tc.name}

			err := cmd.RunE(cmd, args)
			assertError(t, k, "add", err, tc.pass)
		}

		// Test already exists separate to avoid "Add" executing after it and fail
		args := []string{"test"}
		err := cmd.RunE(cmd, args)
		if err == nil {
			t.Errorf("%q already exists and we expected an error but got nil", args[0])
		}
	}
}

func copy(db *bolt.DB) func(*testing.T) {
	return func(t *testing.T) {
		cases := map[string]struct {
			name    string
			timeout string
			pass    bool
		}{
			"Copy":                {name: "test", pass: true},
			"Copy w/Timeout":      {name: "test", timeout: "30ms", pass: true},
			"Invalid name":        {name: "", pass: false},
			"Non existent wallet": {name: "non-existent", pass: false},
		}

		cmd := copySubCmd(db)
		f := cmd.Flags()

		for k, tc := range cases {
			args := []string{tc.name}
			if tc.timeout != "" {
				f.Set("timeout", tc.timeout)
			}

			err := cmd.RunE(cmd, args)
			assertError(t, k, "copy", err, tc.pass)
		}
	}
}

func ls(db *bolt.DB) func(*testing.T) {
	return func(t *testing.T) {
		cases := map[string]struct {
			name   string
			filter string
			hide   string
			pass   bool
		}{
			"List one":              {name: "test", pass: true},
			"Filter by name":        {name: "test", filter: "true", pass: true},
			"List all":              {name: "", pass: true},
			"List one and hide":     {name: "test", hide: "true", pass: true},
			"Wallet does not exist": {name: "non-existent", filter: "false", pass: false},
			"No wallets found":      {name: "non-existent", filter: "true", pass: false},
		}

		cmd := lsSubCmd(db)
		f := cmd.Flags()

		for k, tc := range cases {
			args := []string{tc.name}
			f.Set("filter", tc.filter)
			f.Set("hide", tc.hide)

			err := cmd.RunE(cmd, args)
			assertError(t, k, "ls", err, tc.pass)
		}
	}
}

func rm(db *bolt.DB) func(*testing.T) {
	return func(t *testing.T) {
		cases := map[string]struct {
			name  string
			input string
			pass  bool
		}{
			"Remove":              {name: "test", input: "y", pass: true},
			"Do not proceed":      {name: "quit", input: "n", pass: true},
			"Invalid name":        {name: "", pass: false},
			"Non existent wallet": {name: "non-existent", input: "y", pass: false},
		}

		for k, tc := range cases {
			buf := bytes.NewBufferString(tc.input)

			cmd := rmSubCmd(db, buf)
			args := []string{tc.name}

			err := cmd.RunE(cmd, args)
			assertError(t, k, "rm", err, tc.pass)
		}
	}
}

func TestWalletAddInput(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	defer db.Close()

	expected := &pb.Wallet{
		Name:         "test",
		Type:         "type",
		ScriptType:   "script",
		KeystoreType: "keystore",
		SeedPhrase:   "seed",
		PublicKey:    "public",
		PrivateKey:   "private",
	}

	buf := bytes.NewBufferString("type\nscript\nkeystore\nseed\npublic\nprivate")

	got, err := walletInput(db, "Test", buf)
	if err != nil {
		t.Fatalf("Failed creating the card: %v", err)
	}

	if !reflect.DeepEqual(expected, got) {
		t.Errorf("Expected %v, got %+v", expected, got)
	}
}

func TestPostRun(t *testing.T) {
	copy := copySubCmd(nil)
	f := copy.PostRun
	f(copy, nil)

	ls := lsSubCmd(nil)
	f2 := ls.PostRun
	f2(ls, nil)
}

func assertError(t *testing.T, name, funcName string, err error, pass bool) {
	if err != nil && pass {
		t.Errorf("%s: failed running %s: %v", name, funcName, err)
	}
	if err == nil && !pass {
		t.Errorf("%s: expected an error and got nil", name)
	}
}
