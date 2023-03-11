package file

import (
	"bytes"
	"compress/gzip"
	"io"

	"github.com/GGP1/kure/crypt"
	dbutil "github.com/GGP1/kure/db"
	"github.com/GGP1/kure/db/bucket"
	"github.com/GGP1/kure/pb"

	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"
)

// Create a new file with its content compressed.
func Create(db *bolt.DB, file *pb.File) error {
	compressedContent, err := compress(file.Content)
	if err != nil {
		return err
	}
	file.Content = compressedContent

	return db.Update(func(tx *bolt.Tx) error {
		return dbutil.Put(tx, file)
	})
}

// Get retrieves the file with the specified name.
func Get(db *bolt.DB, name string) (*pb.File, error) {
	file := &pb.File{}
	if err := dbutil.Get(db, name, file); err != nil {
		return nil, err
	}

	decompressedContent, err := decompress(file.Content)
	if err != nil {
		return nil, err
	}
	file.Content = decompressedContent

	return file, nil
}

// GetCheap is like Get but without getting the file content.
func GetCheap(db *bolt.DB, name string) (*pb.FileCheap, error) {
	file := &pb.FileCheap{}
	if err := dbutil.Get(db, name, file); err != nil {
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

	b := tx.Bucket(bucket.File.GetName())
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

		decompressedContent, err := decompress(file.Content)
		if err != nil {
			return err
		}
		file.Content = decompressedContent
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
	return dbutil.ListNames(db, bucket.FileNames.GetName())
}

// Remove removes one or more files from the database.
func Remove(db *bolt.DB, names ...string) error {
	return db.Update(func(tx *bolt.Tx) error {
		return dbutil.Remove(tx, &pb.File{}, names...)
	})
}

// Rename recreates a file with a new key and deletes the old one.
func Rename(db *bolt.DB, oldName, newName string) error {
	return db.Update(func(tx *bolt.Tx) error {
		file, err := Get(db, oldName)
		if err != nil {
			return err
		}

		compressedContent, err := compress(file.Content)
		if err != nil {
			return err
		}
		file.Name = newName
		file.Content = compressedContent

		if err := dbutil.Put(tx, file); err != nil {
			return err
		}

		return dbutil.Remove(tx, &pb.File{}, oldName)
	})
}

func compress(content []byte) ([]byte, error) {
	var gzipBuf bytes.Buffer
	gw := gzip.NewWriter(&gzipBuf)

	if _, err := gw.Write(content); err != nil {
		return nil, errors.Wrap(err, "compress content")
	}

	if err := gw.Close(); err != nil {
		return nil, errors.Wrap(err, "close gzip writer")
	}

	return gzipBuf.Bytes(), nil
}

func decompress(content []byte) ([]byte, error) {
	compressed := bytes.NewBuffer(content)
	gr, err := gzip.NewReader(compressed)
	if err != nil {
		return nil, errors.Wrap(err, "decompress content")
	}
	defer gr.Close()

	var decompressed bytes.Buffer
	if _, err = io.Copy(&decompressed, gr); err != nil {
		return nil, errors.Wrap(err, "copy decompressed content")
	}

	return decompressed.Bytes(), nil
}
