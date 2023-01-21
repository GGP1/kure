package add

import (
	"bytes"
	"testing"

	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/card"
	"github.com/GGP1/kure/pb"

	"github.com/stretchr/testify/assert"
)

func TestAdd(t *testing.T) {
	db := cmdutil.SetContext(t)

	cases := []struct {
		desc string
		name string
	}{
		{
			desc: "Add",
			name: "test",
		},
		{
			desc: "Add2",
			name: "test2",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			buf := bytes.NewBufferString("type\n123456789\n1234\n2021/06\nnotes<\n")
			cmd := NewCmd(db, buf)
			cmd.SetArgs([]string{tc.name})

			err := cmd.Execute()
			assert.NoError(t, err)

			_, err = card.Get(db, tc.name)
			assert.NoError(t, err, "Card wasn't created correctly")
		})
	}
}

func TestAddErrors(t *testing.T) {
	db := cmdutil.SetContext(t)

	err := card.Create(db, &pb.Card{Name: "test"})
	assert.NoError(t, err)

	cases := []struct {
		desc string
		name string
	}{
		{
			desc: "Already exists",
			name: "test",
		},
		{
			desc: "Invalid name",
			name: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			buf := bytes.NewBufferString("type\n123456789\n1234\n2021/06\nnotes<\n")
			cmd := NewCmd(db, buf)
			cmd.SetArgs([]string{tc.name})

			err := cmd.Execute()
			assert.Error(t, err)
		})
	}
}

func TestInput(t *testing.T) {
	db := cmdutil.SetContext(t)

	expected := &pb.Card{
		Name:         "test",
		Type:         "type",
		Number:       "123456789",
		SecurityCode: "1234",
		ExpireDate:   "2021/06",
		Notes:        "notes",
	}

	buf := bytes.NewBufferString("type\n123456789\n1234\n2021/06\nnotes<")

	got, err := input(db, "test", buf)
	assert.NoError(t, err, "Failed creating the card")

	assert.Equal(t, expected, got)
}
