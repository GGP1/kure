package file

import (
	"bytes"
	"compress/gzip"
	"io"
	"strings"

	"github.com/GGP1/kure/crypt"
	dbutils "github.com/GGP1/kure/db"
	"github.com/GGP1/kure/pb"

	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"
)

var fileBucket = []byte("kure_file")

// Create a new file compressing its content.
func Create(db *bolt.DB, file *pb.File) error {
	// Ensure the name does not contain null characters
	if strings.ContainsRune(file.Name, '\x00') {
		return errors.New("file name contains null characters")
	}

	// Compress content
	var gzipBuf bytes.Buffer
	gw := gzip.NewWriter(&gzipBuf)

	if _, err := gw.Write(file.Content); err != nil {
		return errors.Wrap(err, "compressing file")
	}

	if err := gw.Close(); err != nil {
		return errors.Wrap(err, "closing gzip writer")
	}

	file.Content = gzipBuf.Bytes()

	return db.Batch(func(tx *bolt.Tx) error {
		b := tx.Bucket(fileBucket)

		buf, err := proto.Marshal(file)
		if err != nil {
			return errors.Wrap(err, "marshal file")
		}

		encFile, err := crypt.Encrypt(buf)
		if err != nil {
			return errors.Wrap(err, "encrypt file")
		}

		if err := b.Put([]byte(file.Name), encFile); err != nil {
			return errors.Wrap(err, "save file")
		}

		return nil
	})
}

// Get retrieves the file with the specified name.
func Get(db *bolt.DB, name string) (*pb.File, error) {
	file := &pb.File{}

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(fileBucket)

		encFile := b.Get([]byte(name))
		if encFile == nil {
			return errors.Errorf("file %q does not exist", name)
		}

		decFile, err := crypt.Decrypt(encFile)
		if err != nil {
			return errors.Wrap(err, "decrypt file")
		}

		if err := proto.Unmarshal(decFile, file); err != nil {
			return errors.Wrap(err, "unmarshal file")
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Decompress
	compressed := bytes.NewBuffer(file.Content)
	gr, err := gzip.NewReader(compressed)
	if err != nil {
		return nil, errors.Wrap(err, "decompressing file")
	}
	defer gr.Close()

	var decompressed bytes.Buffer
	if _, err = io.Copy(&decompressed, gr); err != nil {
		return nil, errors.Wrap(err, "copying decompressed file")
	}

	file.Content = decompressed.Bytes()

	return file, nil
}

// GetCheap is like Get but without getting the file content.
func GetCheap(db *bolt.DB, name string) (*pb.FileCheap, error) {
	file := &pb.FileCheap{}

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(fileBucket)

		encFile := b.Get([]byte(name))
		if encFile == nil {
			return errors.Errorf("file %q does not exist", name)
		}

		decFile, err := crypt.Decrypt(encFile)
		if err != nil {
			return errors.Wrap(err, "decrypt file")
		}

		if err := proto.Unmarshal(decFile, file); err != nil {
			return errors.Wrap(err, "unmarshal file")
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return file, nil
}

// List returns a slice with all the files stored in the file bucket.
func List(db *bolt.DB) ([]*pb.File, error) {
	tx, err := db.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	b := tx.Bucket(fileBucket)
	files := make([]*pb.File, 0, b.Stats().KeyN)

	err = b.ForEach(func(k, v []byte) error {
		file := &pb.File{}

		decFile, err := crypt.Decrypt(v)
		if err != nil {
			return errors.Wrap(err, "decrypt file")
		}

		if err := proto.Unmarshal(decFile, file); err != nil {
			return errors.Wrap(err, "unmarshal file")
		}

		// Decompress
		compressed := bytes.NewBuffer(file.Content)
		gr, err := gzip.NewReader(compressed)
		if err != nil {
			return errors.Wrap(err, "decompressing file")
		}

		var decompressed bytes.Buffer
		if _, err = io.Copy(&decompressed, gr); err != nil {
			return errors.Wrap(err, "copying decompressed file")
		}
		gr.Close()

		file.Content = decompressed.Bytes()
		files = append(files, file)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}

// ListNames returns a slice with all the files names.
func ListNames(db *bolt.DB) ([]string, error) {
	return dbutils.ListNames(db, fileBucket)
}

// Remove removes a file from the database.
func Remove(db *bolt.DB, name string) error {
	return dbutils.Remove(db, fileBucket, name)
}
