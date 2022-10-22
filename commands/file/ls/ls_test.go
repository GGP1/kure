package ls

import (
	"testing"
	"time"

	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/file"
	"github.com/GGP1/kure/pb"

	"github.com/stretchr/testify/assert"
)

func TestLs(t *testing.T) {
	db := cmdutil.SetContext(t, "../../../db/testdata/database")

	err := file.Create(db, &pb.File{Name: "test.txt"})
	assert.NoError(t, err, "Failed creating file")

	cases := []struct {
		desc   string
		name   string
		filter string
	}{
		{
			desc: "List one",
			name: "test.txt",
		},
		{
			desc:   "Filter by name",
			name:   "test*",
			filter: "true",
		},
		{
			desc: "List all",
			name: "",
		},
	}

	cmd := NewCmd(db)

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			f := cmd.Flags()
			cmd.SetArgs([]string{tc.name})
			f.Set("filter", tc.filter)

			err := cmd.Execute()
			assert.NoError(t, err)
		})
	}
}

func TestLsErrors(t *testing.T) {
	db := cmdutil.SetContext(t, "../../../db/testdata/database")

	err := file.Create(db, &pb.File{Name: "test.txt"})
	assert.NoError(t, err, "Failed creating file")

	cases := []struct {
		desc   string
		name   string
		filter string
	}{
		{
			desc: "File does not exist",
			name: "non-existent",
		},
		{
			desc:   "No files found",
			name:   "non-existent",
			filter: "true",
		},
		{
			desc:   "Filter syntax error",
			name:   "[error",
			filter: "true",
		},
	}

	cmd := NewCmd(db)

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			f := cmd.Flags()
			cmd.SetArgs([]string{tc.name})
			f.Set("filter", tc.filter)

			err := cmd.Execute()
			assert.Error(t, err)
		})
	}
}

func TestPrintFile(t *testing.T) {
	cases := []struct {
		desc string
		size int64
	}{
		{
			desc: "Bytes",
			size: 100,
		},
		{
			desc: "Kilo bytes",
			size: KB,
		},
		{
			desc: "Mega bytes",
			size: MB,
		},
		{
			desc: "Giga bytes",
			size: GB,
		},
		{
			desc: "Tera bytes",
			size: TB,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			printFile(&pb.FileCheap{Size: tc.size})
		})
	}
}

func TestPrintFileUpdatedAt(t *testing.T) {
	cases := []struct {
		desc string
		time int64
	}{
		{
			desc: "Without updated at",
			time: time.Time{}.Unix(),
		},
		{
			desc: "With updated at",
			time: 100,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			printFile(&pb.FileCheap{UpdatedAt: tc.time})
		})
	}
}

func TestPostRun(t *testing.T) {
	NewCmd(nil).PostRun(nil, nil)
}
