package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/GGP1/kure/crypt"
	"github.com/GGP1/kure/db/card"
	"github.com/GGP1/kure/db/entry"
	"github.com/GGP1/kure/db/file"
	"github.com/GGP1/kure/db/note"
	"github.com/GGP1/kure/orderedmap"
	"github.com/GGP1/kure/pb"
	"github.com/awnumar/memguard"

	"github.com/pkg/errors"
	"github.com/skip2/go-qrcode"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	bolt "go.etcd.io/bbolt"
)

var errInvalidMasterPassword = errors.New("invalid master password")

const (
	hBar   = "─"
	vBar   = "│"
	tLeft  = "┌"
	bLeft  = "└"
	tRight = "┐"
	bRight = "┘"
)

// RunEFunc runs a cobra function returning an error.
type RunEFunc func(cmd *cobra.Command, args []string) error

// BuildBox constructs an object box used to display its information.
//
// ┌──── Sample ────┐
// │ Key  │ Value   │
// └────────────────┘
func BuildBox(name string, mp *orderedmap.Map) (*memguard.LockedBuffer, string) {
	defer mp.Destroy() // Destroy map's key/value pairs after finishing

	lockedBuf, sb := pb.SecureStringBuilder()
	// Do not use folders as part of the name
	name = filepath.Base(name)

	if !strings.Contains(name, ".") {
		name = strings.Title(name)
	}

	lenName := len([]rune(name))
	longestKey := 0
	longestValue := lenName

	// Range to take the longest key and value
	// Keys will always be 1 byte characters
	// Values may be 1, 2 or 3 bytes, to take the length use len([]rune(v))
	for _, k := range mp.Keys() {
		v := mp.Get(k) // Get key's value

		// Take map's longest key
		if len(k) > longestKey {
			longestKey = len(k)
		}

		// Split each value by a new line (fields like Notes contain multiple lines)
		values := strings.Split(v, "\n")
		for _, v := range values {
			lenV := len([]rune(v))

			// Take map's longest value
			if lenV > longestValue {
				longestValue = lenV
			}
		}
	}

	// lenHeader is smaller as the name has more characters.
	// By default (name and fields empty) the header has 4 characters and the body 8,
	// the number -4- solves that difference.
	lenHeader := longestKey + longestValue + 4 - lenName
	half := lenHeader / 2

	// Left side header
	sb.WriteString(tLeft)
	sb.WriteString(strings.Repeat(hBar, half))

	// Header name
	sb.WriteString(fmt.Sprintf(" %s ", name))

	// Adjust the right side of the header if the width is odd
	if lenHeader%2 == 0 {
		half--
	}

	// Right side header
	sb.WriteString(strings.Repeat(hBar, half))
	sb.WriteString(tRight)
	sb.WriteString("\n")

	// Body
	for _, k := range mp.Keys() {
		v := mp.Get(k) // Get key's value
		// Start
		sb.WriteString(vBar)

		// Key
		sb.WriteString(fmt.Sprintf(" %s ", k))
		sb.WriteString(strings.Repeat(" ", longestKey-len(k))) // Padding

		// Middle
		sb.WriteString(vBar)

		// Value
		for i, value := range strings.Split(v, "\n") {
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

			sb.WriteString(fmt.Sprintf(" %s", value))
			sb.WriteString(strings.Repeat(" ", longestValue-len([]rune(value)))) // Padding

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
	lenFooter := longestKey + longestValue + 5
	sb.WriteString(bLeft)
	sb.WriteString(strings.Repeat(hBar, lenFooter))
	sb.WriteString(bRight)

	return lockedBuf, sb.String()
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

// Exists checks if there are records with the same name. It looks for matches
// on the same level, returns an error if a record already has the name passed.
//
// Given a path "Naboo/Padmé" and comparing it with "Naboo/Padmé Amidala":
//
// "Padmé" != "Padmé Amidala", skip.
//
// Given a path "jedi/Yoda" and comparing it with "jedi/Obi-Wan Kenobi":
//
// "jedi/Obi-Wan Kenobi" does not contain "jedi/Yoda", skip.
func Exists(db *bolt.DB, name, objectType string) error {
	var (
		records []string
		err     error
	)

	switch objectType {
	case "card":
		records, err = card.ListNames(db)

	case "entry":
		records, err = entry.ListNames(db)

	case "file":
		records, err = file.ListNames(db)

	case "note":
		records, err = note.ListNames(db)

	default:
		return errors.Errorf("%q is not a Kure's object", objectType)
	}
	if err != nil {
		return err
	}

	parts := strings.Split(name, "/")
	n := len(parts) - 1  // record name index
	basename := parts[n] // name without folders

	for _, r := range records {
		if strings.Contains(r, name) {
			rName := strings.Split(r, "/")[n]

			if rName == basename {
				return errors.Errorf("already exists a record or folder named %q", name)
			}
		}
	}

	return nil
}

// GetConfigPath returns the path to the config file.
func GetConfigPath() (string, error) {
	path := os.Getenv("KURE_CONFIG")
	if path != "" {
		base := filepath.Base(path)
		if filepath.Ext(base) != "" {
			return path, nil
		}

		path += ".yaml"
		return path, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", errors.Errorf("couldn't find user home directory: %v", err)
	}

	home = filepath.Join(filepath.Clean(home), ".kure.yaml")

	return home, nil
}

// Proceed asks the user if he wants to continue or not.
func Proceed(r io.Reader) bool {
	scanner := bufio.NewScanner(r)
	fmt.Print("Are you sure you want to proceed? [y/N]: ")

	scanner.Scan()
	text := scanner.Text()
	text = strings.ToLower(text)

	if strings.Contains(text, "y") {
		return true
	}

	return false
}

// RequirePassword verifies that the person that is trying to execute
// a command is effectively the owner.
//
// If it's the first records it asks the user twice to avoid miswriting.
func RequirePassword(db *bolt.DB) RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		tx, err := db.Begin(false)
		if err != nil {
			return errors.Wrap(err, "transaction failed")
		}

		// Get each bucket number of records and close the transaction
		cards := tx.Bucket([]byte("kure_card")).Stats().KeyN
		entries := tx.Bucket([]byte("kure_entry")).Stats().KeyN
		files := tx.Bucket([]byte("kure_file")).Stats().KeyN
		notes := tx.Bucket([]byte("kure_note")).Stats().KeyN
		tx.Rollback()

		// If it's the first one ask the user to verify the password
		if cards+entries+files+notes == 0 && viper.Get("user.password") == nil {
			password, err := crypt.AskPassword(true)
			if err != nil {
				return err
			}
			viper.Set("user.password", password)
			return nil
		}

		_, err = crypt.GetMasterPassword()
		if err != nil {
			return err
		}

		// Fetch only one record (the fastest) to decrypt it and make sure the password is correct
		if cards > 0 {
			if !card.ListFastest(db) {
				return errInvalidMasterPassword
			}
			return nil
		}
		if entries > 0 {
			if !entry.ListFastest(db) {
				return errInvalidMasterPassword
			}
			return nil
		}
		if files > 0 {
			if !file.ListFastest(db) {
				return errInvalidMasterPassword
			}
			return nil
		}
		if notes > 0 {
			if !note.ListFastest(db) {
				return errInvalidMasterPassword
			}
			return nil
		}

		return nil
	}
}

// Scan scans a single line and returns the input.
func Scan(scanner *bufio.Scanner, field string) string {
	fmt.Printf("%s: ", field)
	scanner.Scan()
	return strings.TrimSpace(scanner.Text())
}

// Scanlns scans multiple lines and returns the input.
func Scanlns(scanner *bufio.Scanner, field string) string {
	fmt.Printf("%s (type !end to finish): ", field)

	var lines []string
	for scanner.Scan() {
		t := scanner.Text()

		lines = append(lines, t)

		// We could break before appending but that would force
		// users to insert a new line to type it
		if strings.Contains(t, "!end") {
			break
		}
	}

	text := strings.Join(lines, "\n")
	text = strings.ReplaceAll(text, "!end", "")
	defer memguard.WipeBytes([]byte(text))

	return strings.TrimSpace(text)
}

// SetContext sets up the testing environement.
func SetContext(t *testing.T, path string) *bolt.DB {
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		t.Fatalf("Failed connecting to the database: %v", err)
	}

	viper.Reset()
	viper.Set("user.password", memguard.NewEnclave([]byte("test")))
	// Reduce argon2 parameters to speed up tests
	viper.Set("argon2.memory", 1)
	viper.Set("argon2.iterations", 1)

	err = db.Update(func(tx *bolt.Tx) error {
		buckets := [4]string{"kure_card", "kure_entry", "kure_file", "kure_note"}
		for _, bucket := range buckets {
			tx.DeleteBucket([]byte(bucket)) // Ignore error on purpose
			if _, err := tx.CreateBucketIfNotExists([]byte(bucket)); err != nil {
				return errors.Wrapf(err, "couldn't create %q bucket", bucket)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	return db
}
