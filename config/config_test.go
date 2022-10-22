package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/awnumar/memguard"
	"github.com/stretchr/testify/assert"
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

			err := Init()
			assert.NoError(t, err)

			gotStr := config.Get("test.string")
			assert.Equal(t, expectedStr, gotStr)
			gotNum := config.Get("test.number")
			assert.Equal(t, expectedNum, gotNum)
			gotBool := config.Get("test.bool")
			assert.Equal(t, expectedBool, gotBool)
		})
	}

	cleanup(t)
}

func TestInitErrors(t *testing.T) {
	cases := []struct {
		set  func()
		desc string
		path string
	}{
		{
			desc: "Invalid path",
			path: "invalid_file.yaml",
		},
		{
			desc: "Invalid extension",
			path: "testdata/mock_config",
		},
		{
			desc: "Home path error",
			path: "",
		},
		{
			desc: "Auth key presence",
			path: "testdata/mock_config.yaml",
			set:  func() { Set("auth", 200) },
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			if tc.set != nil {
				tc.set()
			}
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

			err := Init()
			assert.Error(t, err)
		})
	}
}

func TestFilename(t *testing.T) {
	expected := "test"
	config.filename = expected

	got := Filename()
	assert.Equal(t, expected, got)
}

func TestGetEnclave(t *testing.T) {
	key := "test"
	expected := memguard.NewEnclave([]byte("test"))
	config.mp = map[string]interface{}{
		key: expected,
	}

	got := GetEnclave(key)
	assert.Equal(t, expected, got)

	t.Run("Nil", func(t *testing.T) {
		config.mp = map[string]interface{}{}
		got := GetEnclave(key)
		assert.Nil(t, got)
	})
}

func TestGetDuration(t *testing.T) {
	key := "test"
	expected := time.Duration(10)
	config.mp = map[string]interface{}{
		key: expected,
	}

	got := GetDuration(key)
	assert.Equal(t, expected, got)
}

func TestGetString(t *testing.T) {
	key := "test"
	expected := "getstring"
	config.mp = map[string]interface{}{
		key: expected,
	}

	got := GetString(key)
	assert.Equal(t, expected, got)
}

func TestStringMapString(t *testing.T) {
	expected := map[string]string{"login": "test"}
	config.mp = map[string]interface{}{
		"scripts": map[string]string{
			"login": "test",
		},
	}

	got := GetStringMapString("scripts")
	assert.Equal(t, expected, got)
}

func TestGetUint32(t *testing.T) {
	key := "test"
	expected := uint32(12)
	config.mp = map[string]interface{}{
		key: expected,
	}

	got := GetUint32(key)
	assert.Equal(t, expected, got)
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
		assert.Equal(t, v, got)
	}
}

func TestSetFilename(t *testing.T) {
	expected := "test"
	SetFilename(expected)
	assert.Equal(t, expected, config.filename)
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
			"scripts": map[string]string{},
			"timeout": "",
		},
	}

	expected, _ := config.marshal(filepath.Ext(filename))
	config.mp = temp

	err := WriteStruct(filename)
	assert.NoError(t, err, "Failed writing config struct")

	got, err := os.ReadFile(filename)
	assert.NoError(t, err, "Failed reading file")
	os.Remove(filename)

	assert.Equal(t, expected, got)
}

// Remove config file created on user home directory.
func cleanup(t *testing.T) {
	home, err := os.UserHomeDir()
	assert.NoError(t, err)

	os.RemoveAll(filepath.Join(home, ".kure"))
	assert.NoError(t, err)
}
