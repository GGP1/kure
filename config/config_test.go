package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/spf13/viper"
)

func TestLoad(t *testing.T) {
	cases := []struct {
		desc string
		path string
	}{
		{desc: "Env var path", path: "testdata/mock_config.yaml"},
		{desc: "Path without extension", path: "testdata/mock_config.yaml"},
		{desc: "Home directory", path: ""},
	}

	// Create the home folder with mode 0666 so the command uses it
	// Otherwise github actions' test will fail
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("couldn't find the home directory: %v", err)
	}
	home = filepath.Join(home, ".kure")
	if err := os.MkdirAll(home, 0666); err != nil {
		t.Fatalf("couldn't create the configuration directory: %v", err)
	}

	expectedStr := "test"
	expectedNum := 1
	expectedBool := true

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			os.Setenv("KURE_CONFIG", tc.path)

			if err := Load(); err != nil {
				t.Fatalf("Load() failed: %v", err)
			}

			gotStr := viper.Get("test.string")
			if gotStr != expectedStr {
				t.Errorf("Expected %s, got %s", expectedStr, gotStr)
			}

			gotNum := viper.Get("test.number")
			if gotNum != expectedNum {
				t.Errorf("Expected %d, got %d", expectedNum, gotNum)
			}

			gotBool := viper.Get("test.bool")
			if gotBool != expectedBool {
				t.Errorf("Expected %t, got %t", expectedBool, gotBool)
			}
		})
	}

	cleanup(t)
}

func TestLoadErrors(t *testing.T) {
	cases := []struct {
		desc string
		path string
	}{
		{desc: "Invalid path", path: "invalid_file.yaml"},
		{desc: "Invalid extension", path: "invalid_extension"},
		{desc: "Home path error", path: ""},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
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

			if err := Load(); err == nil {
				t.Error("Expected Load() to fail but got nil")
			}
		})
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
