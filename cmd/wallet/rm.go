package wallet

import (
	"fmt"
	"io"
	"strings"

	cmdutil "github.com/GGP1/kure/cmd"
	"github.com/GGP1/kure/db/wallet"

	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var rmExample = `
* Remove a wallet
kure wallet rm walletName`

func rmSubCmd(db *bolt.DB, r io.Reader) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rm <name>",
		Short:   "Remove a wallet from the database",
		Example: rmExample,
		RunE:    runRm(db, r),
	}

	return cmd
}

func runRm(db *bolt.DB, r io.Reader) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")
		if name == "" {
			return errInvalidName
		}

		if cmdutil.Proceed(r) {
			if err := wallet.Remove(db, name); err != nil {
				return err
			}
		}

		fmt.Printf("\nSuccessfully removed %q wallet.\n", name)
		return nil
	}
}
