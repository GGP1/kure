package crypt

import (
	"crypto/rand"
	"crypto/subtle"
	"testing"

	"github.com/GGP1/kure/config"

	"github.com/awnumar/memguard"
	"github.com/stretchr/testify/assert"
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
		config.Set("auth.password", memguard.NewEnclave([]byte(tc.password)))

		ciphertext, err := Encrypt([]byte(tc.data))
		assert.NoError(t, err)

		assert.NotEqual(t, string(ciphertext), tc.data, "Data hasn't been encrypted")

		plaintext, err := Decrypt(ciphertext)
		assert.NoError(t, err)

		assert.Equal(t, string(plaintext), tc.data)
	}
}

func TestInvalidData(t *testing.T) {
	_, err := Encrypt(nil)
	assert.Error(t, err)

	_, err = Decrypt(nil)
	assert.Error(t, err)
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
				r := recover()
				assert.NotNil(t, r, "Expected Decrypt() to panic")
			}()

			config.Set("auth.password", tc.key)

			Decrypt([]byte(tc.data))
		})
	}
}

func TestDecryptError(t *testing.T) {
	config.Set("auth.password", memguard.NewEnclave([]byte("test")))

	// Data must be between 32 and 45 bytes long to fail
	data := []byte("t8aNDgbSxlnPn ehxsYFnuDwzU4eqgydh2k")

	_, err := Decrypt(data)
	assert.Error(t, err)
}

func TestDeriveKey(t *testing.T) {
	reduceArgon2Params(t)

	salt := make([]byte, 32)
	_, _ = rand.Read(salt)

	key := memguard.NewEnclave([]byte("test"))
	config.Set("auth.password", key)

	cases := []struct {
		setDefaults func()
		desc        string
		salt        []byte
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
				config.Set("auth.iterations", "1")
				config.Set("auth.memory", "5000")
				config.Set("auth.threads", "4")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			if tc.setDefaults != nil {
				tc.setDefaults()
			}

			pwd, salt, err := deriveKey(tc.salt)
			assert.NoError(t, err)
			password := pwd.Bytes()

			assert.Equal(t, 32, len(password))
			assert.Equal(t, saltSize, len(salt))

			keyBuf, err := key.Open()
			assert.NoError(t, err, "Failed opening key enclave")

			comparison := subtle.ConstantTimeCompare(password, keyBuf.Bytes())
			assert.Equal(t, 0, comparison, "KDF failed, expected a different password and got the same one")
		})
	}
}

func reduceArgon2Params(t *testing.T) {
	t.Helper()

	// Reduce argon2 parameters to speed up tests
	config.Set("auth.memory", 1)
	config.Set("auth.iterations", 1)
	config.Set("auth.threads", 1)
}
