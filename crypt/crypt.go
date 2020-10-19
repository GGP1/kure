package crypt

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"syscall"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/ssh/terminal"
)

// Encrypt ciphers data with the given passphrase.
func Encrypt(data []byte) ([]byte, error) {
	password, err := getMasterPassword()
	if err != nil {
		return nil, err
	}

	hash, err := createHash(password)
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
	password, err := getMasterPassword()
	if err != nil {
		return nil, err
	}

	hash, err := createHash(password)
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

// Create a SHA256 hash (256 bits) with the key provided.
func createHash(key []byte) ([]byte, error) {
	h := sha256.New()

	_, err := h.Write(key)
	if err != nil {
		return nil, errors.Wrap(err, "create sha-256 hash")
	}

	return h.Sum(nil), nil
}

func getMasterPassword() ([]byte, error) {
	password := viper.GetString("user.password")
	if password != "" {
		return []byte(password), nil
	}

	filename := viper.GetString("user.password_path")
	if filename != "" {
		mPassword, err := ioutil.ReadFile(filename)
		if err != nil {
			return nil, errors.Wrap(err, "reading file")
		}

		p := strings.TrimSpace(string(mPassword))
		h := sha512.New()

		_, err = h.Write([]byte(p))
		if err != nil {
			return nil, errors.Wrap(err, "password hash")
		}

		return h.Sum(nil), nil
	}

	fmt.Print("Enter master password: ")
	masterPwd, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return nil, errors.Wrap(err, "reading password")
	}

	fmt.Print("\nConfirm master password: ")
	masterPwd2, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return nil, errors.Wrap(err, "reading confirmation password")
	}

	if bytes.Compare(masterPwd, masterPwd2) != 0 {
		fmt.Print("\n")
		return nil, errors.New("passwords must be equal")
	}

	p := strings.TrimSpace(string(masterPwd))
	h := sha512.New()

	_, err = h.Write([]byte(p))
	if err != nil {
		return nil, errors.Wrap(err, "password hash")
	}

	return h.Sum(nil), nil
}
