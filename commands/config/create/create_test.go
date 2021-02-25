package create

import (
	"os"
	"testing"

	"github.com/spf13/viper"
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
			viper.Set("editor", tc.editor)
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

func TestMarshaler(t *testing.T) {
	data := struct {
		Test string
	}{
		Test: "Go",
	}

	cases := []struct {
		desc     string
		path     string
		expected string
	}{
		{
			desc: "Marshal to JSON",
			path: "test.json",
			expected: `{
  "Test": "Go"
}`,
		},
		{
			desc:     "Marshal to YAML",
			path:     "test.yaml",
			expected: "test: Go\n",
		},
		{
			desc:     "Marshal to TOML",
			path:     "test.toml",
			expected: "Test = \"Go\"\n",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := marshaler(data, tc.path)
			if err != nil {
				t.Fatalf("Failed marshaling data: %v", err)
			}

			if string(got) != tc.expected {
				t.Errorf("Expected \n%s, got \n%s", tc.expected, string(got))
			}
		})
	}
}

func TestMarshalerErrors(t *testing.T) {
	cases := []struct {
		desc string
		path string
	}{
		{
			desc: "Invalid extension",
			path: "some/path",
		},
		{
			desc: "Unsupported format",
			path: "some/path.xml",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			if _, err := marshaler(nil, tc.path); err == nil {
				t.Error("Expected an error and got nil")
			}
		})
	}
}

func TestPostRun(t *testing.T) {
	NewCmd().PostRun(nil, nil)
}
