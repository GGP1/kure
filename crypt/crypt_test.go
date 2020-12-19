package crypt

import (
	"crypto/rand"
	"crypto/subtle"
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
		{"advanced standard encryption", "test2"},
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

func TestCryptError(t *testing.T) {
	_, err := Encrypt(nil)
	if err == nil {
		t.Error("Expected Encrypt() to fail but got nil")
	}

	_, err = Decrypt(nil)
	if err == nil {
		t.Error("Expected Decrypt() to fail but got nil")
	}
}

func TestEncryptInvalidPassword(t *testing.T) {
	viper.Reset()

	_, err := Encrypt([]byte("test_fail"))
	if err == nil {
		t.Error("Expected Encrypt() to fail but got nil")
	}
}

func TestDecryptInvalidPassword(t *testing.T) {
	viper.Reset()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected Decrypt() to fail but got nil")
		}
	}()

	Decrypt([]byte("test_fail"))
}

func TestDeriveKey(t *testing.T) {
	salt := make([]byte, 32)
	_, err := rand.Read(salt)
	if err != nil {
		t.Fatalf("Failed generating salt: %v", err)
	}

	cases := []struct {
		desc string
		key  []byte
		salt []byte
	}{
		{
			desc: "Predefined random salt",
			key:  []byte("test"),
			salt: salt,
		},
		{
			desc: "Generating random salt",
			key:  []byte("test"),
			salt: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			password, salt, err := deriveKey(tc.key, tc.salt)
			if err != nil {
				t.Fatal(err)
			}

			if len(password) != 32 {
				t.Errorf("Expected a 32 byte long password, got %d bytes", len(password))
			}

			if len(salt) != 32 {
				t.Errorf("Expected a 32 byte long salt, got %d bytes", len(salt))
			}

			if subtle.ConstantTimeCompare(password, tc.key) == 1 {
				t.Error("KDF failed, expected a different password and got the same one")
			}
		})
	}
}
