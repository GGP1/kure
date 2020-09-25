package crypt

import (
	"crypto/rand"
	"io"
	"testing"

	"github.com/pkg/errors"
	"golang.org/x/crypto/chacha20poly1305"
)

func TestCrypt(t *testing.T) {
	testCases := []struct {
		data string
	}{
		{"kure cli password manager"},
		{"encrypting and decrypting"},
		{"chacha20-poly1305"},
		{"sha-256"},
	}

	passphrase := []byte("crypt test")

	for _, tC := range testCases {
		ciphertext, err := mockEncrypt([]byte(tC.data), passphrase)
		if err != nil {
			t.Errorf("Failed encrypting data: %v", err)
		}

		if tC.data == string(ciphertext) {
			t.Error("Test failed, data hasn't been encrypted correctly")
		}

		plaintext, err := mockDecrypt(ciphertext, passphrase)
		if err != nil {
			t.Errorf("Failed decrypting data: %v", err)
		}

		if tC.data != string(plaintext) {
			t.Errorf("Test failed, expected: %s, got: %s", tC.data, string(plaintext))
		}
	}
}

// mockEntrypt and mockDecrypt differ from Encrypt and Decrypt only in
// that they receive the passphrase as a parameter.
func mockEncrypt(data, passphrase []byte) ([]byte, error) {
	hash, err := createHash(passphrase)
	if err != nil {
		return nil, err
	}

	AEAD, err := chacha20poly1305.NewX(hash)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, AEAD.NonceSize())

	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, errors.Wrap(err, "reading nonce")
	}

	ciphertext := AEAD.Seal(nonce, nonce, data, nil)

	return ciphertext, nil
}

func mockDecrypt(data, passphrase []byte) ([]byte, error) {
	hash, err := createHash(passphrase)
	if err != nil {
		return nil, err
	}

	AEAD, err := chacha20poly1305.NewX(hash)
	if err != nil {
		return nil, err
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
