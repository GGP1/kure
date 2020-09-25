package crypt

import (
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
)

// EncryptedFile creates a file with Chacha20Poly1305 encryption in the path specified.
func EncryptedFile(data []byte, filename string) error {
	encrypted, err := Encrypt(data)
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
}

// DecryptFile an encrypted file and decrypts it.
func DecryptFile(filename string) ([]byte, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, errors.Wrap(err, "read database file")
	}

	return Decrypt(data)
}
