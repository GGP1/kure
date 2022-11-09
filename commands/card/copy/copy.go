package copy

import (
	"strings"
	"time"

	"github.com/GGP1/kure/auth"
	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/card"

	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

const example = `
* Copy the number
kure card copy Sample

* Copy the security code
kure card copy Sample -c

* Copy and clean after 30s
kure card copy Sample -t 30s`

type copyOptions struct {
	cvc     bool
	timeout time.Duration
}

// NewCmd returns a new command.
func NewCmd(db *bolt.DB) *cobra.Command {
	opts := copyOptions{}

	cmd := &cobra.Command{
		Use:     "copy <name>",
		Short:   "Copy card number or security code",
		Aliases: []string{"cp"},
		Example: example,
		Args:    cmdutil.MustExist(db, cmdutil.Card),
		PreRunE: auth.Login(db),
		RunE:    runCard(db, &opts),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset variables (session)
			opts = copyOptions{}
		},
	}

	f := cmd.Flags()
	f.BoolVarP(&opts.cvc, "cvc", "c", false, "copy card security code")
	f.DurationVarP(&opts.timeout, "timeout", "t", 0, "clipboard clearing timeout")

	return cmd
}

func runCard(db *bolt.DB, opts *copyOptions) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")
		name = cmdutil.NormalizeName(name)

		c, err := card.Get(db, name)
		if err != nil {
			return err
		}

		field := "Number"
		copy := c.Number
		if opts.cvc {
			field = "Security code"
			copy = c.SecurityCode
		}

		return cmdutil.WriteClipboard(cmd, opts.timeout, field, copy)
	}
}
