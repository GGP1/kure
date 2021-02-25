package auth

import (
	"crypto/subtle"
	"fmt"
	"syscall"

	"github.com/GGP1/kure/sig"

	"github.com/awnumar/memguard"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh/terminal"
)

// ErrInvalidPassword is used to identify the error received when editing
// an entry.
var ErrInvalidPassword = errors.New("invalid password")

// AskPassword returns the input password encrypted inside an Enclave.
//
// This function could be tested by stubbing terminal.ReadPassword()
// or using Netflix/go-expect library but it provides almost no benefits.
func AskPassword(message string, verify bool) (*memguard.Enclave, error) {
	fd := int(syscall.Stdin)
	oldState, err := terminal.GetState(fd)
	if err != nil {
		return nil, errors.Wrap(err, "terminal state")
	}

	// Restore the terminal to its previous state
	sig.Signal.AddCleanup(func() error { return terminal.Restore(fd, oldState) })
	defer func() error { return terminal.Restore(fd, oldState) }()

	fmt.Print(message + ": ")
	password, err := terminal.ReadPassword(fd)
	if err != nil {
		return nil, errors.Wrap(err, "reading password")
	}
	fmt.Print("\n")

	if subtle.ConstantTimeCompare(password, nil) == 1 {
		return nil, ErrInvalidPassword
	}

	pwd := memguard.NewBufferFromBytes(password)
	memguard.WipeBytes(password)

	if verify {
		fmt.Print("Retype to verify: ")
		password2, err := terminal.ReadPassword(fd)
		if err != nil {
			return nil, errors.Wrap(err, "reading password")
		}
		fmt.Print("\n")

		if subtle.ConstantTimeCompare(pwd.Bytes(), password2) != 1 {
			return nil, errors.New("passwords didn't match")
		}
		memguard.WipeBytes(password2)
	}

	// Seal destroys the locked buffer
	return pwd.Seal(), nil
}
