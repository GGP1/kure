package cmd

import (
	"bufio"
	"crypto/sha512"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/GGP1/kure/db"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh/terminal"
)

var errCreatingConfig = "failed creating config file"

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Set master password",
	Run: func(cmd *cobra.Command, args []string) {
		if err := setMasterPassword(); err != nil {
			fatal(err)
		}

		configPath := os.Getenv("KURE_CONFIG")

		if configPath != "" {
			filename := fmt.Sprintf("%s/config.yaml", configPath)

			if err := viper.WriteConfigAs(filename); err != nil {
				fatalf(errCreatingConfig, err)
			}
		} else {
			home, err := os.UserHomeDir()
			if err != nil {
				fatalf("failed fetching home directory: %v", err)
			}

			filename := fmt.Sprintf("%s/config.yaml", home)

			if err := viper.WriteConfigAs(filename); err != nil {
				fatalf(errCreatingConfig, err)
			}
		}

		fmt.Println("\nYou have successfully logged in")
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}

// setMasterPassword asks the user for a password, hashes it with SHA-512,
// sets it in viper and verifies if it's capable of decrypting past records,
// if not it will ask the user for a confirmation to proceed.
func setMasterPassword() error {
	fmt.Print("Enter master password: ")
	password, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return errors.Wrap(err, "reading password")
	}

	p := strings.TrimSpace(string(password))
	h := sha512.New()

	_, err = h.Write([]byte(p))
	if err != nil {
		errors.Wrap(err, "password hash")
	}

	pwd := fmt.Sprintf("%x", h.Sum(nil))
	viper.Set("user.password", pwd)

	// Check if the hashed password provided is capable of decrypting saved entries
	_, err = db.ListEntries()
	if err != nil {
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Print("\nAlready stored records were encrypted with a different password. Do you want to proceed? [y/n] ")

		scanner.Scan()
		text := scanner.Text()
		input := strings.ToLower(text)

		if !strings.Contains(input, "y") {
			os.Exit(0)
		}
	}

	return nil
}
