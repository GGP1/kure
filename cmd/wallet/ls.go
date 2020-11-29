package wallet

import (
	"fmt"
	"strings"

	cmdutil "github.com/GGP1/kure/cmd"
	"github.com/GGP1/kure/db/wallet"
	"github.com/GGP1/kure/pb"
	"github.com/GGP1/kure/tree"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var filter, hide bool

var lsExample = `
* List a wallet
kure wallet ls walletName

* Filter wallets by name
kure wallet ls walletName -f

* List all wallets
kure wallet ls`

func lsSubCmd(db *bolt.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ls <name> [-f filter] [-H hide]",
		Short:   "List wallets",
		Example: lsExample,
		RunE:    runLs(db),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset flags defaults (session)
			filter, hide = false, false
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	f := cmd.Flags()
	f.BoolVarP(&filter, "filter", "f", false, "filter wallets")
	f.BoolVarP(&hide, "hide", "H", false, "hide wallet seed phrase and private key")

	return cmd
}

func runLs(db *bolt.DB) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")

		if name == "" {
			wallets, err := wallet.ListNames(db)
			if err != nil {
				return errors.Wrap(err, "error")
			}

			paths := make([]string, len(wallets))

			for i, wallet := range wallets {
				paths[i] = wallet.Name
			}

			tree.Print(paths)
			return nil
		}

		if filter {
			wallets, err := wallet.ListByName(db, name)
			if err != nil {
				return errors.Wrap(err, "error")
			}

			if len(wallets) == 0 {
				return errors.New("error: no wallets were found")
			}

			for _, w := range wallets {
				printWallet(w)
			}
			return nil
		}

		w, err := wallet.Get(db, name)
		if err != nil {
			return errors.Wrap(err, "error")
		}

		printWallet(w)
		return nil
	}
}

func printWallet(w *pb.Wallet) {
	cmdutil.PrintObjectName(w.Name)

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

	fmt.Println("└────────────────+────────────────────────────────────────>")
}
