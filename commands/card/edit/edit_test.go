package edit

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"testing"

	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/config"
	"github.com/GGP1/kure/db/card"
	"github.com/GGP1/kure/pb"

	"github.com/stretchr/testify/assert"
	bolt "go.etcd.io/bbolt"
)

func TestEditErrors(t *testing.T) {
	db := cmdutil.SetContext(t, "../../../db/testdata/database")
	createCard(t, db, "test")

	cases := []struct {
		set  func()
		desc string
		name string
		it   string
	}{
		{
			desc: "Invalid name",
			name: "",
			set:  func() {},
		},
		{
			desc: "Non-existent entry",
			name: "non-existent",
			it:   "true",
			set: func() {
				config.Set("editor", "non-existent")
			},
		},
	}

	cmd := NewCmd(db)

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			tc.set()
			cmd.SetArgs([]string{tc.name})
			cmd.Flags().Set("it", tc.it)

			err := cmd.Execute()
			assert.Error(t, err)
		})
	}
}

func TestCreateTempFile(t *testing.T) {
	card := &pb.Card{Name: "test-create-file"}
	filename, err := createTempFile(card)
	assert.NoError(t, err, "Failed creating the file")
	defer os.Remove(filename)

	content, err := os.ReadFile(filename)
	assert.NoError(t, err, "Failed reading the file")

	var got pb.Card
	err = json.Unmarshal(content, &got)
	assert.NoError(t, err, "Failed reading the file")

	if !reflect.DeepEqual(card, &got) {
		t.Error("Expected cards to be deep equal")
	}
}

func TestReadTmpFile(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		_, err := readTmpFile("testdata/test_read.json")
		assert.NoError(t, err)
	})

	t.Run("Errors", func(t *testing.T) {
		dir, _ := os.Getwd()
		os.Chdir("testdata")
		// Go back to the initial directory
		defer os.Chdir(dir)

		cases := []struct {
			desc     string
			filename string
		}{
			{
				desc:     "Does not exists",
				filename: "does_not_exists.json",
			},
			{
				desc:     "EOF",
				filename: "test_read_EOF.json",
			},
		}

		for _, tc := range cases {
			t.Run(tc.desc, func(t *testing.T) {
				_, err := readTmpFile(tc.filename)
				assert.Error(t, err)
			})
		}
	})
}

func TestUpdateCard(t *testing.T) {
	db := cmdutil.SetContext(t, "../../../db/testdata/database")
	name := "test_update"
	createCard(t, db, name)

	newName := "new_name"
	newCard := &pb.Card{
		Name:         newName,
		Type:         "",
		Number:       "12345678",
		SecurityCode: "1234",
		ExpireDate:   "12/2023",
		Notes:        "",
	}

	err := updateCard(db, name, newCard)
	assert.NoError(t, err)

	c, err := card.Get(db, newName)
	assert.NoError(t, err)

	assert.NotEqual(t, newCard, c)

	t.Run("Invalid name", func(t *testing.T) {
		newCard.Name = ""
		err := updateCard(db, "fail", newCard)
		assert.Error(t, err)
	})
}

func TestUseStdin(t *testing.T) {
	db := cmdutil.SetContext(t, "../../../db/testdata/database")

	oldCard := &pb.Card{
		Name:         "test",
		Type:         "",
		Number:       "12345678",
		SecurityCode: "1234",
		ExpireDate:   "06/2023",
		Notes:        "test\nnotes",
	}

	buf := bytes.NewBufferString("\nCredit\n\n-\n\n-<\n")

	err := useStdin(db, buf, oldCard)
	assert.NoError(t, err)

	got, err := card.Get(db, "test")
	assert.NoError(t, err)

	assert.Equal(t, "Credit", got.Type)
	assert.Equal(t, oldCard.Number, got.Number)
	assert.Empty(t, got.SecurityCode)
	assert.Equal(t, oldCard.ExpireDate, got.ExpireDate)
	assert.Empty(t, got.Notes)
}

func TestPostRun(t *testing.T) {
	NewCmd(nil).PostRun(nil, nil)
}

func createCard(t *testing.T, db *bolt.DB, name string) {
	t.Helper()
	err := card.Create(db, &pb.Card{Name: name})
	assert.NoError(t, err, "Failed creating the card")
}
