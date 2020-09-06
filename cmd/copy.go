package cmd

import (
	"fmt"
	"os"
	"syscall"
	"time"

	"github.com/GGP1/kure/crypt"
	"github.com/GGP1/kure/db"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	timeout time.Duration
	copyCmd = &cobra.Command{
		Use:   "copy",
		Short: "Copy password to clipboard",
		Run: func(cmd *cobra.Command, args []string) {
			entry, err := db.GetEntry(title)
			if err != nil {
				fmt.Println(err)
				return
			}

			if secure && entry.Secure {
				fmt.Print("Enter Password: ")
				pwd, err := terminal.ReadPassword(int(syscall.Stdin))
				if err != nil {
					fmt.Println("error:", err)
					return
				}

				decryptedPwd, err := crypt.Decrypt([]byte(entry.Password), pwd)
				if err != nil {
					fmt.Printf("\nerror: %v\n", err)
					return
				}

				entry.Password = decryptedPwd
			}

			if err := clipboard.WriteAll(string(entry.Password)); err != nil {
				fmt.Println("error:", err)
				return
			}

			if timeout > 0 {
				<-time.After(timeout)
				clipboard.WriteAll("")
				os.Exit(1)
			}
		},
	}
)

func init() {
	RootCmd.AddCommand(copyCmd)
	copyCmd.Flags().StringVarP(&title, "title", "t", "", "entry title")
	copyCmd.Flags().DurationVarP(&timeout, "clean", "c", 0, "clipboard cleaning timeout")
	copyCmd.Flags().BoolVarP(&secure, "secure", "S", false, "decrypt password before copying")
}
