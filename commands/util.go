package cmdutil

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/GGP1/kure/config"
	dbutil "github.com/GGP1/kure/db"
	"github.com/GGP1/kure/db/card"
	"github.com/GGP1/kure/db/entry"
	"github.com/GGP1/kure/db/file"
	"github.com/GGP1/kure/db/totp"
	"github.com/GGP1/kure/orderedmap"
	"github.com/GGP1/kure/sig"

	"github.com/atotto/clipboard"
	"github.com/awnumar/memguard"
	"github.com/pkg/errors"
	"github.com/skip2/go-qrcode"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var (
	// ErrInvalidLength is returned when generating a password/passphrase and the length passed is < 1.
	ErrInvalidLength = errors.New("invalid length")
	// ErrInvalidName is returned when a name is required and received "" or contains "//".
	ErrInvalidName = errors.New("invalid name")
	// ErrInvalidPath is returned when a path is required and received "".
	ErrInvalidPath = errors.New("invalid path")
)

const (
	// Card object
	Card object = iota
	// Entry object
	Entry
	// File object
	File
	// TOTP object
	TOTP

	// Box
	hBar       = "─"
	vBar       = "│"
	upperLeft  = "╭"
	lowerLeft  = "╰"
	upperRight = "╮"
	lowerRight = "╯"
)

// RunEFunc runs a cobra function returning an error.
type RunEFunc func(cmd *cobra.Command, args []string) error

type object int

// BuildBox constructs a responsive box used to display records information.
//
// ┌──── Sample ────┐
// │ Key  │ Value   │
// └────────────────┘
func BuildBox(name string, mp *orderedmap.Map) string {
	var sb strings.Builder

	// Do not use folders as part of the name
	name = filepath.Base(name)
	if !strings.Contains(name, ".") {
		name = strings.Title(name)
	}

	nameLen := len([]rune(name))
	longestKey := 0
	longestValue := nameLen

	// Range to take the longest key and value
	// Keys will always be 1 byte characters
	// Values may be 1, 2 or 3 bytes, to take the length use len([]rune(v))
	for _, key := range mp.Keys() {
		value := mp.Get(key) // Get key's value

		// Take map's longest key
		if len(key) > longestKey {
			longestKey = len(key)
		}

		// Split each value by a new line (fields like Notes contain multiple lines)
		for _, v := range strings.Split(value, "\n") {
			lenV := len([]rune(v))

			// Take map's longest value
			if lenV > longestValue {
				longestValue = lenV
			}
		}
	}

	// -4-: 2 spaces that wrap name and 2 corners
	headerLen := longestKey + longestValue - nameLen + 4
	headerHalfLen := headerLen / 2

	// Left side header
	sb.WriteString(upperLeft)
	sb.WriteString(strings.Repeat(hBar, headerHalfLen))

	// Header name
	sb.WriteRune(' ')
	sb.WriteString(name)
	sb.WriteRune(' ')

	// Adjust the right side of the header if its width is even
	if headerLen%2 == 0 {
		headerHalfLen--
	}

	// Right side header
	sb.WriteString(strings.Repeat(hBar, headerHalfLen))
	sb.WriteString(upperRight)
	sb.WriteString("\n")

	// Body
	for _, key := range mp.Keys() {
		value := mp.Get(key) // Get key's value
		// Start
		sb.WriteString(vBar)

		// Key
		sb.WriteRune(' ')
		sb.WriteString(key)
		sb.WriteRune(' ')
		sb.WriteString(strings.Repeat(" ", longestKey-len(key))) // Padding

		// Middle
		sb.WriteString(vBar)

		// Value
		for i, v := range strings.Split(value, "\n") {
			// In case the value contains multi-lines,
			// repeat the process above but do not add the key
			if i >= 1 {
				sb.WriteString("\n")
				sb.WriteString(vBar)
				// -2- represents key leading and trailing spaces
				//   1   2
				// (│ key │), here key = ""
				sb.WriteString(strings.Repeat(" ", longestKey+2)) // Padding
				sb.WriteString(vBar)
			}

			sb.WriteRune(' ')
			sb.WriteString(v)
			sb.WriteString(strings.Repeat(" ", longestValue-len([]rune(v)))) // Padding

			// End
			sb.WriteString(" ")
			sb.WriteString(vBar)
		}
		sb.WriteString("\n")
	}

	// Footer
	// -5- represents the characters that wrap both key and value
	//  1   234     5
	// ( key │ value )
	footerLen := longestKey + longestValue + 5
	sb.WriteString(lowerLeft)
	sb.WriteString(strings.Repeat(hBar, footerLen))
	sb.WriteString(lowerRight)

	return sb.String()
}

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

	fmt.Println(qr.ToSmallString(false))
	return nil
}

