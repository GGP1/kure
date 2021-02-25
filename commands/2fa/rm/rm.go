package rm

import (
	"fmt"
	"io"
	"strings"

	"github.com/GGP1/kure/auth"
	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/totp"

	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var example = `
kure 2fa rm Sample`

// NewCmd returns the a new command.
func NewCmd(db *bolt.DB, r io.Reader) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rm <name>",
		Short:   "Remove a two-factor authentication code from an entry",
		Example: example,
		Args:    cmdutil.MustExist(db, cmdutil.TOTP),
		PreRunE: auth.Login(db),
		RunE:    runRm(db, r),
	}

	return cmd
}

func runRm(db *bolt.DB, r io.Reader) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")
		name = cmdutil.NormalizeName(name)

		if !cmdutil.Confirm(r, "Are you sure you want to proceed?") {
			return nil
		}

		if err := totp.Remove(db, name); err != nil {
			return err
		}

		fmt.Printf("\n%q TOTP removed\n", name)
		return nil
	}
}
