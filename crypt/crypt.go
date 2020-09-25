package crypt

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"syscall"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/ssh/terminal"
)

// Encrypt ciphers data with the given passphrase.
func Encrypt(data []byte) ([]byte, error) {
	passphrase, err := getMasterPassword()
	if err != nil {
		return nil, err
	}

	if string(passphrase) == "" {
		fmt.Print("Enter Passphrase: ")
		passphrase, err = terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return nil, errors.Wrap(err, "reading password")
		}
		fmt.Print("\nConfirm Passphrase: ")
		passphrase2, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return nil, errors.Wrap(err, "reading password")
		}

		if bytes.Compare(passphrase, passphrase2) != 0 {
			fmt.Println("")
			return nil, errors.New("passphrases must be equal")
		}
	}

	hash, err := createHash(passphrase)
	if err != nil {
		return nil, err
	}

	AEAD, err := chacha20poly1305.NewX(hash)
	if err != nil {
		return nil, errors.Wrap(err, "creating AEAD")
	}

	nonce := make([]byte, AEAD.NonceSize())

	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, errors.Wrap(err, "reading nonce")
	}

	ciphertext := AEAD.Seal(nonce, nonce, data, nil)

	return ciphertext, nil
}

// Decrypt deciphers data with the given passphrase.
func Decrypt(data []byte) ([]byte, error) {
	passphrase, err := getMasterPassword()
	if err != nil {
		return nil, err
	}

	if string(passphrase) == "" {
		fmt.Print("Enter Passphrase: ")
		passphrase, err = terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return nil, errors.Wrap(err, "reading password")
		}
	}

	hash, err := createHash(passphrase)
	if err != nil {
		return nil, err
	}

	AEAD, err := chacha20poly1305.NewX(hash)
	if err != nil {
		return nil, errors.Wrap(err, "creating AEAD")
	}

	nonceSize := AEAD.NonceSize()

	if len(data) < nonceSize {
		return nil, errors.New("encrypted data is too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	plaintext, err := AEAD.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, errors.New("invalid password")
	}

	return plaintext, nil
}

// Create a SHA256 hash (32 bytes) with the key provided.
func createHash(key []byte) ([]byte, error) {
	hasher := sha256.New()

	_, err := hasher.Write(key)
	if err != nil {
		return nil, errors.Wrap(err, "create sha-256 hash")
	}

	return hasher.Sum(nil), nil
}

func getMasterPassword() ([]byte, error) {
	masterPwd := viper.GetString("user.password")
	passphrase := []byte(masterPwd)

	if masterPwd == "" {
		filename := viper.GetString("user.password_path")

		mPassword, err := ioutil.ReadFile(filename)
		if err != nil {
			return nil, errors.Wrap(err, "reading file")
		}

		passphrase = mPassword
	}

	return passphrase, nil
}
