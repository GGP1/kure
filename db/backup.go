package db

import (
	"bytes"
	"fmt"
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

// EncryptedBackup creates a file with AES encryption in the path specified.
func EncryptedBackup(filename string, passphrase []byte) error {
	return db.View(func(tx *bolt.Tx) error {
		var buf bytes.Buffer

		_, err := tx.WriteTo(&buf)
		if err != nil {
			return err
		}

		encrypted, err := crypt.Encrypt(buf.Bytes(), passphrase)
		if err != nil {
			return err
		}

		path := fmt.Sprintf("%s.db", filename)

		f, err := os.Create(path)
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

// DecryptBackup takes the database backup and decrypts it
func DecryptBackup(filename string, passphrase []byte) ([]byte, error) {
	path := fmt.Sprintf("%s.db", filename)

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "read database file")
	}

	return crypt.Decrypt(data, passphrase)
}
