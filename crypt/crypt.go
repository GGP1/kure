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

	password, err := GetMasterPassword()
	if err != nil {
		return nil, err
	}

	lockedBuf, err := password.Open()
	if err != nil {
		return nil, errEncrypt
	}

	key, salt, err := deriveKey(lockedBuf.Bytes(), nil)
	lockedBuf.Destroy()
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errEncrypt
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errEncrypt
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
	if data == nil {
		return nil, errDecrypt
	}

	salt, data := data[len(data)-saltSize:], data[:len(data)-saltSize]

	password, err := GetMasterPassword()
	if err != nil {
		return nil, err
	}

	lockedBuf, err := password.Open()
	if err != nil {
		return nil, errDecrypt
	}

	key, _, err := deriveKey(lockedBuf.Bytes(), salt)
	lockedBuf.Destroy()
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errDecrypt
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errDecrypt
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errDecrypt
	}

	plaintext, err := gcm.Open(nil, data[:nonceSize], data[nonceSize:], nil)
	if err != nil {
		return nil, errDecrypt
	}

	return plaintext, nil
}

// deriveKey derives the key from the password, salt and other parameters using
// the key derivation function argon2id. Use parameters from the configuration file if they exist.
func deriveKey(key []byte, salt []byte) ([]byte, []byte, error) {
	var (
		iters uint32 = 1
		// memory is measured in kibibytes, 1 kibibyte = 1024 bytes.
		memory  uint32 = 1 << 20 // 1048576 kibibytes -> 1GB
		threads uint8  = uint8(runtime.NumCPU())
	)

	if i := viper.GetUint32("argon2.iterations"); i > 0 {
		iters = i
	}
	if m := viper.GetUint32("argon2.memory"); m > 0 {
		memory = m
	}
	if t := viper.GetUint32("argon2.threads"); t > 0 {
		threads = uint8(t)
	}

	if salt == nil {
		salt = make([]byte, saltSize)
		if _, err := rand.Read(salt); err != nil {
			return nil, nil, errors.New("failed generating salt")
		}
	}

	password := argon2.IDKey(key, salt, iters, memory, threads, 32)

	return password, salt, nil
}
