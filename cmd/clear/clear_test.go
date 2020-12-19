package clear

import (
	"testing"

	"github.com/atotto/clipboard"
)

func TestClear(t *testing.T) {
	cases := []struct {
		flag  string
		value string
		run   string
	}{
		{flag: "both", value: "true"},
		{flag: "clipboard", value: "true"},
		{flag: "terminal", value: "true"},
	}

	cmd := NewCmd()
	f := cmd.Flags()
	for _, tc := range cases {
		t.Run("Clear "+tc.flag, func(t *testing.T) {
			f.Set(tc.flag, tc.value)

			if err := cmd.RunE(cmd, nil); err != nil {
				t.Fatalf("Failed clearing %s: %v", tc.flag, err)
			}

			switch tc.flag {
			case "clipboard", "both":
				got, _ := clipboard.ReadAll()
				if got != "" {
					t.Errorf("Expected clipboard to be empty but got: %s", got)
				}
			}

			cmd.ResetFlags()
		})
	}
}

func TestPostRun(t *testing.T) {
	cmd := NewCmd()
	f := cmd.PostRun
	f(cmd, nil)
}
