package cmdutil

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/GGP1/kure/config"
	"github.com/GGP1/kure/db/bucket"
	"github.com/GGP1/kure/db/card"
	"github.com/GGP1/kure/db/entry"
	"github.com/GGP1/kure/db/file"
	"github.com/GGP1/kure/db/totp"
	"github.com/GGP1/kure/orderedmap"
	"github.com/GGP1/kure/sig"
	"github.com/GGP1/kure/terminal"

	"github.com/atotto/clipboard"
	"github.com/awnumar/memguard"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
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
//	┌──── Sample ────┐
//	│ Key  │ Value   │
//	└────────────────┘
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

// Erase overwrites the file content with random bytes and then deletes it.
func Erase(filename string) error {
	f, err := os.Stat(filename)
	if err != nil {
		return errors.Wrap(err, "obtaining file information")
	}

	buf := make([]byte, f.Size())
	_, _ = rand.Read(buf)

	// WriteFile truncates the file and overwrites it
	if err := os.WriteFile(filename, buf, 0o600); err != nil {
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
	return func(_ *cobra.Command, args []string) error {
		if len(args) == 0 {
			return ErrInvalidName
		}

		names, objType, err := listNames(db, obj)
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
				if err := exists(names, name, objType); err == nil {
					return errors.Errorf("directory %q does not exist", strings.TrimSuffix(name, "/"))
				}
				return nil
			}

			exists := false
			for _, record := range names {
				if name == record {
					exists = true
					break
				}
			}

			if !exists {
				suggestions := getNameSuggestions(name, names)
				if len(suggestions) == 0 {
					return errors.Errorf("%q does not exist", name)
				}

				return errors.Errorf("%q does not exist. Did you mean %s?",
					name,
					formatSuggestions(suggestions),
				)
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
		if len(args) == 0 {
			return ErrInvalidName
		}

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
func SetContext(t testing.TB) *bolt.DB {
	t.Helper()

	dbFile, err := os.CreateTemp("", "*")
	assert.NoError(t, err)

	db, err := bolt.Open(dbFile.Name(), 0o600, &bolt.Options{Timeout: 1 * time.Second})
	assert.NoError(t, err, "Failed connecting to the database")

	config.Reset()
	// Reduce argon2 parameters to speed up tests
	auth := map[string]interface{}{
		"password":   memguard.NewEnclave([]byte("1")),
		"iterations": 1,
		"memory":     1,
		"threads":    1,
		"key":        []byte("01234567890123456789012345678901"),
	}
	config.Set("auth", auth)

	db.Update(func(tx *bolt.Tx) error {
		buckets := bucket.GetNames()
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
		assert.NoError(t, db.Close(), "Failed connecting to the database")
	})

	return db
}

// SupportedManagers validates if the password manager used to import/export records is supported.
func SupportedManagers() cobra.PositionalArgs {
	return func(_ *cobra.Command, args []string) error {
		manager := strings.Join(args, " ")

		switch strings.ToLower(manager) {
		case "1password", "bitwarden", "keepass", "keepassx", "keepassxc", "lastpass":

		default:
			return errors.Errorf(`%q is not supported

Supported managers: 1Password, Bitwarden, Keepass/X/XC, Lastpass`, manager)
		}
		return nil
	}
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

// WriteClipboard writes the value to the clipboard and deletes it after
// "t" if "t" is higher than 0 or if there is a default timeout set in the configuration.
// Otherwise it does nothing.
func WriteClipboard(cmd *cobra.Command, d time.Duration, field, value string) error {
	if err := clipboard.WriteAll(value); err != nil {
		return errors.Wrap(err, "writing to clipboard")
	}
	memguard.WipeBytes([]byte(value))

	// Use the config value if it's specified and the timeout flag wasn't used
	configKey := "clipboard.timeout"
	if config.IsSet(configKey) && !cmd.Flags().Changed("timeout") {
		d = config.GetDuration(configKey)
	}

	if d <= 0 {
		fmt.Println(field, "copied to clipboard")
		return nil
	}

	sig.Signal.AddCleanup(func() error { return clipboard.WriteAll("") })
	done := make(chan struct{})
	start := time.Now()

	go terminal.Ticker(done, true, func() {
		timeLeft := d - time.Since(start)
		fmt.Printf("(%v) %s copied to clipboard", timeLeft.Round(time.Second), field)
	})

	<-time.After(d)
	done <- struct{}{}
	clipboard.WriteAll("")

	return nil
}

func exists(names []string, name, objType string) error {
	if len(names) == 0 {
		return nil
	}

	found := func(name string) error {
		return errors.Errorf("already exists a folder or %s named %q", objType, name)
	}
	// Remove slash to do the comparison
	name = strings.TrimSuffix(name, "/")

	for _, n := range names {
		if name == n {
			return found(name)
		}

		// n = "Padmé/Amidala", name = "Padmé/" should return an error
		if hasPrefix(n, name) {
			return found(name)
		}

		// name = "Padmé/Amidala", n = "Padmé/" should return an error
		if hasPrefix(name, n) {
			return found(n)
		}
	}

	return nil
}

func formatSuggestions(suggestions []string) string {
	suggestionsStr := ""
	for i, suggestion := range suggestions {
		if len(suggestions) != 1 && i == len(suggestions)-1 {
			suggestionsStr += " or "
		} else if i != 0 {
			suggestionsStr += ", "
		}
		suggestionsStr += "\"" + suggestion + "\""
	}

	return suggestionsStr
}

// getNameSuggestions returns a list of names that are similar to the one provided.
func getNameSuggestions(name string, names []string) []string {
	suggestions := make([]string, 0)
	for _, n := range names {
		levenshteinDistance := levenshteinDistance(name, n)
		if levenshteinDistance <= 2 || strings.HasPrefix(n, name) {
			suggestions = append(suggestions, n)
		}
	}
	return suggestions
}

// hasPrefix is a modified version of strings.HasPrefix() that suits our use case, prefix is not modified to save an allocation.
func hasPrefix(s, prefix string) bool {
	prefixLen := len(prefix)
	return len(s) > prefixLen && s[0:prefixLen] == prefix && s[prefixLen] == '/'
}

// levenshteinDistance compares two strings and returns the levenshtein distance between them.
func levenshteinDistance(s, t string) int {
	d := make([][]int, len(s)+1)
	for i := range d {
		d[i] = make([]int, len(t)+1)
		d[i][0] = i
	}
	for j := range d[0] {
		d[0][j] = j
	}
	for j := 1; j <= len(t); j++ {
		for i := 1; i <= len(s); i++ {
			if s[i-1] == t[j-1] {
				d[i][j] = d[i-1][j-1]
			} else {
				min := d[i-1][j]
				if d[i][j-1] < min {
					min = d[i][j-1]
				}
				if d[i-1][j-1] < min {
					min = d[i-1][j-1]
				}
				d[i][j] = min + 1
			}
		}
	}
	return d[len(s)][len(t)]
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
