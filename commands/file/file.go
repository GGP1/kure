package file

import (
	"os"

	fadd "github.com/GGP1/kure/commands/file/add"
	fcat "github.com/GGP1/kure/commands/file/cat"
	fedit "github.com/GGP1/kure/commands/file/edit"
	fls "github.com/GGP1/kure/commands/file/ls"
	fmv "github.com/GGP1/kure/commands/file/mv"
	frm "github.com/GGP1/kure/commands/file/rm"
	ftouch "github.com/GGP1/kure/commands/file/touch"

	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

const example = `
kure file (add|cat|edit|ls|mv|rm|touch)`

// NewCmd returns a new command.
func NewCmd(db *bolt.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "file",
		Short:   "File operations",
		Example: example,
	}

	cmd.AddCommand(
		fadd.NewCmd(db, os.Stdin),
		fcat.NewCmd(db, os.Stdout),
		fedit.NewCmd(db),
		fls.NewCmd(db),
		fmv.NewCmd(db),
		frm.NewCmd(db, os.Stdin),
		ftouch.NewCmd(db),
	)

	return cmd
}
