package crypt

import (
	"crypto/subtle"
	"fmt"
	"syscall"

	"github.com/awnumar/memguard"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh/terminal"
)

// AskPassword returns the input password encrypted inside an Enclave.
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

	if subtle.ConstantTimeCompare(password, nil) == 1 {
		return nil, errors.New("invalid password")
	}

	pwd := memguard.NewBufferFromBytes(password)
	memguard.WipeBytes(password)

	if verify {
		fmt.Print("Retype to verify: ")
		password2, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return nil, errors.Wrap(err, "reading password")
		}
		fmt.Print("\n")

		if subtle.ConstantTimeCompare(pwd.Bytes(), password2) != 1 {
			return nil, errors.New("passwords must be equal")
		}
		memguard.WipeBytes(password2)
	}

	// Seal destroys the buffer
	return pwd.Seal(), nil
}

// GetMasterPassword takes the user master password from the config or requests it.
func GetMasterPassword() (*memguard.Enclave, error) {
	password := viper.Get("user.password")

	switch password.(type) {
	case *memguard.Enclave:
		return password.(*memguard.Enclave), nil

	default:
		pwd, err := AskPassword(false)
		if err != nil {
			return nil, err
		}

		viper.Set("user.password", pwd)

		return pwd, nil
	}
}
