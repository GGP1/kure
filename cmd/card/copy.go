package card

import (
	"fmt"
	"strings"
	"time"

	cmdutil "github.com/GGP1/kure/cmd"
	"github.com/GGP1/kure/db/card"
	"github.com/awnumar/memguard"

	"github.com/atotto/clipboard"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var (
	cvc     bool
	timeout time.Duration
)

var copyExample = `
* Copy the number
kure card copy

* Copy the security code
kure card copy -c

* Copy and clean after 30s
kure card copy -t 30s`

// copySubCmd returns the copy subcommand
func copySubCmd(db *bolt.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "copy <name>",
		Short:   "Copy card number or security code",
		Aliases: []string{"cp"},
		Example: copyExample,
		PreRunE: cmdutil.RequirePassword(db),
		RunE:    runCard(db),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset flags (session)
			cvc = false
			timeout = 0
		},
	}

	f := cmd.Flags()
	f.BoolVarP(&cvc, "cvc", "c", false, "copy card security code")
	f.DurationVarP(&timeout, "timeout", "t", 0, "clipboard cleaning timeout")

	return cmd
}

func runCard(db *bolt.DB) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")
		if name == "" {
			return errInvalidName
		}

		cardBuf, card, err := card.Get(db, name)
		if err != nil {
			return err
		}

		field := "Number"
		copy := card.Number
		if cvc {
			field = "Security code"
			copy = card.SecurityCode
		}
		cardBuf.Destroy()

		if err := clipboard.WriteAll(copy); err != nil {
			return errors.Wrap(err, "failed writing to the clipboard")
		}
		memguard.WipeBytes([]byte(copy))

		fmt.Printf("%s copied to clipboard\n", field)

		if timeout > 0 {
			<-time.After(timeout)
			clipboard.WriteAll("")
		}

		return nil
	}
}
