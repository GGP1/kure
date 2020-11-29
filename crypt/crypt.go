package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
	"runtime"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"golang.org/x/crypto/argon2"
)

// Encrypt ciphers data.
func Encrypt(data []byte) ([]byte, error) {
	password, err := GetMasterPassword()
	if err != nil {
		return nil, err
	}

	lockedBuf, err := password.Open()
	if err != nil {
		return nil, errors.Wrap(err, "failed opening enclave")
	}

	key, salt, err := deriveKey(lockedBuf.Bytes(), nil)
	if err != nil {
		return nil, err
	}
	go lockedBuf.Destroy()

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.Wrap(err, "failed creating block")
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.Wrap(err, "failed creating gcm")
	}

	nonce := make([]byte, gcm.NonceSize())

	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	ciphertext = append(ciphertext, salt...)

	return ciphertext, nil
}

// Decrypt deciphers data.
func Decrypt(data []byte) ([]byte, error) {
	salt, data := data[len(data)-32:], data[:len(data)-32]

	password, err := GetMasterPassword()
	if err != nil {
		return nil, err
	}

	lockedBuf, err := password.Open()
	if err != nil {
		return nil, errors.Wrap(err, "failed opening enclave")
	}

	key, _, err := deriveKey(lockedBuf.Bytes(), salt)
	if err != nil {
		return nil, err
	}
	go lockedBuf.Destroy()

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.Wrap(err, "failed creating block")
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.Wrap(err, "failed creating gcm")
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("data is too short")
	}

	return gcm.Open(nil, data[:nonceSize], data[nonceSize:], nil)
}

// deriveKey derives the key from the password, salt and other parameters using
// the key derivation function argon2id. Use parameters from the configuration file if they exist.
func deriveKey(key []byte, salt []byte) ([]byte, []byte, error) {
	var (
		iters   uint32 = 1
		memory  uint32 = 1 << 20 // 1048576
		threads uint8  = uint8(runtime.NumCPU())
	)

	if i := viper.GetUint32("argon2id.iterations"); i != 0 {
		iters = i
	}
	if m := viper.GetUint32("argon2id.memory"); m != 0 {
		memory = m
	}

	if salt == nil {
		salt = make([]byte, 32)
		_, err := rand.Read(salt)
		if err != nil {
			return nil, nil, errors.New("failed generating salt")
		}
	}

	password := argon2.IDKey(key, salt, iters, memory, threads, 32)

	return password, salt, nil
}
