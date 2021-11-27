package restore

import (
	"bytes"
	"os"
	"testing"

	"github.com/GGP1/kure/pb"

	"github.com/golang/protobuf/proto"
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
	if err != nil {
		t.Fatal(err)
	}
	defer log.Close()

	for _, e := range expected {
		if err := log.Write(e); err != nil {
			t.Fatal(err)
		}
	}

	if err := log.Sync(); err != nil {
		t.Fatal(err)
	}

	got, err := log.Read()
	if err != nil {
		t.Fatal(err)
	}

	for i, g := range got {
		e := expected[i]
		if !bytes.Equal(g, e) {
			t.Errorf("Record [%v] is corrupted, expected: %s, got %s", i, e, g)
		}
	}

	bName := log.BucketName()
	if !bytes.Equal(bName, bucketName) {
		t.Errorf("Invalid bucket name, expected: %s, got %s", bucketName, bName)
	}

	if err := log.Close(); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(log.file.Name()); err == nil {
		t.Fatal("The file wasn't erased correctly")
	}
}

func TestErrClosed(t *testing.T) {
	log, err := newLog([]byte("test"))
	if err != nil {
		t.Fatal(err)
	}

	if err := log.Close(); err != nil {
		t.Errorf("Failed closing file: %v", err)
	}

	if _, err := log.Read(); err == nil {
		t.Error("Expected an error when reading and got nil")
	}

	if err := log.Write([]byte("data")); err == nil {
		t.Error("Expected an error when writing and got nil")
	}
}

func BenchmarkRead(b *testing.B) {
	l, err := newLog([]byte("benchmark"))
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() {
		l.Close()
	})

	encEntry := createEncodedEntry(b)
	if err := l.Write(encEntry); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Read()
	}
}

func BenchmarkWrite(b *testing.B) {
	l, err := newLog([]byte("benchmark"))
	if err != nil {
		b.Fatal(err)
	}
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
	if err != nil {
		b.Fatal(err)
	}

	return encEntry
}
