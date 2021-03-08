package config

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/awnumar/memguard"
)

func TestInit(t *testing.T) {
	cases := []struct {
		desc string
		path string
	}{
		{desc: "Env var path", path: "testdata/mock_config.yaml"},
		{desc: "Home directory", path: ""},
	}

	expectedStr := "test"
	expectedNum := 1
	expectedBool := true

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			os.Setenv("KURE_CONFIG", tc.path)

			if err := Init(); err != nil {
				t.Fatalf("Init() failed: %v", err)
			}

			gotStr := config.Get("test.string")
			if gotStr != expectedStr {
				t.Errorf("Expected %s, got %s", expectedStr, gotStr)
			}

			gotNum := config.Get("test.number")
			if gotNum != expectedNum {
				t.Errorf("Expected %d, got %d", expectedNum, gotNum)
			}

			gotBool := config.Get("test.bool")
			if gotBool != expectedBool {
				t.Errorf("Expected %t, got %t", expectedBool, gotBool)
			}
		})
	}

	cleanup(t)
}

func TestInitErrors(t *testing.T) {
	cases := []struct {
		desc string
		path string
		set  func()
	}{
		{
			desc: "Invalid path",
			path: "invalid_file.yaml",
			set:  func() {},
		},
		{
			desc: "Invalid extension",
			path: "testdata/mock_config",
			set:  func() {},
		},
		{
			desc: "Home path error",
			path: "",
			set:  func() {},
		},
		{
			desc: "Auth key presence",
			path: "testdata/mock_config.yaml",
			set:  func() { Set("auth", 200) }},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			tc.set()
			if tc.path == "" {
				env := "HOME"
				switch runtime.GOOS {
				case "windows":
					env = "USERPROFILE"
				case "plan9":
					env = "home"
				}
				os.Setenv(env, "")
			}

			os.Setenv("KURE_CONFIG", tc.path)

			if err := Init(); err == nil {
				t.Error("Expected Init() to fail but got nil")
			}
		})
	}
}

func TestFileUsed(t *testing.T) {
	expected := "test"
	config.filename = expected

	got := FileUsed()
	if got != expected {
		t.Errorf("Expected %q, got %q", expected, got)
	}
}

func TestGetEnclave(t *testing.T) {
	key := "test"
	expected := memguard.NewEnclave([]byte("test"))
	config.mp = map[string]interface{}{
		key: expected,
	}

	got := GetEnclave(key)
	if got != expected {
		t.Errorf("Expected %v, got %v", expected, got)
	}

	t.Run("Nil", func(t *testing.T) {
		config.mp = map[string]interface{}{}
		got := GetEnclave(key)
		if got != nil {
			t.Errorf("Expected nil and got %v", got)
		}
	})
}

func TestGetDuration(t *testing.T) {
	key := "test"
	expected := time.Duration(10)
	config.mp = map[string]interface{}{
		key: expected,
	}

	got := GetDuration(key)
	if got != expected {
		t.Errorf("Expected %v, got %v", expected, got)
	}
}

func TestGetString(t *testing.T) {
	key := "test"
	expected := "getstring"
	config.mp = map[string]interface{}{
		key: expected,
	}

	got := GetString(key)
	if got != expected {
		t.Errorf("Expected %v, got %v", expected, got)
	}
}

func TestGetUint32(t *testing.T) {
	key := "test"
	expected := uint32(12)
	config.mp = map[string]interface{}{
		key: expected,
	}

	got := GetUint32(key)
	if got != expected {
		t.Errorf("Expected %v, got %v", expected, got)
	}
}

func TestSetDefaults(t *testing.T) {
	defaults := map[string]string{
		"clipboard.timeout": "0s",
		"database.path":     "test",
		"editor":            "vim",
		"keyfile.path":      "",
		"session.prefix":    "kure:~ $",
		"session.timeout":   "0s",
	}

	SetDefaults("test")

	for k, v := range defaults {
		got := Get(k)
		if got != v {
			t.Errorf("Expected %q, got %q", v, got)
		}
	}
}

func TestSetFile(t *testing.T) {
	expected := "test"
	SetFile(expected)

	if config.filename != expected {
		t.Errorf("Expected %q, got %q", expected, config.filename)
	}
}

func TestWriteStruct(t *testing.T) {
	filename := "test.toml"
	temp := config.mp
	config.mp = map[string]interface{}{
		"clipboard": map[string]interface{}{
			"timeout": "",
		},
		"database": map[string]interface{}{
			"path": "",
		},
		"editor": "",
		"keyfile": map[string]interface{}{
			"path": "",
		},
		"session": map[string]interface{}{
			"prefix":  "",
			"timeout": "",
		},
	}

	expected, _ := config.marshal(filepath.Ext(filename))
	config.mp = temp

	if err := WriteStruct(filename); err != nil {
		t.Fatalf("Failed writing config struct: %v", err)
	}

	got, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed reading file: %v", err)
	}
	os.Remove(filename)

	if !bytes.Equal(got, expected) {
		t.Errorf("Expected %s,\n got %s", string(expected), string(got))
	}
}

// Remove config file created on user home directory.
func cleanup(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	if os.RemoveAll(filepath.Join(home, ".kure")); err != nil {
		t.Fatal(err)
	}
}
