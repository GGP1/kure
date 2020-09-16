package crypt

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"syscall"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/crypto/twofish"
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

	gcm, err := chooseAlgorithm(hash)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())

	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, errors.Wrap(err, "reading nonce")
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)

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

	gcm, err := chooseAlgorithm(hash)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, errors.New("invalid password")
	}

	return plaintext, nil
}

// Create a HMAC-SHA256 hash (32 bytes) with the key provided.
func createHash(key []byte) ([]byte, error) {
	hasher := hmac.New(sha256.New, key)

	_, err := hasher.Write(key)
	if err != nil {
		return nil, errors.Wrap(err, "create hmac sha-256 hash")
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

func chooseAlgorithm(hash []byte) (cipher.AEAD, error) {
	var block cipher.Block
	var err error

	algorithm := strings.ToLower(viper.GetString("algorithm"))

	switch algorithm {
	case "aes":
		block, err = aes.NewCipher(hash)
		if err != nil {
			return nil, errors.Wrap(err, "create cipher")
		}
	case "twofish":
		block, err = twofish.NewCipher(hash)
		if err != nil {
			return nil, errors.Wrap(err, "create cipher")
		}
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.Wrap(err, "create GCM")
	}

	return gcm, nil
}
