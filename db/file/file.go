package file

import (
	"bytes"
	"compress/gzip"
	"io"
	"strings"

	"github.com/GGP1/kure/crypt"
	dbutil "github.com/GGP1/kure/db"
	"github.com/GGP1/kure/pb"

	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"
)

// Create a new file compressing its content.
func Create(db *bolt.DB, file *pb.File) error {
	// Ensure the name does not contain null characters
	if strings.ContainsRune(file.Name, '\x00') {
		return errors.New("file name contains null characters")
	}

	compressedContent, err := compress(file.Content)
	if err != nil {
		return err
	}
	file.Content = compressedContent

	return db.Batch(func(tx *bolt.Tx) error {
		b := tx.Bucket(dbutil.FileBucket)

		return save(b, file)
	})
}

// Get retrieves the file with the specified name.
func Get(db *bolt.DB, name string) (*pb.File, error) {
	file := &pb.File{}

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(dbutil.FileBucket)

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

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(dbutil.FileBucket)

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

	b := tx.Bucket(dbutil.FileBucket)
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
	return dbutil.ListNames(db, dbutil.FileBucket)
}

// Remove removes a file from the database.
func Remove(db *bolt.DB, name string) error {
	return dbutil.Remove(db, dbutil.FileBucket, name)
}

// Rename recreates a file with a new key and deletes the old one.
func Rename(db *bolt.DB, oldName, newName string) error {
	// Ensure the name does not contain null characters
	if strings.ContainsRune(newName, '\x00') {
		return errors.New("new name contains null characters")
	}

	return db.Batch(func(tx *bolt.Tx) error {
		b := tx.Bucket(dbutil.FileBucket)

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

		if err := save(b, file); err != nil {
			return err
		}

		return b.Delete([]byte(oldName))
	})
}

func compress(content []byte) ([]byte, error) {
	var gzipBuf bytes.Buffer
	gw := gzip.NewWriter(&gzipBuf)

	if _, err := gw.Write(content); err != nil {
		return nil, errors.Wrap(err, "compressing content")
	}

	if err := gw.Close(); err != nil {
		return nil, errors.Wrap(err, "closing gzip writer")
	}

	return gzipBuf.Bytes(), nil
}

func decompress(content []byte) ([]byte, error) {
	compressed := bytes.NewBuffer(content)
	gr, err := gzip.NewReader(compressed)
	if err != nil {
		return nil, errors.Wrap(err, "decompressing content")
	}
	defer gr.Close()

	var decompressed bytes.Buffer
	if _, err = io.Copy(&decompressed, gr); err != nil {
		return nil, errors.Wrap(err, "copying decompressed content")
	}

	return decompressed.Bytes(), nil
}

func save(b *bolt.Bucket, file *pb.File) error {
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
}
