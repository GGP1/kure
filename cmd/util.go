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
	"github.com/GGP1/kure/db/wallet"
	"github.com/GGP1/kure/pb"
	"github.com/awnumar/memguard"

	"github.com/pkg/errors"
	"github.com/skip2/go-qrcode"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	bolt "go.etcd.io/bbolt"
)

var errInvalidMasterPassword = errors.New("error: invalid master password")

// RunEFunc runs a cobra function returning an error.
type RunEFunc func(cmd *cobra.Command, args []string) error

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

// GetConfigPath returns the path to the config file.
func GetConfigPath() (string, error) {
	var path string
	cfgPath := os.Getenv("KURE_CONFIG")

	if cfgPath != "" {
		base := filepath.Base(cfgPath)
		if strings.Contains(base, ".") {
			return cfgPath, nil
		}

		path = filepath.Join(cfgPath, ".kure.yaml")
		return path, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", errors.Errorf("couldn't find user home directory: %v", err)
	}

	path = filepath.Join(filepath.Clean(home), ".kure.yaml")

	return path, nil
}

// PrintObjectName prints the name of the object adapting
// the top line of the chart so it stays always in the center.
//
// Example: +────── Name ──────>.
func PrintObjectName(name string) {
	// Do not use folders as part of the name
	if strings.Contains(name, "/") {
		split := strings.Split(name, "/")
		name = split[len(split)-1]
	}

	if !strings.Contains(name, ".") {
		name = strings.Title(name)
	}

	dashes := 57
	halfBar := ((dashes - len([]rune(name))) / 2) - 1

	// Top left
	fmt.Printf("\n┌%s", strings.Repeat("─", halfBar))

	// Name
	fmt.Printf(" %s ", name)
	if (len(name) % 2) == 0 {
		halfBar++
	}

	// Top right
	fmt.Printf("%s>\n", strings.Repeat("─", halfBar))
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
//
// This function is not tested as stubbing terminal.ReadPassword() provides
// almost no benefits.
func RequirePassword(db *bolt.DB) error {
	tx, err := db.Begin(false)
	if err != nil {
		return errors.Wrap(err, "transaction failed")
	}

	// Get each bucket number of records and close the transaction
	cards := tx.Bucket([]byte("kure_card")).Stats().KeyN
	entries := tx.Bucket([]byte("kure_entry")).Stats().KeyN
	files := tx.Bucket([]byte("kure_file")).Stats().KeyN
	wallets := tx.Bucket([]byte("kure_wallet")).Stats().KeyN
	tx.Rollback()

	// If it's the first one ask the user to verify the password
	if cards+entries+files+wallets == 0 {
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

	if cards > 0 {
		_, err := card.ListNames(db)
		if err != nil {
			return errInvalidMasterPassword
		}
		return nil
	}
	if entries > 0 {
		_, err := entry.ListNames(db)
		if err != nil {
			return errInvalidMasterPassword
		}
		return nil
	}
	if files > 0 {
		_, err := file.ListNames(db)
		if err != nil {
			return errInvalidMasterPassword
		}
		return nil
	}
	if wallets > 0 {
		_, err := wallet.ListNames(db)
		if err != nil {
			return errInvalidMasterPassword
		}
		return nil
	}

	return nil
}

// Scan scans a single line and returns the input.
func Scan(scanner *bufio.Scanner, field string) string {
	fmt.Printf("%s: ", field)

	scanner.Scan()
	text := scanner.Text()

	return strings.TrimSpace(text)
}

// Scanlns scans multiple lines and returns the input.
func Scanlns(scanner *bufio.Scanner, field string) string {
	fmt.Printf("%s (type !end to finish): ", field)

	var text []string
	for scanner.Scan() {
		t := scanner.Text()

		if t == "!end" {
			break
		}

		text = append(text, t)
	}
	return strings.Join(text, "\n")
}

// SetContext sets up the testing environement.
func SetContext(t *testing.T, path string) *bolt.DB {
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		t.Fatalf("Failed connecting to the database: %v", err)
	}

	viper.Reset()
	password := memguard.NewBufferFromBytes([]byte("test"))
	defer password.Destroy()
	viper.Set("user.password", password.Seal())

	// Reduce argon2id parameters to speed up tests
	viper.Set("argon2id.memory", 1024)
	viper.Set("argon2id.iterations", 1)

	err = db.Update(func(tx *bolt.Tx) error {
		buckets := [4]string{"kure_card", "kure_entry", "kure_file", "kure_wallet"}
		for _, bucket := range buckets {
			tx.DeleteBucket([]byte(bucket))
			_, err := tx.CreateBucketIfNotExists([]byte(bucket))
			if err != nil {
				return errors.Wrapf(err, "couldn't create %q bucket", bucket)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	e := &pb.Entry{
		Name:    "May the force be with you",
		Expires: "Never",
	}

	if err := entry.Create(db, e); err != nil {
		t.Fatal(err)
	}

	return db
}
