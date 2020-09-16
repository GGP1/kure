package db

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/GGP1/kure/crypt"

	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
)

// HTTPBackup writes a consistent view of the database to a http endpoint.
func HTTPBackup(w http.ResponseWriter, r *http.Request) {
	err := db.View(func(tx *bolt.Tx) error {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", `attachment; filename="kure.db"`)
		w.Header().Set("Content-Length", strconv.Itoa(int(tx.Size())))
		_, err := tx.WriteTo(w)
		return err
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// EncryptedFile creates a file with AES encryption in the path specified.
func EncryptedFile(filename string) error {
	return db.View(func(tx *bolt.Tx) error {
		var buf bytes.Buffer

		_, err := tx.WriteTo(&buf)
		if err != nil {
			return err
		}

		encrypted, err := crypt.Encrypt(buf.Bytes())
		if err != nil {
			return err
		}

		f, err := os.Create(filename)
		if err != nil {
			return errors.Wrap(err, "create encrypted file")
		}
		defer f.Close()

		_, err = f.Write(encrypted)
		if err != nil {
			return errors.Wrap(err, "write encrypted file")
		}

		return nil
	})
}

// DecryptFile takes the database backup and decrypts it.
func DecryptFile(filename string) ([]byte, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, errors.Wrap(err, "read database file")
	}

	return crypt.Decrypt(data)
}
