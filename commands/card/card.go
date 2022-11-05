package card

import (
	"os"

	cadd "github.com/GGP1/kure/commands/card/add"
	ccopy "github.com/GGP1/kure/commands/card/copy"
	cedit "github.com/GGP1/kure/commands/card/edit"
	cls "github.com/GGP1/kure/commands/card/ls"
	crm "github.com/GGP1/kure/commands/card/rm"

	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

const example = `
kure card (add|copy|edit|ls|rm)`

// NewCmd returns a new command.
func NewCmd(db *bolt.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "card",
		Short:   "Card operations",
		Example: example,
	}

	cmd.AddCommand(
		cadd.NewCmd(db, os.Stdin),
		ccopy.NewCmd(db),
		cedit.NewCmd(db),
		cls.NewCmd(db),
		crm.NewCmd(db, os.Stdin),
	)

	return cmd
}
