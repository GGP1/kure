package rm

import (
	"fmt"
	"io"
	"strings"

	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/totp"
	"github.com/GGP1/kure/terminal"

	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

const example = `
* Remove a TOTP
kure 2fa rm Sample

* Remove a directory
kure 2fa rm SampleDir/`

// NewCmd returns the a new command.
func NewCmd(db *bolt.DB, r io.Reader) *cobra.Command {
	return &cobra.Command{
		Use:     "rm <name>",
		Short:   "Remove one or many two-factor authentication codes",
		Example: example,
		Args:    cmdutil.MustExist(db, cmdutil.TOTP, true),
		RunE:    runRm(db, r),
	}
}

func runRm(db *bolt.DB, r io.Reader) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")
		name = cmdutil.NormalizeName(name)

		if !terminal.Confirm(r, "Are you sure you want to proceed?") {
			return nil
		}

		if !strings.HasSuffix(name, "/") {
			if err := totp.Remove(db, name); err != nil {
				return err
			}

			fmt.Printf("\n%q TOTP removed\n", name)
			return nil
		}

		totps, err := totp.ListNames(db)
		if err != nil {
			return err
		}

		for _, t := range totps {
			if strings.HasPrefix(t, name) {
				if err := totp.Remove(db, t); err != nil {
					return err
				}

				fmt.Println("Remove:", t)
			}
		}

		return nil
	}
}
