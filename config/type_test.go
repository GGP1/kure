package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	Reset()
	Set("test", "test")

	cases := []struct {
		expected interface{}
		desc     string
		key      string
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
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestLoad(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		err := Load("./testdata/mock_config.yaml")
		assert.NoError(t, err, "Failed loading file to config")

		str := Get("test.string")
		assert.Equal(t, "test", str)

		num := Get("test.number")
		assert.Equal(t, 1, num)

		bl := Get("test.bool")
		assert.True(t, bl.(bool))
	})

	t.Run("Invalid filename", func(t *testing.T) {
		err := Load("")
		assert.Error(t, err)
	})
}

func TestSet(t *testing.T) {
	cases := []struct {
		value    interface{}
		expected interface{}
		desc     string
		key      string
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
			assert.Equal(t, tc.expected, got)
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

	err := config.Write(filename, os.O_CREATE|os.O_RDWR)
	assert.NoError(t, err, "Failed creating file")
	defer os.Remove(filename)

	content, err := os.ReadFile(filename)
	assert.NoError(t, err, "Failed reading file")

	err = config.populateMap(content, filepath.Ext(filename))
	assert.NoError(t, err, "Failed populating config map")

	got := cast.ToInt(config.mp[key])
	assert.Equal(t, value, got)
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
			err := config.Write(tc.filename, tc.flags)
			assert.Error(t, err)
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
			assert.NoError(t, err, "Failed marshaling data")

			assert.Equal(t, tc.expected, string(got))
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
			_, err := config.marshal(filepath.Ext(tc.path))
			assert.Error(t, err)
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
			assert.NoError(t, err, "Failed marshaling config map")

			err = c.populateMap(data, ext)
			assert.NoError(t, err, "Failed populating map")

			got := cast.ToInt(c.mp["test"])
			assert.Equal(t, value, got)
		})
	}

	t.Run("Invalid type", func(t *testing.T) {
		c := New()
		err := c.populateMap([]byte("invalid"), ".env")
		assert.Error(t, err)
	})
}

func TestInsertAndSearch(t *testing.T) {
	cases := []struct {
		value interface{}
		desc  string
		key   string
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
			assert.Equal(t, tc.value, got)
		})
	}

	t.Run("Nil keys", func(t *testing.T) {
		mp := make(map[string]interface{})
		insert(mp, nil, nil)
		got := search(mp, nil)
		assert.Nil(t, got)
	})
}
