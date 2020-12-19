package config

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"runtime"
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
	threads := "2"

	s := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s",
		dbName, dbPath, format, repeat, port, prefix, timeout, memory, iter, threads)
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

	gotMemory := viper.GetString("argon2.memory")
	assertEqual(t, memory, gotMemory)

	gotIter := viper.GetString("argon2.iterations")
	assertEqual(t, iter, gotIter)

	gotThreads := viper.GetString("argon2.threads")
	assertEqual(t, threads, gotThreads)

	if err := os.Remove("testdata/test_config.yaml"); err != nil {
		t.Fatalf("Failed removing config: %v", err)
	}
}

func TestRead(t *testing.T) {
	var path string

	db := cmdutil.SetContext(t, "../../db/testdata/database")
	defer db.Close()

	cases := []struct {
		desc  string
		setup func()
	}{
		{
			desc: "Empty path",
			setup: func() {
				path = ""
				os.Setenv("KURE_CONFIG", "testdata/mock_config.yaml")
			},
		},
		{
			desc: "Non-empty path",
			setup: func() {
				path = "testdata/mock_config.yaml"
				os.Setenv("KURE_CONFIG", "")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			tc.setup()

			cmd := NewCmd(db, nil)
			f := cmd.Flags()
			f.Set("path", path)

			if err := cmd.RunE(cmd, nil); err != nil {
				t.Fatalf("Dailed reading config: %v", err)
			}
		})
	}

	// Reset after finished
	os.Setenv("KURE_CONFIG", "")
}

func TestInvalidFields(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	defer db.Close()

	cases := []struct {
		desc   string
		port   string
		format string
	}{
		{desc: "Invalid format", port: "8800", format: "a, b, c"},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			s := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s\n%s", "", "", tc.format, "", tc.port, "", "")
			buf := bytes.NewBufferString(s)

			cmd := NewCmd(db, buf)
			f := cmd.Flags()
			f.Set("path", path)
			f.Set("create", "true")

			if err := cmd.RunE(cmd, nil); err == nil {
				t.Fatal("Expected an error and got nil")
			}
		})
	}
}

func TestTest(t *testing.T) {
	cases := []struct {
		desc       string
		iterations uint32
		memory     uint32
		threads    uint8
	}{
		{
			desc:       "Test 1",
			iterations: 1,
			memory:     400000,
			threads:    uint8(runtime.NumCPU() - 2),
		},
		{
			desc:       "Test 2",
			iterations: 15,
			memory:     3000,
			threads:    uint8(runtime.NumCPU()),
		},
		{
			desc:       "Test 3",
			iterations: 2,
			memory:     716500,
			threads:    uint8(runtime.NumCPU() - 1),
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			cmd := testSubCmd()
			f := cmd.Flags()
			f.Set("iterations", fmt.Sprintf("%d", tc.iterations))
			f.Set("memory", fmt.Sprintf("%d", tc.memory))
			f.Set("threads", fmt.Sprintf("%d", tc.threads))

			if err := cmd.RunE(cmd, nil); err != nil {
				t.Fatalf("Test sub command failed: %v", err)
			}
		})
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
