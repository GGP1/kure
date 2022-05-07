package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cast"
)

func TestGet(t *testing.T) {
	Reset()
	Set("test", "test")

	cases := []struct {
		desc     string
		key      string
		expected interface{}
	}{
		{
			desc:     "Non nil",
			key:      "test",
			expected: "test",
		},
		{
			desc:     "Nil",
			key:      "",
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			got := Get(tc.key)
			if got != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, got)
			}
		})
	}
}

func TestLoad(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		if err := Load("./testdata/mock_config.yaml"); err != nil {
			t.Errorf("Failed loading file to config: %v", err)
		}

		str := Get("test.string")
		expStr := "test"
		if str != expStr {
			t.Errorf("Expected %q, got %q", expStr, str)
		}

		num := Get("test.number")
		expNum := 1
		if num != expNum {
			t.Errorf("Expected %d, got %d", expNum, num)
		}

		bl := Get("test.bool")
		expBool := true
		if bl != expBool {
			t.Errorf("Expected %v, got %v", expBool, bl)
		}
	})

	t.Run("Invalid filename", func(t *testing.T) {
		if err := Load(""); err == nil {
			t.Error("Expected an error and got nil")
		}
	})
}

func TestSet(t *testing.T) {
	cases := []struct {
		desc     string
		key      string
		value    interface{}
		expected interface{}
	}{
		{
			desc:     "Empty key",
			key:      "",
			value:    "nothing",
			expected: nil,
		},
		{
			desc:     "1st level",
			key:      "test",
			value:    "test1",
			expected: "test1",
		},
		{
			desc:     "2nd level",
			key:      "test.second",
			value:    "test2",
			expected: "test2",
		},
		{
			desc:     "3rd level",
			key:      "test.third.level",
			value:    "test3",
			expected: "test3",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			Set(tc.key, tc.value)

			got := Get(tc.key)
			if got != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, got)
			}
		})
	}
}

func TestWrite(t *testing.T) {
	filename := "test.json"
	key := "one"
	value := 1
	config.mp = map[string]interface{}{
		key: value,
	}

	if err := config.Write(filename, os.O_CREATE|os.O_RDWR); err != nil {
		t.Errorf("Failed creating file: %v", err)
	}
	defer os.Remove(filename)

	content, err := os.ReadFile(filename)
	if err != nil {
		t.Errorf("Failed reading file: %v", err)
	}

	if err := config.populateMap(content, filepath.Ext(filename)); err != nil {
		t.Errorf("Failed populating config map: %v", err)
	}

	got := cast.ToInt(config.mp[key])
	if got != value {
		t.Errorf("Expected %d, got %d", value, got)
	}
}

func TestWriteErrors(t *testing.T) {
	fName := "test.json"
	f, _ := os.Create(fName)
	f.Close()
	defer os.Remove(fName)

	cases := []struct {
		desc     string
		filename string
		flags    int
	}{
		{
			desc:     "Invalid type",
			filename: "test.hcl",
		},
		{
			desc:     "Already exists",
			filename: fName,
			flags:    os.O_CREATE | os.O_EXCL | os.O_RDWR,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			if err := config.Write(tc.filename, tc.flags); err == nil {
				t.Error("Expected an error and got nil")
			}
		})
	}
}

func TestMarshaler(t *testing.T) {
	Reset()
	Set("test", "Go")

	cases := []struct {
		desc     string
		path     string
		expected string
	}{
		{
			desc: "Marshal to JSON",
			path: "test.json",
			expected: `{
   "test": "Go"
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
			expected: "test = 'Go'\n",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := config.marshal(filepath.Ext(tc.path))
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
			if _, err := config.marshal(filepath.Ext(tc.path)); err == nil {
				t.Error("Expected an error and got nil")
			}
		})
	}
}

func TestPopulateMap(t *testing.T) {
	cases := []struct {
		desc string
	}{
		{desc: "json"},
		{desc: "toml"},
		{desc: "yaml"},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			ext := "." + tc.desc
			c := New()
			value := 100
			insert(c.mp, []string{"test"}, value)

			data, err := c.marshal(ext)
			if err != nil {
				t.Errorf("Failed marshaling config map: %v", err)
			}

			if err := c.populateMap(data, ext); err != nil {
				t.Errorf("Failed populating map: %v", err)
			}

			got := cast.ToInt(c.mp["test"])
			if got != value {
				t.Errorf("Expected %d, got %d", value, got)
			}
		})
	}

	t.Run("Invalid type", func(t *testing.T) {
		c := New()
		if err := c.populateMap([]byte("invalid"), ".env"); err == nil {
			t.Error("Expected file type error and got nil")
		}
	})
}

func TestInsertAndSearch(t *testing.T) {
	cases := []struct {
		desc  string
		key   string
		value interface{}
	}{
		{
			desc:  "No value",
			key:   "",
			value: nil,
		},
		{
			desc:  "1st level",
			key:   "test",
			value: "test1",
		},
		{
			desc:  "2nd level",
			key:   "test.second",
			value: "test2",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			mp := make(map[string]interface{})
			path := strings.Split(tc.key, ".")
			insert(mp, path, tc.value)

			got := search(mp, path)
			if got != tc.value {
				t.Errorf("Expected %v, got %v", tc.value, got)
			}
		})
	}

	t.Run("Nil keys", func(t *testing.T) {
		mp := make(map[string]interface{})
		insert(mp, nil, nil)
		got := search(mp, nil)
		if got != nil {
			t.Errorf("Expected nil and got %v", got)
		}
	})
}
