package crypt

import (
	"testing"

	"github.com/spf13/viper"
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

	for _, tC := range testCases {
		viper.Set("user.password", "testingCrypt")

		ciphertext, err := Encrypt([]byte(tC.data))
		if err != nil {
			t.Errorf("Failed encrypting data: %v", err)
		}

		if tC.data == string(ciphertext) {
			t.Error("Test failed, data hasn't been encrypted correctly")
		}

		plaintext, err := Decrypt(ciphertext)
		if err != nil {
			t.Errorf("Failed decrypting data: %v", err)
		}

		if tC.data != string(plaintext) {
			t.Errorf("Test failed, expected: %s, got: %s", tC.data, string(plaintext))
		}
	}
}
