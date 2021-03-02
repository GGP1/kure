package add

import (
	"bytes"
	"reflect"
	"testing"

	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/entry"
	"github.com/GGP1/kure/pb"
)

func TestAdd(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")

	cases := []struct {
		desc   string
		name   string
		length string
		format string
	}{
		{
			desc:   "Add",
			name:   "test",
			length: "18",
			format: "1,2,3,4,5",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			buf := bytes.NewBufferString("username\nurl\n03/05/2024\nnotes<")
			cmd := NewCmd(db, buf)
			cmd.SetArgs([]string{tc.name})

			f := cmd.Flags()
			f.Set("length", tc.length)
			f.Set("format", tc.format)

			if err := cmd.Execute(); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestAddErrors(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")

	if err := entry.Create(db, &pb.Entry{Name: "test"}); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		desc   string
		name   string
		length string
		levels string
	}{
		{
			desc: "Invalid name",
			name: "",
		},
		{
			desc:   "Invalid length",
			name:   "test-invalid-length",
			length: "0",
		},
		{
			desc:   "Already exists",
			name:   "test",
			length: "10",
		},
		{
			desc:   "Invalid levels",
			name:   "test-levels",
			length: "6",
			levels: "1,2,3,7",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			buf := bytes.NewBufferString("username\nurl\n03/05/2024\nnotes<")
			cmd := NewCmd(db, buf)
			cmd.SetArgs([]string{tc.name})
			f := cmd.Flags()
			f.Set("length", tc.length)
			f.Set("levels", tc.levels)

			if err := cmd.Execute(); err == nil {
				t.Error("Expected an error and got nil")
			}
		})
	}
}

func TestGenPassword(t *testing.T) {
	cases := []struct {
		desc string
		opts *addOptions
	}{
		{
			desc: "Generate password",
			opts: &addOptions{
				length: 10,
				levels: []int{1, 2, 3},
				repeat: false,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			if _, err := genPassword(tc.opts); err != nil {
				t.Errorf("Failed generating password: %v", err)
			}
		})
	}
}

func TestGenPasswordErrors(t *testing.T) {
	cases := []struct {
		desc string
		opts *addOptions
	}{
		{
			desc: "Invalid length",
			opts: &addOptions{
				length: 0,
				levels: []int{1, 2, 3},
			},
		},
		{
			desc: "Include and exclude error",
			opts: &addOptions{
				length:  5,
				include: "a",
				exclude: "a",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			if _, err := genPassword(tc.opts); err == nil {
				t.Error("Expected an error and got nil")
			}
		})
	}
}

func TestEntryInput(t *testing.T) {
	expected := &pb.Entry{
		Name:     "test",
		Username: "username",
		URL:      "url",
		Notes:    "notes",
		Expires:  "Fri, 03 May 2024 00:00:00 +0000",
	}

	buf := bytes.NewBufferString("username\nurl\n03/05/2024\nnotes<")

	got, err := entryInput(buf, "test", false)
	if err != nil {
		t.Fatalf("Failed creating entry: %v", err)
	}

	// As it's randomly generated, use the same one
	expected.Password = got.Password
	if !reflect.DeepEqual(expected, got) {
		t.Errorf("Expected %v, got %v", expected, got)
	}
}

func TestInvalidExpirationTime(t *testing.T) {
	buf := bytes.NewBufferString("username\nurl\nnotes\ninvalid<\n")

	if _, err := entryInput(buf, "test", false); err == nil {
		t.Error("Expected an error and got nil")
	}
}

func TestPostRun(t *testing.T) {
	NewCmd(nil, nil).PostRun(nil, nil)
}
