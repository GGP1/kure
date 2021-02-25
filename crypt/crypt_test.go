package crypt

import (
	"crypto/rand"
	"crypto/subtle"
	"testing"

	"github.com/awnumar/memguard"
	"github.com/spf13/viper"
)

func TestCrypt(t *testing.T) {
	reduceArgon2Params(t)

	cases := []struct {
		data     string
		password string
	}{
		{"kure cli password manager", "test1"},
		{"advanced standard encryption", "test2"},
	}

	for _, tc := range cases {
		viper.Set("auth.password", memguard.NewEnclave([]byte(tc.password)))

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

func TestInvalidData(t *testing.T) {
	if _, err := Encrypt(nil); err == nil {
		t.Error("Expected Encrypt() to fail but it didn't")
	}

	if _, err := Decrypt(nil); err == nil {
		t.Error("Expected Decrypt() to fail but it didn't")
	}
}

func TestDecryptPanics(t *testing.T) {
	cases := []struct {
		desc string
		key  *memguard.Enclave
		data string
	}{
		{
			desc: "Invalid password",
			key:  memguard.NewEnclave(nil),
			data: "test_invalid_password",
		},
		{
			desc: "Slice bounds out of range",
			key:  memguard.NewEnclave([]byte("test")),
			data: "short",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Error("Expected Decrypt() to panic but it didn't")
				}
			}()

			viper.Set("auth.password", tc.key)

			Decrypt([]byte(tc.data))
		})
	}
}

func TestDecryptError(t *testing.T) {
	viper.Set("auth.password", memguard.NewEnclave([]byte("test")))

	// Data must be between 32 and 45 bytes long to fail
	data := []byte("t8aNDgbSxlnPn ehxsYFnuDwzU4eqgydh2k")

	if _, err := Decrypt(data); err == nil {
		t.Error("Expected Decrypt() to fail and got nil")
	}
}

func TestDeriveKey(t *testing.T) {
	reduceArgon2Params(t)

	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		t.Fatalf("Failed generating salt: %v", err)
	}

	key := memguard.NewEnclave([]byte("test"))
	viper.Set("auth.password", key)

	cases := []struct {
		desc        string
		salt        []byte
		setDefaults func()
	}{
		{
			desc: "Predefined random salt",
			salt: salt,
		},
		{
			desc: "Generating random salt",
			salt: nil,
		},
		{
			desc: "Argon2 custom parameters",
			salt: nil,
			setDefaults: func() {
				viper.Set("auth.iterations", "1")
				viper.Set("auth.memory", "5000")
				viper.Set("auth.threads", "4")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			if tc.setDefaults != nil {
				tc.setDefaults()
			}

			pwd, salt, err := deriveKey(tc.salt)
			if err != nil {
				t.Fatal(err)
			}
			password := pwd.Bytes()

			if len(password) != 32 {
				t.Errorf("Expected a 32 byte long password, got %d bytes", len(password))
			}

			if len(salt) != saltSize {
				t.Errorf("Expected a 32 byte long salt, got %d bytes", len(salt))
			}

			keyBuf, err := key.Open()
			if err != nil {
				t.Errorf("Failed opening key enclave: %v", err)
			}

			if subtle.ConstantTimeCompare(password, keyBuf.Bytes()) == 1 {
				t.Error("KDF failed, expected a different password and got the same one")
			}
		})
	}
}

func reduceArgon2Params(t *testing.T) {
	t.Helper()

	// Reduce argon2 parameters to speed up tests
	viper.Set("auth.memory", 1)
	viper.Set("auth.iterations", 1)
	viper.Set("auth.threads", 1)
}
