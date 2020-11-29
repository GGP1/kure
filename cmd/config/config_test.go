package config

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"testing"

	cmdutil "github.com/GGP1/kure/cmd"
	"github.com/spf13/viper"
)

func TestCreate(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	defer db.Close()

	dbName := "database"
	dbPath := "../../db/testdata"
	format := "1,2,3"
	repeat := "true"
	port := "4000"
	prefix := "$"
	timeout := "10m"
	memory := "8192"
	iter := "50"

	s := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s",
		dbName, dbPath, format, repeat, port, prefix, timeout, memory, iter)
	buf := bytes.NewBufferString(s)

	cmd := NewCmd(db, buf)
	f := cmd.Flags()
	f.Set("path", "testdata/test_config")
	f.Set("create", "true")

	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("Failed creating config: %v", err)
	}

	gotDBName := viper.GetString("database.name")
	assertEqual(t, dbName, gotDBName)

	gotDBPath := viper.GetString("database.path")
	assertEqual(t, dbPath, gotDBPath)

	gotFormat := viper.GetIntSlice("entry.format")
	expectedFormat := []int{1, 2, 3}
	if !reflect.DeepEqual(gotFormat, expectedFormat) {
		t.Errorf("Expected %v, got %v", expectedFormat, gotFormat)
	}

	gotRepeat := viper.GetString("entry.repeat")
	assertEqual(t, repeat, gotRepeat)

	gotPort := viper.GetString("http.port")
	assertEqual(t, port, gotPort)

	gotPrefix := viper.GetString("session.prefix")
	assertEqual(t, prefix, gotPrefix)

	gotTimeout := viper.GetString("session.timeout")
	assertEqual(t, timeout, gotTimeout)

	gotMemory := viper.GetString("argon2id.memory")
	assertEqual(t, memory, gotMemory)

	gotIter := viper.GetString("argon2id.iterations")
	assertEqual(t, iter, gotIter)

	if err := os.Remove("testdata/test_config.yaml"); err != nil {
		t.Fatalf("Failed removing config: %v", err)
	}
}

func TestRead(t *testing.T) {
	var path string

	db := cmdutil.SetContext(t, "../../db/testdata/database")
	defer db.Close()

	cases := map[string]struct {
		setup func()
	}{
		"Empty path": {
			setup: func() {
				path = ""
				os.Setenv("KURE_CONFIG", "testdata/mock_config.yaml")
			},
		},
		"Non-empty path": {
			setup: func() {
				path = "testdata/mock_config.yaml"
				os.Setenv("KURE_CONFIG", "")
			},
		},
	}

	for k, tc := range cases {
		tc.setup()

		cmd := NewCmd(db, nil)
		f := cmd.Flags()
		f.Set("path", path)

		if err := cmd.RunE(cmd, nil); err != nil {
			t.Fatalf("%s: failed reading config: %v", k, err)
		}
	}

	os.Setenv("KURE_CONFIG", "")
}

func TestInvalidFields(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	defer db.Close()

	cases := map[string]struct {
		port   string
		format string
	}{
		"Invalid port":   {port: "", format: "1,2,3"},
		"Invalid format": {port: "8800", format: "a, b, c"},
	}

	for k, tc := range cases {
		s := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s\n%s", "", "", tc.format, "", tc.port, "", "")
		buf := bytes.NewBufferString(s)

		cmd := NewCmd(db, buf)
		f := cmd.Flags()
		f.Set("path", path)
		f.Set("create", "true")

		if err := cmd.RunE(cmd, nil); err == nil {
			t.Fatalf("%s: expected an error and got nil", k)
		}
	}
}

func TestPostRun(t *testing.T) {
	config := NewCmd(nil, nil)
	f := config.PostRun
	f(config, nil)
}

func assertEqual(t *testing.T, expected, got string) {
	if got != expected {
		t.Errorf("Expected %s, got %s", expected, got)
	}
}