// Erase overwrites the file content with random bytes and then deletes it.
func Erase(filename string) error {
	f, err := os.Stat(filename)
	if err != nil {
		return errors.Wrap(err, "obtaining file information")
	}

	buf := make([]byte, f.Size())
	if _, err := rand.Read(buf); err != nil {
		return errors.Wrap(err, "generating random bytes")
	}

	// WriteFile truncates the file and overwrites it
	if err := os.WriteFile(filename, buf, 0600); err != nil {
		return errors.Wrap(err, "overwriting file")
	}

	if err := os.Remove(filename); err != nil {
		return errors.Wrap(err, "removing file")
	}

	return nil
}

// Exists checks if name or one of its folders is already being used.
//
// Returns an error if a match was found.
func Exists(db *bolt.DB, name string, obj object) error {
	records, objType, err := listNames(db, obj)
	if err != nil {
		return err
	}

	return exists(records, name, objType)
}

// FmtExpires returns expires formatted.
func FmtExpires(expires string) (string, error) {
	switch strings.ToLower(expires) {
	case "never", "", " ", "0", "0s":
		return "Never", nil

	default:
		expires = strings.ReplaceAll(expires, "-", "/")

		// If the first format fails, try the second
		exp, err := time.Parse("02/01/2006", expires)
		if err != nil {
			exp, err = time.Parse("2006/01/02", expires)
			if err != nil {
				return "", errors.New("\"expires\" field has an invalid format. Valid formats: d/m/y or y/m/d")
			}
		}

		return exp.Format(time.RFC1123Z), nil
	}
}

// MustExist returns an error if a record does not exist or if the name is invalid.
func MustExist(db *bolt.DB, obj object, allowDir ...bool) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return ErrInvalidName
		}

		records, objType, err := listNames(db, obj)
		if err != nil {
			return err
		}

		for _, name := range args {
			if name == "" || strings.Contains(name, "//") {
				return ErrInvalidName
			}
			name = NormalizeName(name, allowDir...)

			if strings.HasSuffix(name, "/") {
				// Take directories into consideration only when the user
				// is trying to perform an action with one
				if err := exists(records, name, objType); err == nil {
					return errors.Errorf("directory %q does not exist", strings.TrimSuffix(name, "/"))
				}
				return nil
			}

			exists := false
			for _, record := range records {
				if name == record {
					exists = true
					break
				}
			}
			if !exists {
				return errors.Errorf("%q does not exist", name)
			}
		}

		return nil
	}
}

// MustExistLs is like MustExist but it doesn't fail if
// there are no arguments or if the user is using the filter flag.
func MustExistLs(db *bolt.DB, obj object) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 || cmd.Flags().Changed("filter") {
			return nil
		}

		// If an empty string is joined in session/it command
		// it returns a 1 item long slice [""]
		if strings.Join(args, "") == "" {
			return nil
		}

		// Pass on cmd and args
		return MustExist(db, obj)(cmd, args)
	}
}

// MustNotExist returns an error if the record exists or if the name is invalid.
func MustNotExist(db *bolt.DB, obj object, allowDir ...bool) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		for _, name := range args {
			if name == "" || strings.Contains(name, "//") {
				return ErrInvalidName
			}
			name = NormalizeName(name, allowDir...)

			if err := Exists(db, name, obj); err != nil {
				return err
			}
		}

		return nil
	}
}

