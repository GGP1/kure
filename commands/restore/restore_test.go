package restore

import (
	"testing"

	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/bucket"
	"github.com/GGP1/kure/db/card"
	"github.com/GGP1/kure/db/entry"
	"github.com/GGP1/kure/pb"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestLogs(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")

	expected := &pb.Entry{
		Name:     "test",
		Username: "test@test.com",
		Password: "Pb*9' fxd%,IS:Zo_1JVw",
		Expires:  "Never",
	}
	err := entry.Create(db, expected)
	assert.NoError(t, err)

	l, err := newLog(bucket.Entry.GetName())
	assert.NoError(t, err)
	defer l.Close()

	logs := []*log{l}
	err = writeLogs(db, logs)
	assert.NoError(t, err, "Failed writing logs")

	err = readLogs(db, logs)
	assert.NoError(t, err, "Failed reading logs")

	got, err := entry.Get(db, expected.Name)
	assert.NoError(t, err, "Failed fetching entry")

	equal := proto.Equal(expected, got)
	assert.True(t, equal)
}

func TestReadLogs(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")

	expected := &pb.Card{
		Name:         "testRead",
		Number:       "47964212",
		SecurityCode: "442",
	}
	err := card.Create(db, expected)
	assert.NoError(t, err)

	l, err := newLog(bucket.Card.GetName())
	assert.NoError(t, err)
	defer l.Close()

	logs := []*log{l}
	err = readLogs(db, logs)
	assert.NoError(t, err, "Failed reading logs")

	got, err := card.Get(db, expected.Name)
	assert.NoError(t, err, "Failed fetching card")

	equal := proto.Equal(expected, got)
	assert.True(t, equal)
}

func TestWriteLogs(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")

	l, err := newLog(bucket.Entry.GetName())
	assert.NoError(t, err)
	defer l.Close()

	err = writeLogs(db, []*log{l})
	assert.NoError(t, err, "Failed writing logs")
}
