package create

import (
	"os"
	"testing"

	"github.com/GGP1/kure/config"
)

func TestCreateErrors(t *testing.T) {
	cases := []struct {
		desc   string
		path   string
		editor string
	}{
		{
			desc: "Invalid path",
			path: "",
		},
		{
			desc:   "Executable not found",
			path:   ".test.yaml",
			editor: "arhoan",
		},
	}

	cmd := NewCmd()

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			config.Set("editor", tc.editor)
			config.SetFilename(tc.path)

			f := cmd.Flags()
			f.Set("path", tc.path)

			if err := cmd.Execute(); err == nil {
				t.Error("Expected an error and got nil")
			}
		})
	}

	// Cleanup
	if err := os.Remove(".test.yaml"); err != nil {
		t.Errorf("Failed removing file")
	}
}

func TestPostRun(t *testing.T) {
	NewCmd().PostRun(nil, nil)
}
