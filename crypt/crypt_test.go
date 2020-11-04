package crypt

import (
	"testing"
)

func TestCrypt(t *testing.T) {
	testCases := []struct {
		data     string
		password string
	}{
		{"kure cli password manager", "test1"},
		{"encrypting and decrypting", "test2"},
		{"chacha20-poly1305", "test3"},
		{"sha-256", "test4"},
	}

	for _, tC := range testCases {
		ciphertext, err := EncryptX([]byte(tC.data), []byte(tC.password))
		if err != nil {
			t.Errorf("Failed encrypting data: %v", err)
		}

		if tC.data == string(ciphertext) {
			t.Error("Test failed, data hasn't been encrypted correctly")
		}

		plaintext, err := DecryptX(ciphertext, []byte(tC.password))
		if err != nil {
			t.Errorf("Failed decrypting data: %v", err)
		}

		if tC.data != string(plaintext) {
			t.Errorf("Test failed, expected: %s, got: %s", tC.data, string(plaintext))
		}
	}
}
