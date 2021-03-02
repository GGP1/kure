package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"

	"github.com/GGP1/kure/config"

	"github.com/awnumar/memguard"
	"github.com/pkg/errors"
	"golang.org/x/crypto/argon2"
)

const saltSize = 32

var (
	// Do not provide the reason of failure to potential attackers
	errEncrypt = errors.New("encryption failed")
	errDecrypt = errors.New("decryption failed")
)

// Encrypt ciphers data.
func Encrypt(data []byte) ([]byte, error) {
	if data == nil {
		return nil, errEncrypt
	}

	key, salt, err := deriveKey(nil)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key.Bytes())
	if err != nil {
		return nil, errEncrypt
	}
	key.Destroy()

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errEncrypt
	}

	// make 12 byte long nonce
	nonce := make([]byte, gcm.NonceSize())

	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, errEncrypt
	}

	dst := make([]byte, gcm.NonceSize())
	copy(dst, nonce)

	// Encrypt, authenticate and append the salt to the end of it
	ciphertext := gcm.Seal(dst, nonce, data, nil)
	ciphertext = append(ciphertext, salt...)

	return ciphertext, nil
}

// Decrypt deciphers data.
func Decrypt(data []byte) ([]byte, error) {
	if data == nil {
		return nil, errDecrypt
	}

	// Split salt (last 32 bytes) from the data
	salt, data := data[len(data)-saltSize:], data[:len(data)-saltSize]

	key, _, err := deriveKey(salt)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key.Bytes())
	if err != nil {
		return nil, errDecrypt
	}
	key.Destroy()

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errDecrypt
	}

	// The nonce is 12 bytes long
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errDecrypt
	}

	// Decrypt and authenticate ciphertext
	plaintext, err := gcm.Open(nil, data[:nonceSize], data[nonceSize:], nil)
	if err != nil {
		return nil, errDecrypt
	}

	return plaintext, nil
}

// deriveKey derives the key from the password, salt and other parameters using
// the key derivation function argon2id.
//
// It destroys the buffer of the enclave passed and returns the derived key, and the salt used.
func deriveKey(salt []byte) (*memguard.LockedBuffer, []byte, error) {
	password := config.GetEnclave("auth.password")
	iters := config.GetUint32("auth.iterations")
	memory := config.GetUint32("auth.memory")
	threads := config.GetUint32("auth.threads")

	// When decrypting the salt is taken from the encrypted data and when encrypting it's randomly generated
	if salt == nil {
		salt = make([]byte, saltSize)
		if _, err := rand.Read(salt); err != nil {
			return nil, nil, errors.New("generating salt")
		}
	}

	// Decrypt enclave and save its content in a locked buffer
	pwd, err := password.Open()
	if err != nil {
		return nil, nil, errors.New("decrypting key")
	}

	derivedKey := memguard.NewBufferFromBytes(argon2.IDKey(pwd.Bytes(), salt, iters, memory, uint8(threads), 32))
	pwd.Destroy()

	return derivedKey, salt, nil
}
