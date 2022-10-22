package create

import (
	"os"
	"testing"

	"github.com/GGP1/kure/config"

	"github.com/stretchr/testify/assert"
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

			err := cmd.Execute()
			assert.Error(t, err)
		})
	}

	// Cleanup
	err := os.Remove(".test.yaml")
	assert.NoError(t, err, "Failed removing file")
}

func TestPostRun(t *testing.T) {
	NewCmd().PostRun(nil, nil)
}
