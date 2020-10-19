package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/GGP1/kure/db"
	"github.com/GGP1/kure/model/wallet"

	"github.com/atotto/clipboard"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var errInvalidName = errors.New("please specify a name")

var walletCmd = &cobra.Command{
	Use:   "wallet",
	Short: "Wallet operations",
}

var addWallet = &cobra.Command{
	Use:   "add <name>",
	Short: "Add a wallet to the database.",
	Run: func(cmd *cobra.Command, args []string) {
		name := strings.Join(args, " ")
		if name == "" {
			fatal(errInvalidName)
		}

		wallet, err := walletInput(name)
		if err != nil {
			fatal(err)
		}

		if err := db.CreateWallet(wallet); err != nil {
			fatal(err)
		}

		fmt.Print("\nSucessfully created the wallet.")
	},
}

var copyWallet = &cobra.Command{
	Use:   "copy <name> [-t timeout]",
	Short: "Copy wallet public key",
	Run: func(cmd *cobra.Command, args []string) {
		name := strings.Join(args, " ")
		if name == "" {
			fatal(errInvalidName)
		}

		wallet, err := db.GetWallet(name)
		if err != nil {
			fatal(err)
		}

		if err := clipboard.WriteAll(wallet.PublicKey); err != nil {
			fatal(errors.New("failed writing to the clipboard"))
		}

		if timeout > 0 {
			<-time.After(timeout)
			clipboard.WriteAll("")
			os.Exit(1)
		}
	},
}

var deleteWallet = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a wallet from the database",
	Run: func(cmd *cobra.Command, args []string) {
		name := strings.Join(args, " ")
		if name == "" {
			fatal(errInvalidName)
		}

		_, err := db.GetWallet(name)
		if err != nil {
			fatalf("%s wallet does not exist", name)
		}

		scanner := bufio.NewScanner(os.Stdin)
		fmt.Print("Are you sure you want to proceed? [y/n]: ")

		scanner.Scan()
		text := scanner.Text()
		input := strings.ToLower(text)

		if strings.Contains(input, "y") {
			if err := db.DeleteWallet(name); err != nil {
				fatal(err)
			}

			fmt.Printf("\nSuccessfully deleted %s wallet.", name)
		}
	},
}

var listWallet = &cobra.Command{
	Use:   "list <name>",
	Short: "List a wallet or all the wallets from the database",
	Run: func(cmd *cobra.Command, args []string) {
		name := strings.Join(args, " ")

		if name != "" {
			wallet, err := db.GetWallet(name)
			if err != nil {
				fatal(err)
			}

			fmt.Print("\n")
			printWallet(wallet)
			return
		}

		wallets, err := db.ListWallets()
		if err != nil {
			fatal(err)
		}

		for _, wallet := range wallets {
			fmt.Print("\n")
			printWallet(wallet)
		}
	},
}

func init() {
	rootCmd.AddCommand(walletCmd)

	walletCmd.AddCommand(addWallet)
	walletCmd.AddCommand(copyWallet)
	walletCmd.AddCommand(deleteWallet)
	walletCmd.AddCommand(listWallet)

	copyWallet.Flags().DurationVarP(&timeout, "timeout", "t", 0, "clipboard cleaning timeout")
}

func walletInput(name string) (*wallet.Wallet, error) {
	var wType, scriptT, keystoreT, seedPhrase, publicKey, privateKey string

	if len(name) > 43 {
		return nil, errors.New("card name must contain 43 letters or less")
	}

	scanner := bufio.NewScanner(os.Stdin)

	scan(scanner, "Type", &wType)
	scan(scanner, "Script type", &scriptT)
	scan(scanner, "Keystore Type", &keystoreT)
	scan(scanner, "Seed Phrase", &seedPhrase)
	scan(scanner, "Public Key", &publicKey)
	scan(scanner, "Private Key", &privateKey)

	name = strings.TrimSpace(strings.ToLower(name))

	wallet := wallet.New(name, wType, scriptT, keystoreT, seedPhrase, publicKey, privateKey)

	return wallet, nil
}

func printWallet(w *wallet.Wallet) {
	w.Name = strings.Title(w.Name)

	if hide {
		w.SeedPhrase = "•••••••••••••••"
		w.PrivateKey = "•••••••••••••••"
	}

	dashes := 43
	halfBar := ((dashes - len(w.Name)) / 2) - 1

	fmt.Print("+")
	for i := 0; i < halfBar; i++ {
		fmt.Print("─")
	}
	fmt.Printf(" %s ", w.Name)

	if (len(w.Name) % 2) == 0 {
		halfBar++
	}

	for i := 0; i < halfBar; i++ {
		fmt.Print("─")
	}
	fmt.Print(">\n")

	if w.Type != "" {
		fmt.Printf("│ Type      	 │ %s\n", w.Type)
	}

	if w.ScriptType != "" {
		fmt.Printf("│ Script Type    │ %s\n", w.ScriptType)
	}

	if w.KeystoreType != "" {
		fmt.Printf("│ Keystore Type  │ %s\n", w.KeystoreType)
	}

	if w.SeedPhrase != "" {
		fmt.Printf("│ Seed Phrase    │ %s\n", w.SeedPhrase)
	}

	if w.PublicKey != "" {
		fmt.Printf("│ Public Key     │ %s\n", w.PublicKey)
	}

	if w.PrivateKey != "" {
		fmt.Printf("│ Private Key    │ %s\n", w.PrivateKey)
	}

	fmt.Println("+─────────────+─────────────────────────────>")
}
