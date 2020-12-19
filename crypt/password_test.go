package crypt

import (
	"bytes"
	"crypto/subtle"
	"testing"

	"github.com/awnumar/memguard"
	"github.com/spf13/viper"
)

func TestGetMasterPassword(t *testing.T) {
	viper.Reset()
	password := memguard.NewBufferFromBytes([]byte("test password"))
	expected := password.Bytes()
	viper.Set("user.password", password.Seal())

	enclave, err := GetMasterPassword()
	if err != nil {
		t.Fatalf("GetMasterPassword() failed: %v", err)
	}

	got, err := enclave.Open()
	if err != nil {
		t.Errorf("Failed opening enclave: %v", err)
	}

	if subtle.ConstantTimeCompare(got.Bytes(), expected) == 0 {
		t.Errorf("Expected %s, got %s", string(expected), got.String())
	}
	got.Destroy()
}

func TestGetMasterPasswordDefault(t *testing.T) {
	viper.Reset()
	viper.Set("user.password", nil)

	// This should be tested as a success but we are not testing AskPassword
	_, err := GetMasterPassword()
	if err == nil {
		t.Fatal("Expected GetMasterPassword() to fail but it didn't")
	}
}

func TestZero(t *testing.T) {
	buf := []byte("test")

	zero(buf)

	if !bytes.Equal(buf, []byte{0, 0, 0, 0}) {
		t.Error("Failed wiping the buffer")
	}
}
