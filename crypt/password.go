package crypt

import (
	"bytes"
	"fmt"
	"reflect"
	"syscall"

	"github.com/awnumar/memguard"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh/terminal"
)

// AskPassword returns the hash of the input password.
//
// This function is not tested as stubbing terminal.ReadPassword() provides
// almost no benefits.
func AskPassword(verify bool) (*memguard.Enclave, error) {
	fmt.Print("Enter master password: ")
	password, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return nil, errors.Wrap(err, "reading password")
	}
	fmt.Print("\n")

	if password == nil {
		return nil, errors.New("invalid password")
	}

	pwd := memguard.NewBufferFromBytes(bytes.TrimSpace(password))

	if verify {
		fmt.Print("Retype to verify: ")
		password2, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return nil, errors.Wrap(err, "reading password")
		}
		fmt.Print("\n")

		if !bytes.Equal(pwd.Bytes(), bytes.TrimSpace(password2)) {
			return nil, errors.New("passwords must be equal")
		}
	}

	// Seal destroys the buffer
	return pwd.Seal(), nil
}

// GetMasterPassword takes the user master password from the config or requests it.
func GetMasterPassword() (*memguard.Enclave, error) {
	password := viper.Get("user.password")

	v := reflect.ValueOf(password)
	switch v.Kind() {
	case reflect.Ptr:
		pwd, ok := password.(*memguard.Enclave)
		if !ok {
			return nil, errors.New("the struct must be an enclave")
		}

		return pwd, nil

	default:
		pwd, err := AskPassword(false)
		if err != nil {
			return nil, err
		}

		viper.Set("user.password", pwd)

		return pwd, nil
	}
}
