package clear

import (
	"bytes"
	"runtime"
	"testing"

	"github.com/atotto/clipboard"
)

func TestClear(t *testing.T) {
	if clipboard.Unsupported {
		t.Skip("No clipboard utilities available")
	}

	cases := []struct {
		flag  string
		value string
		run   string
	}{
		{flag: "clipboard", value: "true"},
		{flag: "terminal", value: "true"},
	}

	cmd := NewCmd()
	for _, tc := range cases {
		t.Run("Clear "+tc.flag, func(t *testing.T) {
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			f := cmd.Flags()
			f.Set(tc.flag, tc.value)

			if tc.flag == "terminal" && runtime.GOOS == "darwin" {
				t.Skip("macOS returns an exit status 1 when clearing the terminal")
			}

			if err := cmd.Execute(); err != nil {
				t.Fatalf("Failed: %v", err)
			}

			if tc.flag == "clipboard" {
				got, _ := clipboard.ReadAll()
				if got != "" {
					t.Errorf("Expected clipboard to be empty but got: %s", got)
				}
			}
		})
	}
}

func TestPostRun(t *testing.T) {
	NewCmd().PostRun(nil, nil)
}
