package file

import (
	"strings"

	"github.com/GGP1/kure/crypt"
	"github.com/GGP1/kure/pb"
	"google.golang.org/protobuf/proto"

	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
)

var (
	fileBucket       = []byte("kure_file")
	errInvalidBucket = errors.New("invalid bucket")
)

// Create saves a new file into the database.
func Create(db *bolt.DB, file *pb.File) error {
	files, err := ListNames(db)
	if err != nil {
		return err
	}

	for _, f := range files {
		if f.Name == file.Name {
			return errors.Errorf("already exists a file or folder named %q", file.Name)
		}
	}

	return db.Update(func(tx *bolt.Tx) error {
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
		return nil, err
	}

	return file, nil
}

// List returns a slice with all the files stored in the file bucket.
func List(db *bolt.DB) ([]*pb.File, error) {
	var files []*pb.File

	_, err := crypt.GetMasterPassword()
	if err != nil {
		return nil, err
	}

	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(fileBucket)
		if b == nil {
			return errInvalidBucket
		}
		c := b.Cursor()

		// Place cursor in the first line of the bucket and move it to the next one
		for k, v := c.First(); k != nil; k, v = c.Next() {
			file := &pb.File{}

			decFile, err := crypt.Decrypt(v)
			if err != nil {
				return errors.Wrap(err, "decrypt file")
			}

			if err := proto.Unmarshal(decFile, file); err != nil {
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

// ListByName filters all the entries and returns those matching with the name passed.
func ListByName(db *bolt.DB, name string) ([]*pb.File, error) {
	var group []*pb.File
	name = strings.TrimSpace(strings.ToLower(name))

	files, err := List(db)
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		if strings.Contains(f.Name, name) {
			group = append(group, f)
		}
	}

	return group, nil
}

// ListNames returns a slice with all the files names.
func ListNames(db *bolt.DB) ([]*pb.FileList, error) {
	var files []*pb.FileList

	_, err := crypt.GetMasterPassword()
	if err != nil {
		return nil, err
	}

	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(fileBucket)
		if b == nil {
			return errInvalidBucket
		}
		c := b.Cursor()

		// Place cursor in the first line of the bucket and move it to the next one
		for k, v := c.First(); k != nil; k, v = c.Next() {
			file := &pb.FileList{}

			decFile, err := crypt.Decrypt(v)
			if err != nil {
				return errors.Wrap(err, "decrypt file")
			}

			if err := proto.Unmarshal(decFile, file); err != nil {
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

// Remove removes a file from the database.
func Remove(db *bolt.DB, name string) error {
	_, err := Get(db, name)
	if err != nil {
		return err
	}

	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(fileBucket)
		name = strings.TrimSpace(strings.ToLower(name))

		if err := b.Delete([]byte(name)); err != nil {
			return errors.Wrap(err, "remove file")
		}

		return nil
	})
}

// Rename replaces an existing file name with the specified one.
func Rename(db *bolt.DB, oldName, newName string) error {
	file, err := Get(db, oldName)
	if err != nil {
		return err
	}
	file.Name = newName

	oldName = strings.TrimSpace(strings.ToLower(oldName))
	newName = strings.TrimSpace(strings.ToLower(newName))

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(fileBucket)

		encFile := b.Get([]byte(newName))
		if encFile != nil {
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
