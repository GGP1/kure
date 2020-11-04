package crypt

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"io"
	"io/ioutil"
	"syscall"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/ssh/terminal"
)

// Encrypt ciphers data.
func Encrypt(data []byte) ([]byte, error) {
	password, err := GetMasterPassword()
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

// Decrypt deciphers data.
func Decrypt(data []byte) ([]byte, error) {
	password, err := GetMasterPassword()
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

// EncryptX is like Encrypt but takes a password.
// Useful for using it inside a loop.
func EncryptX(data, password []byte) ([]byte, error) {
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

// DecryptX is like Encrypt but takes a password.
// Useful for using it inside a loop.
func DecryptX(data, password []byte) ([]byte, error) {
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

// AskPassword returns the hash of the input password.
func AskPassword(confirm bool) (string, error) {
	fmt.Print("Enter master password: ")
	password, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", errors.Wrap(err, "reading password")
	}
	fmt.Print("\n")

	if confirm {
		fmt.Print("\nRetype to verify: ")
		password2, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return "", errors.Wrap(err, "reading password")
		}
		fmt.Print("\n")

		if bytes.Compare(bytes.TrimSpace(password), bytes.TrimSpace(password2)) != 0 {
			return "", errors.New("passwords must be equal")
		}
	}

	password = bytes.TrimSpace(password)
	h := sha512.New()

	_, err = h.Write(password)
	if err != nil {
		return "", errors.Wrap(err, "password hash")
	}

	p := fmt.Sprintf("%x", h.Sum(nil))

	return p, nil
}

// GetMasterPassword takes the user master password from the config
// or requests it.
func GetMasterPassword() ([]byte, error) {
	password := viper.GetString("user.password")
	if password != "" {
		return []byte(password), nil
	}

	filename := viper.GetString("user.password_path")
	if filename != "" {
		p, err := ioutil.ReadFile(filename)
		if err != nil {
			return nil, errors.Wrap(err, "reading file")
		}

		p = bytes.TrimSpace(p)
		h := sha512.New()

		_, err = h.Write(p)
		if err != nil {
			return nil, errors.Wrap(err, "password hash")
		}

		pwd := fmt.Sprintf("%x", h.Sum(nil))

		return []byte(pwd), nil
	}

	pwd, err := AskPassword(false)
	if err != nil {
		return nil, err
	}

	return []byte(pwd), nil
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
