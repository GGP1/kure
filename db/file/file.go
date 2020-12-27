package file

import (
	"bytes"
	"compress/gzip"
	"io"
	"strings"

	"github.com/GGP1/kure/crypt"
	"github.com/GGP1/kure/pb"
	"github.com/awnumar/memguard"
	"google.golang.org/protobuf/proto"

	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
)

var (
	fileBucket       = []byte("kure_file")
	errInvalidBucket = errors.New("invalid bucket")
)

// Create a new file.
func Create(db *bolt.DB, file *pb.File) error {
	// Compress file
	var gzipBuf bytes.Buffer
	gw := gzip.NewWriter(&gzipBuf)

	_, err := gw.Write(file.Content)
	if err != nil {
		return errors.Wrap(err, "failed compressing file")
	}

	if err := gw.Close(); err != nil {
		return errors.Wrap(err, "failed closing gzip writer")
	}

	file.Content = gzipBuf.Bytes()

	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(fileBucket)
		if b == nil {
			return errInvalidBucket
		}

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
func Get(db *bolt.DB, name string) (*memguard.LockedBuffer, *pb.File, error) {
	lockedBuf, file := pb.SecureFile()

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(fileBucket)
		if b == nil {
			return errInvalidBucket
		}
		name = strings.TrimSpace(strings.ToLower(name))

		encFile := b.Get([]byte(name))
		if encFile == nil {
			return errors.Errorf("%q does not exist", name)
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
		return nil, nil, err
	}

	// Decompress
	compressed := bytes.NewBuffer(file.Content)
	gr, err := gzip.NewReader(compressed)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed decompressing file")
	}
	defer gr.Close()

	var decompressed bytes.Buffer
	_, err = io.Copy(&decompressed, gr)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed copying decompressed file")
	}

	file.Content = decompressed.Bytes()

	return lockedBuf, file, nil
}

// GetCheap is like Get but it doesn't handle the file content.
func GetCheap(db *bolt.DB, name string) (*memguard.LockedBuffer, *pb.FileCheap, error) {
	lockedBuf, file := pb.SecureFileCheap()

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(fileBucket)
		if b == nil {
			return errInvalidBucket
		}
		name = strings.TrimSpace(strings.ToLower(name))

		encFile := b.Get([]byte(name))
		if encFile == nil {
			return errors.Errorf("%q does not exist", name)
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
		return nil, nil, err
	}

	return lockedBuf, file, nil
}

// List returns a slice with all the files stored in the file bucket.
func List(db *bolt.DB) (*memguard.LockedBuffer, []*pb.File, error) {
	lockedBuf, files := pb.SecureFileSlice()

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(fileBucket)
		if b == nil {
			return errInvalidBucket
		}

		return b.ForEach(func(k, v []byte) error {
			file := &pb.File{}

			decFile, err := crypt.Decrypt(v)
			if err != nil {
				return errors.Wrap(err, "decrypt file")
			}

			if err := proto.Unmarshal(decFile, file); err != nil {
				return errors.Wrap(err, "unmarshal file")
			}

			// Decrompress
			compressed := bytes.NewBuffer(file.Content)
			gr, err := gzip.NewReader(compressed)
			if err != nil {
				return errors.Wrap(err, "failed decompressing file")
			}
			defer gr.Close()

			var decompressed bytes.Buffer
			_, err = io.Copy(&decompressed, gr)
			if err != nil {
				return errors.Wrap(err, "failed copying decompressed file")
			}

			file.Content = decompressed.Bytes()
			files = append(files, file)

			return nil
		})
	})
	if err != nil {
		return nil, nil, err
	}

	return lockedBuf, files, nil
}

// ListFastest is used to check if the user entered the correct password
// by trying to decrypt every record and returning the fastest result.
func ListFastest(db *bolt.DB) bool {
	succeed := make(chan bool)

	decrypt := func(v []byte) {
		_, err := crypt.Decrypt(v)
		if err != nil {
			succeed <- false
		}

		succeed <- true
	}

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(fileBucket)

		return b.ForEach(func(_, v []byte) error {
			go decrypt(v)
			return nil
		})
	})

	return <-succeed
}

// ListNames returns a slice with all the files names.
func ListNames(db *bolt.DB) ([]string, error) {
	var files []string

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(fileBucket)
		if b == nil {
			return errInvalidBucket
		}

		return b.ForEach(func(k, _ []byte) error {
			files = append(files, string(k))
			return nil
		})
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}

// Remove removes a file from the database.
func Remove(db *bolt.DB, name string) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(fileBucket)
		if b == nil {
			return errInvalidBucket
		}
		name = strings.TrimSpace(strings.ToLower(name))

		file := b.Get([]byte(name))
		if file == nil {
			return errors.Errorf("%q does not exist", name)
		}

		if err := b.Delete([]byte(name)); err != nil {
			return errors.Wrap(err, "remove file")
		}

		return nil
	})
}

// Rename replaces an existing file name with the specified one.
func Rename(db *bolt.DB, oldName, newName string) error {
	oldName = strings.TrimSpace(strings.ToLower(oldName))
	newName = strings.TrimSpace(strings.ToLower(newName))

	oldBuf, file, err := Get(db, oldName)
	if err != nil {
		return err
	}
	defer oldBuf.Destroy()
	file.Name = newName

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(fileBucket)

		exists := b.Get([]byte(newName))
		if exists != nil {
			return errors.Errorf("already exists a file named %q", newName)
		}

		if err := b.Delete([]byte(oldName)); err != nil {
			return errors.Wrap(err, "delete old file")
		}

		return nil
	})
	if err != nil {
		return err
	}

	// Executing nested db.Update() generates a deadlock
	if err := Create(db, file); err != nil {
		return errors.Wrap(err, "save new file")
	}
	return nil
}

// Restore is like Create but do not check for existing files.
// Should be used in the "restore" command only.
//
// Do not pass a locked buffer since the file passed comes from a protected slice.
func Restore(db *bolt.DB, file *pb.File) error {
	// Compress file
	var gzipBuf bytes.Buffer
	gw := gzip.NewWriter(&gzipBuf)

	_, err := gw.Write(file.Content)
	if err != nil {
		return errors.Wrap(err, "failed compressing file")
	}

	if err := gw.Close(); err != nil {
		return errors.Wrap(err, "failed closing gzip writer")
	}

	file.Content = gzipBuf.Bytes()

	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(fileBucket)
		if b == nil {
			return errInvalidBucket
		}

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
