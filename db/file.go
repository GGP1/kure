package db

import (
	"strings"

	"github.com/GGP1/kure/crypt"
	"github.com/GGP1/kure/pb"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
)

// CreateFile saves a new file into the database.
func CreateFile(file *pb.File) error {
	files, err := FilesByName(file.Name)
	if err != nil {
		return err
	}

	var exists bool
	for _, f := range files {
		if strings.Split(f.Name, "/")[0] == file.Name {
			exists = true
			break
		}
	}

	if exists {
		return errors.New("already exists a file or folder with this name")
	}

	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(fileBucket)

		exists := b.Get([]byte(file.Name))
		if exists != nil {
			return errors.Errorf("already exists a file named %s", file.Name)
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

// CreateFileX is like CreateFile but takes a password.
// Useful for using it inside a loop.
func CreateFileX(file *pb.File, password []byte) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(fileBucket)

		exists := b.Get([]byte(file.Name))
		if exists != nil {
			return errors.Errorf("already exists a file named %s", file.Name)
		}

		buf, err := proto.Marshal(file)
		if err != nil {
			return errors.Wrap(err, "marshal file")
		}

		encFile, err := crypt.EncryptX(buf, password)
		if err != nil {
			return errors.Wrap(err, "encrypt file")
		}

		if err := b.Put([]byte(file.Name), encFile); err != nil {
			return errors.Wrap(err, "save file")
		}

		return nil
	})
}

// FilesByName filters all the entries and returns those matching with the name passed.
func FilesByName(name string) ([]*pb.File, error) {
	var group []*pb.File
	name = strings.TrimSpace(strings.ToLower(name))

	files, err := ListFiles()
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		if strings.Contains(f.Name, name) {
			group = append(group, f)
		}
	}

	if len(group) == 0 {
		return nil, errors.New("no files were found")
	}

	return group, nil
}

// GetFile retrieves the file with the specified name.
func GetFile(name string) (*pb.File, error) {
	file := &pb.File{}

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(fileBucket)
		name = strings.TrimSpace(strings.ToLower(name))

		encFile := b.Get([]byte(name))
		if encFile == nil {
			return errors.Errorf("\"%s\" does not exist", name)
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

// ListFiles returns a slice with all the files stored in the file bucket.
func ListFiles() ([]*pb.File, error) {
	var files []*pb.File

	password, err := crypt.GetMasterPassword()
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(fileBucket)
		c := b.Cursor()

		// Place cursor in the first line of the bucket and move it to the next one
		for k, v := c.First(); k != nil; k, v = c.Next() {
			file := &pb.File{}

			decFile, err := crypt.DecryptX(v, password)
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

// RemoveFile removes a file from the database.
func RemoveFile(name string) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(fileBucket)
		name = strings.TrimSpace(strings.ToLower(name))

		_, err := GetFile(name)
		if err != nil {
			return err
		}

		if err := b.Delete([]byte(name)); err != nil {
			return errors.Wrap(err, "remove file")
		}

		return nil
	})
}

// RenameFile replaces an existing file name with the specified one.
func RenameFile(oldName, newName string) error {
	// Executing nested "db.Update" generates a deadlock
	file, err := GetFile(oldName)
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
			return errors.Errorf("already exists a file named %s", newName)
		}

		if err := b.Delete([]byte(oldName)); err != nil {
			return errors.Wrap(err, "delete old entry")
		}

		return nil
	})
	if err != nil {
		return err
	}

	if err := CreateFile(file); err != nil {
		return errors.Wrap(err, "save new entry")
	}
	return nil
}
