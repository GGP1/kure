package wallet

import (
	"strings"
	"time"

	cmdutil "github.com/GGP1/kure/cmd"
	"github.com/GGP1/kure/db/wallet"

	"github.com/atotto/clipboard"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var timeout time.Duration

var errWritingClipboard = errors.New("failed writing to the clipboard")

var copyExample = `
* Copy
kure wallet copy walletName

* Copy with a timeout of 1h
kure wallet copy walletName -t 1h`

func copySubCmd(db *bolt.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "copy <name> [-t timeout]",
		Short:   "Copy wallet public key",
		Aliases: []string{"c"},
		Example: copyExample,
		RunE:    runCopy(db),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset flags defaults (session)
			timeout = 0
		},
	}

	cmd.Flags().DurationVarP(&timeout, "timeout", "t", 0, "clipboard cleaning timeout")

	return cmd
}

func runCopy(db *bolt.DB) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")
		if name == "" {
			return errInvalidName
		}

		wallet, err := wallet.Get(db, name)
		if err != nil {
			return err
		}

		if err := clipboard.WriteAll(wallet.PublicKey); err != nil {
			return errWritingClipboard
		}

		if timeout > 0 {
			<-time.After(timeout)
			clipboard.WriteAll("")
		}

		return nil
	}
}
