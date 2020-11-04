package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/GGP1/kure/db"
	"github.com/GGP1/kure/pb"
	"github.com/GGP1/kure/tree"

	"github.com/atotto/clipboard"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var walletCmd = &cobra.Command{
	Use:   "wallet",
	Short: "Wallet operations",
}

var addWallet = &cobra.Command{
	Use:   "add <name>",
	Short: "Add a wallet to the database",
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

		fmt.Printf("\nSuccessfully created \"%s\" wallet.\n", name)
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
			fatal(errWritingClipboard)
		}

		if timeout > 0 {
			<-time.After(timeout)
			clipboard.WriteAll("")
			os.Exit(1)
		}
	},
}

var lsWallet = &cobra.Command{
	Use:   "ls <name> [-f filter]",
	Short: "List wallets",
	Run: func(cmd *cobra.Command, args []string) {
		name := strings.Join(args, " ")

		if name == "" {
			wallets, err := db.ListWallets()
			if err != nil {
				fatal(err)
			}

			paths := make([]string, len(wallets))

			for i, wallet := range wallets {
				paths[i] = wallet.Name
			}

			tree.Print(paths)
			return
		}

		if filter {
			wallets, err := db.WalletsByName(name)
			if err != nil {
				fatal(err)
			}
			for _, wallet := range wallets {
				fmt.Print("\n")
				printWallet(wallet)
			}
			return
		}

		wallet, err := db.GetWallet(name)
		if err != nil {
			fatal(err)
		}

		fmt.Print("\n")
		printWallet(wallet)
	},
}

var rmwallet = &cobra.Command{
	Use:   "rm <name>",
	Short: "Remove a wallet from the database",
	Run: func(cmd *cobra.Command, args []string) {
		name := strings.Join(args, " ")
		if name == "" {
			fatal(errInvalidName)
		}

		_, err := db.GetWallet(name)
		if err != nil {
			fatalf("%s wallet does not exist", name)
		}

		if proceed() {
			if err := db.RemoveWallet(name); err != nil {
				fatal(err)
			}

			fmt.Printf("\nSuccessfully removed \"%s\" wallet.\n", name)
		}
	},
}

func init() {
	rootCmd.AddCommand(walletCmd)

	walletCmd.AddCommand(addWallet)
	walletCmd.AddCommand(copyWallet)
	walletCmd.AddCommand(lsWallet)
	walletCmd.AddCommand(rmwallet)

	copyWallet.Flags().DurationVarP(&timeout, "timeout", "t", 0, "clipboard cleaning timeout")

	lsWallet.Flags().BoolVarP(&filter, "filter", "f", false, "filter wallets")
}

func walletInput(name string) (*pb.Wallet, error) {
	var wType, scriptT, keystoreT, seedPhrase, publicKey, privateKey string

	if !strings.Contains(name, "/") && len(name) > 43 {
		return nil, errors.New("wallet name must contain 43 letters or less")
	}

	scanner := bufio.NewScanner(os.Stdin)

	scan(scanner, "Type", &wType)
	scan(scanner, "Script type", &scriptT)
	scan(scanner, "Keystore Type", &keystoreT)
	scan(scanner, "Seed Phrase", &seedPhrase)
	scan(scanner, "Public Key", &publicKey)
	scan(scanner, "Private Key", &privateKey)

	name = strings.ToLower(name)

	wallet := &pb.Wallet{
		Name:         name,
		Type:         wType,
		ScriptType:   scriptT,
		KeystoreType: keystoreT,
		SeedPhrase:   seedPhrase,
		PublicKey:    publicKey,
		PrivateKey:   privateKey,
	}

	return wallet, nil
}

func printWallet(w *pb.Wallet) {
	printObjectName(w.Name)

	if hide {
		w.SeedPhrase = "•••••••••••••••"
		w.PrivateKey = "•••••••••••••••"
	}

	fmt.Printf(`│ Type      	 │ %s
│ Script Type    │ %s
│ Keystore Type  │ %s
│ Seed Phrase    │ %s
│ Public Key     │ %s
│ Private Key    │ %s
`, w.Type, w.ScriptType, w.KeystoreType, w.SeedPhrase, w.PublicKey, w.PrivateKey)

	fmt.Println("+─────────────+─────────────────────────────>")
}
