package restore

import (
	"bytes"
	"os"
	"testing"
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
