package crypt

import (
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
)

// CreateEncFile creates a file with XChacha20Poly1305 encryption in the specified path.
func CreateEncFile(data []byte, filename string) error {
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

// DecryptEncFile takes an encrypted file and decrypts it.
func DecryptEncFile(filename string) ([]byte, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, errors.Wrapf(err, "failed reading file on path %s", filename)
	}

	return Decrypt(data)
}
