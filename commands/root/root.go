package root

import (
	"os"

	tfa "github.com/GGP1/kure/commands/2fa"
	"github.com/GGP1/kure/commands/add"
	"github.com/GGP1/kure/commands/backup"
	"github.com/GGP1/kure/commands/card"
	"github.com/GGP1/kure/commands/clear"
	"github.com/GGP1/kure/commands/config"
	"github.com/GGP1/kure/commands/copy"
	"github.com/GGP1/kure/commands/edit"
	"github.com/GGP1/kure/commands/export"
	"github.com/GGP1/kure/commands/file"
	"github.com/GGP1/kure/commands/gen"
	importt "github.com/GGP1/kure/commands/import"
	"github.com/GGP1/kure/commands/it"
	"github.com/GGP1/kure/commands/ls"
	"github.com/GGP1/kure/commands/restore"
	"github.com/GGP1/kure/commands/rm"
	"github.com/GGP1/kure/commands/session"
	"github.com/GGP1/kure/commands/stats"

	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var cmd = &cobra.Command{
	Use:           "kure",
	Short:         "Kure ~ CLI password manager",
	Version:       "0.3.0",
	SilenceErrors: true,
	SilenceUsage:  true,
}

// Execute adds all the subcommands to the root and executes it.
func Execute(db *bolt.DB) error {
	registerCmds(db)

	return cmd.Execute()
}

// registerCmds adds all the commands to the root.
func registerCmds(db *bolt.DB) {
	cmd.AddCommand(tfa.NewCmd(db))
	cmd.AddCommand(add.NewCmd(db, os.Stdin))
	cmd.AddCommand(backup.NewCmd(db))
	cmd.AddCommand(card.NewCmd(db))
	cmd.AddCommand(clear.NewCmd())
	cmd.AddCommand(config.NewCmd(db, os.Stdin))
	cmd.AddCommand(copy.NewCmd(db))
	cmd.AddCommand(edit.NewCmd(db))
	cmd.AddCommand(export.NewCmd(db))
	cmd.AddCommand(file.NewCmd(db))
	cmd.AddCommand(gen.NewCmd())
	cmd.AddCommand(importt.NewCmd(db))
	cmd.AddCommand(it.NewCmd(db))
	cmd.AddCommand(ls.NewCmd(db))
	cmd.AddCommand(restore.NewCmd(db))
	cmd.AddCommand(rm.NewCmd(db, os.Stdin))
	cmd.AddCommand(session.NewCmd(db, os.Stdin))
	cmd.AddCommand(stats.NewCmd(db))
}

// CmdForDocs returns the root command and should be used for documentation purposes only.
func CmdForDocs() *cobra.Command {
	registerCmds(nil)
	return cmd
}
