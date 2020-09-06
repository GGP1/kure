package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"io"

	"github.com/pkg/errors"
)

func createHash(key []byte) ([]byte, error) {
	hasher := md5.New()

	_, err := hasher.Write(key)
	if err != nil {
		return nil, errors.Wrap(err, "create hash")
	}

	return hasher.Sum(nil), nil
}

// Encrypt ciphers data with the given passphrase.
func Encrypt(data []byte, passphrase []byte) ([]byte, error) {
	hash, err := createHash(passphrase)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(hash)
	if err != nil {
		return nil, errors.Wrap(err, "create cipher")
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.Wrap(err, "create GCM")
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
func Decrypt(data []byte, passphrase []byte) ([]byte, error) {
	hash, err := createHash(passphrase)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(hash)
	if err != nil {
		return nil, errors.Wrap(err, "create cipher")
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.Wrap(err, "create GCM")
	}

	nonceSize := gcm.NonceSize()

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, errors.Wrap(err, "invalid password")
	}

	return plaintext, nil
}
