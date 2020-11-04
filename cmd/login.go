package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/GGP1/kure/crypt"
	"github.com/GGP1/kure/db"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Set master password",
	Run: func(cmd *cobra.Command, args []string) {
		p := viper.GetString("user.password")
		if p != "" {
			fmt.Println("Warning: the password is already set, if you want to abort please use ctrl+c")
		}
		if err := setMasterPassword(); err != nil {
			fatal(err)
		}

		path := getConfigPath()

		if err := viper.WriteConfigAs(path); err != nil {
			fatalf("failed creating config file: %s", err)
		}

		fmt.Println("\nYou have successfully logged in.")
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}

// setMasterPassword asks the user for a password, hashes it with SHA-512,
// sets it in viper and verifies if it's capable of decrypting saved records,
// if not it will ask the user for a confirmation to proceed.
func setMasterPassword() error {
	password, err := crypt.AskPassword(false)
	if err != nil {
		return err
	}

	viper.Set("user.password", password)

	// Check if the hashed password provided is capable of decrypting saved records
	_, err = db.ListEntries()
	_, err = db.ListFiles()
	if err != nil {
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Print("\nAlready stored records were encrypted with a different password. Do you want to proceed? [y/N] ")

		scanner.Scan()
		text := scanner.Text()
		text = strings.ToLower(text)

		if !strings.Contains(text, "y") {
			os.Exit(0)
		}
	}

	return nil
}
