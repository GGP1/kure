package edit

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"testing"

	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/card"
	"github.com/GGP1/kure/pb"
	"github.com/spf13/viper"

	bolt "go.etcd.io/bbolt"
)

func TestEditErrors(t *testing.T) {
	db := cmdutil.SetContext(t, "../../../db/testdata/database")
	createCard(t, db, "test")

	cases := []struct {
		desc string
		name string
		it   string
		set  func()
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
				viper.Set("editor", "non-existent")
			},
		},
	}

	cmd := NewCmd(db)

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			tc.set()
			cmd.SetArgs([]string{tc.name})
			cmd.Flags().Set("it", tc.it)

			if err := cmd.Execute(); err == nil {
				t.Error("Expected an error and got nil")
			}
		})
	}
}

func TestCreateTempFile(t *testing.T) {
	card := &pb.Card{Name: "test-create-file"}
	filename, err := createTempFile(card)
	if err != nil {
		t.Fatalf("Failed creating the file: %v", err)
	}
	defer os.Remove(filename)

	content, err := os.ReadFile(filename)
	if err != nil {
		t.Errorf("Failed reading the file: %v", err)
	}

	var got pb.Card
	if err := json.Unmarshal(content, &got); err != nil {
		t.Errorf("Failed decoding the file: %v", err)
	}

	if !reflect.DeepEqual(card, &got) {
		t.Error("Expected cards to be deep equal")
	}
}

func TestReadTmpFile(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		if _, err := readTmpFile("testdata/test_read.json"); err != nil {
			t.Error(err)
		}
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
				if _, err := readTmpFile(tc.filename); err == nil {
					t.Error("Expected an error and got nil")
				}
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

	if err := updateCard(db, name, newCard); err != nil {
		t.Error(err)
	}

	c, err := card.Get(db, newName)
	if err != nil {
		t.Error(err)
	}

	if reflect.DeepEqual(c, newCard) {
		t.Errorf("Expected %v, got %v", newCard, c)
	}

	t.Run("Invalid name", func(t *testing.T) {
		newCard.Name = ""
		if err := updateCard(db, "fail", newCard); err == nil {
			t.Error("Expected an error and got nil")
		}
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

	buf := bytes.NewBufferString("\nCredit\n\n-\n\n-!q\n")

	if err := useStdin(db, buf, oldCard); err != nil {
		t.Fatal(err)
	}

	got, err := card.Get(db, "test")
	if err != nil {
		t.Error(err)
	}

	if got.Type != "Credit" {
		t.Errorf("Expected Credit, got %q", got.Type)
	}
	if got.Number != oldCard.Number {
		t.Errorf("Expected %s, got %q", oldCard.Number, got.Number)
	}
	if got.SecurityCode != "" {
		t.Errorf("Expected an empty string, got %q", got.SecurityCode)
	}
	if got.ExpireDate != oldCard.ExpireDate {
		t.Errorf("Expected %s, got %q", oldCard.ExpireDate, got.ExpireDate)
	}
	if got.Notes != "" {
		t.Errorf("Expected an empty string, got %q", got.Notes)
	}
}

func TestPostRun(t *testing.T) {
	NewCmd(nil).PostRun(nil, nil)
}

func createCard(t *testing.T, db *bolt.DB, name string) {
	t.Helper()
	if err := card.Create(db, &pb.Card{Name: name}); err != nil {
		t.Fatalf("Failed creating the card: %v", err)
	}
}
