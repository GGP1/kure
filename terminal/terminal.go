// Package terminal implements I/O operations used to get and display information to the user.
package terminal

import (
	"bufio"
	"bytes"
	"crypto/subtle"
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/GGP1/kure/sig"
	"github.com/awnumar/memguard"

	"github.com/pkg/errors"
	"github.com/skip2/go-qrcode"
	"golang.org/x/crypto/ssh/terminal"
)

const (
	// ANSI escape codes
	// https://en.wikipedia.org/wiki/ANSI_escape_code
	saveCursorPos = "\033[s"
	clearLine     = "\033[u\033[K"
	showCursor    = "\x1b[?25h"
	hideCursor    = "\x1b[?25l"
)

// ErrInvalidPassword is used to identify the error received when editing
// an entry.
var ErrInvalidPassword = errors.New("invalid password")

// Confirm requests the user for a yes/no response.
func Confirm(r io.Reader, message string) bool {
	fmt.Print(message, " [y/N] ")

	for {
		var res string
		// Scanln returns an error when the input is an empty string,
		// we do accept it here
		if _, err := fmt.Fscanln(r, &res); err != nil && res != "" {
			if err != io.EOF {
				fmt.Fprintln(os.Stderr, "error:", err)
			}
			sig.Signal.Kill()
		}

		switch res {
		case "Yes", "yes", "Y", "y":
			return true

		case "No", "no", "N", "n":
			return false

		default:
			fmt.Print("Invalid response, retry. [y/N] ")
		}
	}
}

// DisplayQRCode creates a qr code with the password provided and writes it to the terminal.
func DisplayQRCode(secret string) error {
	if len([]rune(secret)) > 1273 {
		return errors.New("secret too long to encode to QR code, maximum is 1273")
	}

	qr, err := qrcode.New(secret, qrcode.Highest)
	if err != nil {
		return errors.Wrap(err, "creating QR code")
	}

	fmt.Print(qr.ToSmallString(false))
	return nil
}

// Scanln scans a single line and returns the input.
func Scanln(r *bufio.Reader, field string) string {
	fmt.Printf("%s: ", field)

	text, _, err := r.ReadLine()
	if err != nil {
		if err != io.EOF {
			fmt.Fprintln(os.Stderr, "error:", err)
		}
		sig.Signal.Kill()
	}
	text = bytes.ReplaceAll(text, []byte("\t"), []byte(""))

	return strings.TrimSpace(string(text))
}

// Scanlns scans multiple lines and returns the input.
func Scanlns(r *bufio.Reader, field string) string {
	fmt.Print(field, " (type < to finish): ")

	text, err := r.ReadString('<')
	if err != nil {
		if err != io.EOF {
			fmt.Fprintln(os.Stderr, "error:", err)
		}
		sig.Signal.Kill()
	}

	text = strings.TrimSuffix(text, "<")
	text = strings.ReplaceAll(text, "\r", "")
	text = strings.ReplaceAll(text, "\t", "")
	return strings.TrimSpace(text)
}

// ScanPassword returns the input password encrypted inside an Enclave.
func ScanPassword(message string, verify bool) (*memguard.Enclave, error) {
	fd := int(syscall.Stdin)
	oldState, err := terminal.GetState(fd)
	if err != nil {
		return nil, errors.Wrap(err, "terminal state")
	}

	// Restore the terminal to its previous state
	sig.Signal.AddCleanup(func() error { return terminal.Restore(fd, oldState) })
	defer terminal.Restore(fd, oldState)

	fmt.Fprint(os.Stderr, message+": ")
	password, err := terminal.ReadPassword(fd)
	if err != nil {
		return nil, errors.Wrap(err, "reading password")
	}
	fmt.Fprint(os.Stderr, "\n")

	if subtle.ConstantTimeCompare(password, nil) == 1 {
		return nil, ErrInvalidPassword
	}

	pwd := memguard.NewBufferFromBytes(password)
	memguard.WipeBytes(password)

	if verify {
		fmt.Fprint(os.Stderr, "Retype to verify: ")
		password2, err := terminal.ReadPassword(fd)
		if err != nil {
			return nil, errors.Wrap(err, "reading password")
		}
		fmt.Fprint(os.Stderr, "\n")

		if subtle.ConstantTimeCompare(pwd.Bytes(), password2) != 1 {
			return nil, errors.New("passwords didn't match")
		}
		memguard.WipeBytes(password2)
	}

	// Seal destroys the locked buffer
	return pwd.Seal(), nil
}

// Ticker clears the terminal and executes the log function every second.
func Ticker(done chan struct{}, hiddenCursor bool, log func()) {
	fmt.Print(saveCursorPos)
	if hiddenCursor {
		fmt.Print(hideCursor)
		sig.Signal.AddCleanup(func() error {
			fmt.Print(showCursor)
			return nil
		})
	}
	log()

	ticker := time.NewTicker(time.Second)

	for {
		select {
		case <-done:
			fmt.Print(clearLine)
			if hiddenCursor {
				fmt.Print(showCursor)
			}
			ticker.Stop()
			return

		case <-ticker.C:
			fmt.Print(clearLine)
			log()
		}
	}
}
