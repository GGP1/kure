package wallet

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	cmdutil "github.com/GGP1/kure/cmd"
	"github.com/GGP1/kure/db/wallet"
	"github.com/GGP1/kure/pb"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var errInvalidPath = errors.New("invalid path")

var addExample = `
* Add a new wallet
kure wallet add walletName`

func addSubCmd(db *bolt.DB, r io.Reader) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add <name>",
		Short:   "Add a wallet to the database",
		Aliases: []string{"a", "new", "create"},
		Example: addExample,
		RunE:    runAdd(db, r),
	}

	return cmd
}

func runAdd(db *bolt.DB, r io.Reader) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")
		if name == "" {
			return errInvalidName
		}
		if !strings.Contains(name, "/") && len(name) > 57 {
			return errors.New("wallet name must contain 57 letters or less")
		}

		if err := cmdutil.RequirePassword(db); err != nil {
			return err
		}

		w, err := walletInput(db, name, r)
		if err != nil {
			return err
		}

		if err := wallet.Create(db, w); err != nil {
			return err
		}

		fmt.Printf("\nSuccessfully created %q wallet.\n", name)
		return nil
	}
}

func walletInput(db *bolt.DB, name string, r io.Reader) (*pb.Wallet, error) {
	name = strings.ToLower(name)

	if err := exists(db, name); err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(r)
	wType := cmdutil.Scan(scanner, "Type")
	scriptT := cmdutil.Scan(scanner, "Script Type")
	keystoreT := cmdutil.Scan(scanner, "Keystore Type")
	seedPhrase := cmdutil.Scan(scanner, "Seed Phrase")
	publicKey := cmdutil.Scan(scanner, "Public Key")
	privateKey := cmdutil.Scan(scanner, "Private Key")

	w := &pb.Wallet{
		Name:         name,
		Type:         wType,
		ScriptType:   scriptT,
		KeystoreType: keystoreT,
		SeedPhrase:   seedPhrase,
		PublicKey:    publicKey,
		PrivateKey:   privateKey,
	}

	return w, nil
}

// exists compares "path" with the existing wallet names that contain "path",
// split both name and each of the wallet names and look for matches on the same level.
//
// Given a path "Naboo/Padmé" and comparing it with "Naboo/Padmé Amidala":
//
// "Padmé" != "Padmé Amidala", skip.
//
// Given a path "jedi/Yoda" and comparing it with "jedi/Obi-Wan Kenobi":
//
// "jedi/Obi-Wan Kenobi" does not contain "jedi/Yoda", skip.
func exists(db *bolt.DB, path string) error {
	wallets, err := wallet.ListNames(db)
	if err != nil {
		return err
	}

	parts := strings.Split(path, "/")
	n := len(parts) - 1 // wallet name index
	name := parts[n]    // name without folders

	for _, c := range wallets {
		if strings.Contains(c.Name, path) {
			walletName := strings.Split(c.Name, "/")[n]

			if walletName == name {
				return errors.Errorf("already exists a wallet or folder named %q", path)
			}
		}
	}

	return nil
}
