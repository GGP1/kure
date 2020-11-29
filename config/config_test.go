package config

import (
	"os"
	"runtime"
	"testing"

	"github.com/spf13/viper"
)

func TestLoad(t *testing.T) {
	cases := map[string]string{
		"Env var path":           "testdata/mock_config.yaml",
		"Path without extension": "testdata/mock_config",
		"Home directory":         "",
	}

	expectedStr := "test"
	expectedNum := 1
	expectedBool := true

	for k, path := range cases {
		os.Setenv("KURE_CONFIG", path)

		if err := Load(); err != nil {
			t.Fatalf("%s: Load() failed: %v", k, err)
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
	}

	cleanup(t)
}

func TestLoadErrors(t *testing.T) {
	cases := map[string]string{
		"Invalid path":    "invalid_file.yaml",
		"Home path error": "",
	}

	for k, path := range cases {
		if path == "" {
			env := "HOME"
			switch runtime.GOOS {
			case "windows":
				env = "USERPROFILE"
			case "plan9":
				env = "home"
			}
			os.Setenv(env, "")
		}

		os.Setenv("KURE_CONFIG", path)

		if err := Load(); err == nil {
			t.Fatalf("%s: expected Load() to fail but got nil", k)
		}
	}
}

// Remove config file created on user home directory.
func cleanup(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	if err := os.Remove(home + "/.kure.yaml"); err != nil {
		t.Fatal(err)
	}
}
