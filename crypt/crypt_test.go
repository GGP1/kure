package crypt

import (
	"testing"

	"github.com/awnumar/memguard"
	"github.com/spf13/viper"
)

func TestCrypt(t *testing.T) {
	cases := []struct {
		data     string
		password string
	}{
		{"kure cli password manager", "test1"},
		{"encrypting and decrypting", "test2"},
		{"chacha20-poly1305", "test3"},
		{"sha-256", "test4"},
	}

	for _, tc := range cases {
		viper.Reset()
		password := memguard.NewBufferFromBytes([]byte(tc.password))
		viper.Set("user.password", password.Seal())

		ciphertext, err := Encrypt([]byte(tc.data))
		if err != nil {
			t.Fatalf("Encrypt() failed: %v", err)
		}

		if tc.data == string(ciphertext) {
			t.Error("Data hasn't been encrypted")
		}

		plaintext, err := Decrypt(ciphertext)
		if err != nil {
			t.Fatalf("Decrypt() failed: %v", err)
		}

		if tc.data != string(plaintext) {
			t.Errorf("Expected: %q, got: %q", tc.data, string(plaintext))
		}
	}
}

func TestInvalidPasswordEncrypt(t *testing.T) {
	viper.Reset()

	_, err := Encrypt([]byte("test_fail"))
	if err == nil {
		t.Error("Expected Encrypt() to fail but got nil")
	}
}

func TestInvalidPasswordDecrypt(t *testing.T) {
	viper.Reset()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected Decrypt() to fail but got nil")
		}
	}()

	Decrypt([]byte("test_fail"))
}
