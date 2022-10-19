package rm

import (
	"fmt"
	"io"
	"strings"

	"github.com/GGP1/kure/auth"
	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/card"
	"github.com/GGP1/kure/terminal"

	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

const example = `
* Remove a card
kure card rm Sample

* Remove a directory
kure card rm SampleDir/`

// NewCmd returns the a new command.
func NewCmd(db *bolt.DB, r io.Reader) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rm <name>",
		Short:   "Remove a card or directory",
		Example: example,
		Args:    cmdutil.MustExist(db, cmdutil.Card, true),
		PreRunE: auth.Login(db),
		RunE:    runRm(db, r),
	}

	return cmd
}

func runRm(db *bolt.DB, r io.Reader) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")
		name = cmdutil.NormalizeName(name, true)

		if !terminal.Confirm(r, "Are you sure you want to proceed?") {
			return nil
		}

		// Remove single file
		if !strings.HasSuffix(name, "/") {
			if err := card.Remove(db, name); err != nil {
				return err
			}

			fmt.Printf("\n%q removed\n", name)
			return nil
		}

		fmt.Printf("Removing %q directory...\n", name)

		cards, err := card.ListNames(db)
		if err != nil {
			return err
		}

		for _, c := range cards {
			if strings.HasPrefix(c, name) {
				if err := card.Remove(db, c); err != nil {
					return err
				}
				fmt.Println("Remove:", c)
			}
		}

		return nil
	}
}
