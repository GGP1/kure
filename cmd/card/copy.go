package card

import (
	"strings"
	"time"

	cmdutil "github.com/GGP1/kure/cmd"
	"github.com/GGP1/kure/db/card"

	"github.com/atotto/clipboard"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var (
	field   string
	timeout time.Duration
)

var errWritingClipboard = errors.New("failed writing to the clipboard")

var copyExample = `
* Copy the number
kure copy -f number

* Copy the CVC
kure copy -f cvc

* Copy and clean after 30s
kure copy -t 30s`

// copySubCmd returns the copy subcommand
func copySubCmd(db *bolt.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "copy <name> [-t timeout] [-f field]",
		Short:   "Copy card number or CVC",
		Aliases: []string{"c"},
		Example: copyExample,
		RunE:    runCopy(db),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset flags defaults (session)
			field = "number"
			timeout = 0
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	f := cmd.Flags()
	f.StringVarP(&field, "field", "f", "number", "choose which field to copy")
	f.DurationVarP(&timeout, "timeout", "t", 0, "clipboard cleaning timeout")

	return cmd
}

func runCopy(db *bolt.DB) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")
		if name == "" {
			return errInvalidName
		}

		card, err := card.Get(db, name)
		if err != nil {
			return errors.Wrap(err, "error")
		}

		var f string // change to locked buffer
		field = strings.ToLower(field)

		switch field {
		case "number":
			f = card.Number
		case "code", "cvc":
			f = card.CVC
		default:
			return errors.New("error: invalid card field, use \"number\" or \"code\"/\"cvc\"")
		}

		if err := clipboard.WriteAll(f); err != nil {
			return errors.Wrap(errWritingClipboard, "error")
		}

		if timeout > 0 {
			<-time.After(timeout)
			clipboard.WriteAll("")
		}

		return nil
	}
}
