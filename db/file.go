package db

import (
	"strings"

	"github.com/GGP1/kure/crypt"
	"github.com/GGP1/kure/model/file"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
)

// CreateFile saves a new file into the database.
func CreateFile(file *file.File) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(fileBucket)

		exists := b.Get([]byte(file.Name))
		if exists != nil {
			return errors.New("already exists a file with this name")
		}

		buf, err := proto.Marshal(file)
		if err != nil {
			return errors.Wrap(err, "marshal file")
		}

		encImg, err := crypt.Encrypt(buf)
		if err != nil {
			return errors.Wrap(err, "encrypt file")
		}

		if err := b.Put([]byte(file.Name), encImg); err != nil {
			return errors.Wrap(err, "save file")
		}

		return nil
	})
}

// DeleteFile removes a file from the database.
func DeleteFile(name string) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(fileBucket)
		n := strings.ToLower(name)

		if err := b.Delete([]byte(n)); err != nil {
			return errors.Wrap(err, "delete file")
		}

		return nil
	})
}

// GetFile retrieves the file with the specified name.
func GetFile(name string) (*file.File, error) {
	file := &file.File{}

	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(fileBucket)
		n := strings.ToLower(name)

		result := b.Get([]byte(n))
		if result == nil {
			return errors.Errorf("\"%s\" does not exist", name)
		}

		decImg, err := crypt.Decrypt(result)
		if err != nil {
			return errors.Wrap(err, "decrypt file")
		}

		if err := proto.Unmarshal(decImg, file); err != nil {
			return errors.Wrap(err, "unmarshal file")
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return file, nil
}

// ListFiles returns a slice with all the files stored in the file bucket.
func ListFiles() ([]*file.File, error) {
	var files []*file.File

	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(fileBucket)
		c := b.Cursor()

		// Place cursor in the first line of the bucket and move it to the next one
		for k, v := c.First(); k != nil; k, v = c.Next() {
			file := &file.File{}

			decImg, err := crypt.Decrypt(v)
			if err != nil {
				return errors.Wrap(err, "decrypt file")
			}

			if err := proto.Unmarshal(decImg, file); err != nil {
				return errors.Wrap(err, "unmarshal file")
			}

			files = append(files, file)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}
