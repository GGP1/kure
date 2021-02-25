package add

import (
	"bytes"
	"os"
	"reflect"
	"testing"

	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/card"
	"github.com/GGP1/kure/pb"
)

func TestAdd(t *testing.T) {
	db := cmdutil.SetContext(t, "../../../db/testdata/database")

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

	cmd := NewCmd(db, os.Stdin)

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			cmd.SetArgs([]string{tc.name})
			if err := cmd.Execute(); err != nil {
				t.Error(err)
			}

			if _, err := card.Get(db, tc.name); err != nil {
				t.Errorf("Card wasn't created correctly: %v", err)
			}
		})
	}
}

func TestAddErrors(t *testing.T) {
	db := cmdutil.SetContext(t, "../../../db/testdata/database")

	if err := card.Create(db, &pb.Card{Name: "test"}); err != nil {
		t.Fatalf("Failed creating the card: %v", err)
	}

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

	cmd := NewCmd(db, os.Stdin)

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			cmd.SetArgs([]string{tc.name})
			if err := cmd.Execute(); err == nil {
				t.Error("Expected an error and got nil")
			}
		})
	}
}

func TestInput(t *testing.T) {
	db := cmdutil.SetContext(t, "../../../db/testdata/database")

	expected := &pb.Card{
		Name:         "test",
		Type:         "type",
		Number:       "123456789",
		SecurityCode: "1234",
		ExpireDate:   "2021/06",
		Notes:        "notes",
	}

	buf := bytes.NewBufferString("type\n123456789\n1234\n2021/06\nnotes")

	got, err := input(db, "test", buf)
	if err != nil {
		t.Fatalf("Failed creating the card: %v", err)
	}

	if !reflect.DeepEqual(expected, got) {
		t.Errorf("Expected %v, got %v", expected, got)
	}
}
