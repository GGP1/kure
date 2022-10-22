package restore

import (
	"os"
	"testing"

	"github.com/GGP1/kure/pb"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestLog(t *testing.T) {
	expected := [][]byte{
		[]byte("first"),
		[]byte("second"),
		[]byte("third"),
		[]byte("fourth"),
		[]byte("fifth"),
	}
	bucketName := []byte("test")

	log, err := newLog(bucketName)
	assert.NoError(t, err)
	defer log.Close()

	for _, e := range expected {
		err := log.Write(e)
		assert.NoError(t, err)
	}

	err = log.Sync()
	assert.NoError(t, err)

	got, err := log.Read()
	assert.NoError(t, err)

	for i, g := range got {
		e := expected[i]
		assert.Equal(t, e, g)
	}

	bName := log.BucketName()
	assert.Equal(t, bucketName, bName)

	err = log.Close()
	assert.NoError(t, err)

	_, err = os.Stat(log.file.Name())
	assert.Error(t, err, "The file wasn't erased correctly")
}

func TestErrClosed(t *testing.T) {
	log, err := newLog([]byte("test"))
	assert.NoError(t, err)

	err = log.Close()
	assert.NoError(t, err, "Failed closing file")

	_, err = log.Read()
	assert.Error(t, err)

	err = log.Write([]byte("data"))
	assert.Error(t, err)
}

func BenchmarkRead(b *testing.B) {
	l, err := newLog([]byte("benchmark"))
	assert.NoError(b, err)
	b.Cleanup(func() {
		l.Close()
	})

	encEntry := createEncodedEntry(b)
	err = l.Write(encEntry)
	assert.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Read()
	}
}

func BenchmarkWrite(b *testing.B) {
	l, err := newLog([]byte("benchmark"))
	assert.NoError(b, err)
	b.Cleanup(func() {
		l.Close()
	})

	encEntry := createEncodedEntry(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Write(encEntry)
	}
}

func createEncodedEntry(b *testing.B) []byte {
	b.Helper()
	entry := &pb.Entry{
		Name:     "benchmark",
		Username: "benchMark",
		Password: "@RyL8B0'/h{ .xpG9ZD!/itw7",
		URL:      "https://benchmarkread.com",
		Expires:  "Never",
		Notes:    "9Ns1MfBEzLY7JImZj3Xka0GpcPo846nlD5ATegRt",
	}

	encEntry, err := proto.Marshal(entry)
	assert.NoError(b, err)

	return encEntry
}