// NormalizeName sanitizes the user input name.
func NormalizeName(name string, allowDir ...bool) string {
	if name == "" {
		return name // Avoid allocations
	}
	if len(allowDir) == 0 {
		return strings.ToLower(strings.TrimSpace(strings.Trim(strings.TrimSpace(name), "/")))
	}
	return strings.ToLower(strings.TrimSpace(name))
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

// SelectEditor returns the editor to use, if none is found it returns vim.
func SelectEditor() string {
	if def := config.GetString("editor"); def != "" {
		return def
	} else if e := os.Getenv("EDITOR"); e != "" {
		return e
	} else if v := os.Getenv("VISUAL"); v != "" {
		return v
	}

	return "vim"
}

// SetContext sets up the testing environment.
//
// It uses t.Cleanup() to close the database connection after the test and
// all its subtests are completed.
func SetContext(t testing.TB, path string) *bolt.DB {
	t.Helper()
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		t.Fatalf("Failed connecting to the database: %v", err)
	}

	config.Reset()
	// Reduce argon2 parameters to speed up tests
	auth := map[string]interface{}{
		"password":   memguard.NewEnclave([]byte("1")),
		"iterations": 1,
		"memory":     1,
		"threads":    1,
	}
	config.Set("auth", auth)

	db.Update(func(tx *bolt.Tx) error {
		buckets := [][]byte{dbutil.CardBucket, dbutil.EntryBucket, dbutil.FileBucket, dbutil.TOTPBucket}
		for _, bucket := range buckets {
			// Ignore errors on purpose
			tx.DeleteBucket(bucket)
			tx.CreateBucketIfNotExists(bucket)
		}
		return nil
	})

	os.Stdout = os.NewFile(0, "") // Mute stdout
	os.Stderr = os.NewFile(0, "") // Mute stderr
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("Failed closing database: %v", err)
		}
	})

	return db
}

// WatchFile looks for the file initial state and loops until the first modification.
//
// Preferred over fsnotify since this last returns false events with recently created files.
func WatchFile(filename string, done chan struct{}, errCh chan error) {
	initStat, err := os.Stat(filename)
	if err != nil {
		errCh <- err
		return
	}

	for {
		stat, err := os.Stat(filename)
		if err != nil {
			errCh <- err
			return
		}

		if stat.Size() != initStat.Size() || stat.ModTime() != initStat.ModTime() {
			break
		}

		time.Sleep(300 * time.Millisecond)
	}

	done <- struct{}{}
}

// WriteClipboard writes the content to the clipboard and deletes it after
// "t" if "t" is higher than 0 or if there is a default timeout set in the configuration.
// Otherwise it does nothing.
func WriteClipboard(cmd *cobra.Command, t time.Duration, field, content string) error {
	if err := clipboard.WriteAll(content); err != nil {
		return errors.Wrap(err, "writing to clipboard")
	}
	fmt.Println(field, "copied to clipboard")
	memguard.WipeBytes([]byte(content))

	// Use the config value if it's specified and the timeout flag wasn't used
	configKey := "clipboard.timeout"
	if config.IsSet(configKey) && !cmd.Flags().Changed("timeout") {
		t = config.GetDuration(configKey)
	}

	if t > 0 {
		sig.Signal.AddCleanup(func() error { return clipboard.WriteAll("") })
		<-time.After(t)
		clipboard.WriteAll("")
	}

	return nil
}

func exists(records []string, name, objType string) error {
	if len(records) == 0 {
		return nil
	}

	found := func(name string) error {
		return errors.Errorf("already exists a folder or %s named %q", objType, name)
	}
	// Remove slash to do the comparison
	name = strings.TrimSuffix(name, "/")

	for _, record := range records {
		if name == record {
			return found(name)
		}

		// record = "Padmé/Amidala", name = "Padmé/" should return an error
		if hasPrefix(record, name) {
			return found(name)
		}

		// name = "Padmé/Amidala", record = "Padmé/" should return an error
		if hasPrefix(name, record) {
			return found(record)
		}
	}

	return nil
}

// hasPrefix is a modified version of strings.HasPrefix() that suits this use case, prefix is not modified to save an allocation.
func hasPrefix(s, prefix string) bool {
	prefixLen := len(prefix)
	return len(s) > prefixLen && s[0:prefixLen] == prefix && s[prefixLen] == '/'
}

// listNames lists all the records depending on the object passed.
// It returns a list and the type of object used.
func listNames(db *bolt.DB, obj object) ([]string, string, error) {
	var (
		err     error
		objType string
		records []string
	)

	switch obj {
	case Card:
		objType = "card"
		records, err = card.ListNames(db)

	case Entry:
		objType = "entry"
		records, err = entry.ListNames(db)

	case File:
		objType = "file"
		records, err = file.ListNames(db)

	case TOTP:
		objType = "TOTP"
		records, err = totp.ListNames(db)
	}
	if err != nil {
		return nil, "", err
	}

	return records, objType, nil
}
