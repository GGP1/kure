package restore

import (
	"testing"

	cmdutil "github.com/GGP1/kure/commands"
	dbutil "github.com/GGP1/kure/db"
	"github.com/GGP1/kure/db/card"
	"github.com/GGP1/kure/db/entry"
	"github.com/GGP1/kure/pb"
)

func TestLogs(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")

	expected := &pb.Entry{
		Name:     "test",
		Username: "test@test.com",
		Password: "Pb*9' fxd%,IS:Zo_1JVw",
		Expires:  "Never",
	}
	if err := entry.Create(db, expected); err != nil {
		t.Fatal(err)
	}

	l, err := newLog(dbutil.EntryBucket)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	logs := []*log{l}
	if err := writeLogs(db, logs); err != nil {
		t.Errorf("Failed writing logs: %v", err)
	}

	if err := readLogs(db, logs); err != nil {
		t.Errorf("Failed reading logs: %v", err)
	}

	got, err := entry.Get(db, expected.Name)
	if err != nil {
		t.Errorf("Failed fetching entry: %v", err)
	}

	if expected.Username != got.Username || expected.Password != got.Password {
		t.Errorf("Invalid entry, expected %v, got %v", expected, got)
	}
}

func TestReadLogs(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")

	expected := &pb.Card{
		Name:         "testRead",
		Number:       "47964212",
		SecurityCode: "442",
	}
	if err := card.Create(db, expected); err != nil {
		t.Fatal(err)
	}

	l, err := newLog(dbutil.CardBucket)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	logs := []*log{l}
	if err := readLogs(db, logs); err != nil {
		t.Errorf("Failed reading logs: %v", err)
	}

	got, err := card.Get(db, expected.Name)
	if err != nil {
		t.Errorf("Failed fetching card: %v", err)
	}

	if expected.Number != got.Number || expected.SecurityCode != got.SecurityCode {
		t.Errorf("Invalid card, expected %v, got %v", expected, got)
	}
}

func TestWriteLogs(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")

	l, err := newLog(dbutil.EntryBucket)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	if err := writeLogs(db, []*log{l}); err != nil {
		t.Errorf("Failed writing logs: %v", err)
	}
}
